/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package decryptor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/config"
	"github.com/getsops/sops/v3/keyservice"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/decryptor/kustomize-controller/age"
	"github.com/peak-scale/sops-operator/internal/decryptor/kustomize-controller/awskms"
	"github.com/peak-scale/sops-operator/internal/decryptor/kustomize-controller/azkv"
	intkeyservice "github.com/peak-scale/sops-operator/internal/decryptor/kustomize-controller/keyservice"
	"github.com/peak-scale/sops-operator/internal/decryptor/kustomize-controller/pgp"
)

const (
	// DecryptionProviderSOPS is the SOPS provider name.
	DecryptionProviderSOPS = "sops"
	// DecryptionPGPExt is the extension of the file containing an armored PGP
	// key.
	DecryptionPGPExt = ".asc"
	// DecryptionAgeExt is the extension of the file containing an age key
	// file.
	DecryptionAgeExt = ".agekey"
	// DecryptionVaultTokenFileName is the name of the file containing the
	// Hashicorp Vault token.
	DecryptionVaultTokenFileName = "sops.vault-token"
	// DecryptionAWSKmsFile is the name of the file containing the AWS KMS
	// credentials.
	DecryptionAWSKmsFile = "sops.aws-kms"
	// DecryptionAzureAuthFile is the name of the file containing the Azure
	// credentials.
	DecryptionAzureAuthFile = "sops.azure-kv"
	// DecryptionGCPCredsFile is the name of the file containing the GCP
	// credentials.
	DecryptionGCPCredsFile = "sops.gcp-kms"
	// maxEncryptedFileSize is the max allowed file size in bytes of an encrypted
	// file.
	maxEncryptedFileSize int64 = 5 << 20
	// unsupportedFormat is used to signal no sopsFormatToMarkerBytes format was
	// detected by detectFormatFromMarkerBytes.
	unsupportedFormat = formats.Format(-1)
)

var (
	// sopsFormatToString is the counterpart to
	// https://github.com/mozilla/sops/blob/v3.7.2/cmd/sops/formats/formats.go#L16
	sopsFormatToString = map[formats.Format]string{
		formats.Binary: "binary",
		formats.Dotenv: "dotenv",
		formats.Ini:    "INI",
		formats.Json:   "JSON",
		formats.Yaml:   "YAML",
	}
	// sopsFormatToMarkerBytes contains a list of formats and their byte
	// order markers, used to detect if a Secret data field is SOPS' encrypted.
	sopsFormatToMarkerBytes = map[formats.Format][]byte{
		// formats.Binary is a JSON envelop at encrypted rest
		formats.Binary: []byte("\"mac\": \"ENC["),
		formats.Dotenv: []byte("sops_mac=ENC["),
		formats.Ini:    []byte("[sops]"),
		formats.Json:   []byte("\"mac\": \"ENC["),
		formats.Yaml:   []byte("mac: ENC["),
	}
)

type secretItemSubset struct {
	Data       map[string]string `json:"data,omitempty"`
	StringData map[string]string `json:"stringData,omitempty"`
	Sops       *api.SopsMetadata `json:"sops,omitempty"`
}

// Decryptor performs decryption operations for a v1.Kustomization.
// The only supported decryption provider at present is
// DecryptionProviderSOPS.
type SOPSDecryptor struct {
	// maxFileSize is the max size in bytes a file is allowed to have to be
	// decrypted. Defaults to maxEncryptedFileSize.
	maxFileSize int64
	// checkSopsMac instructs the decryptor to perform the SOPS data integrity
	// check using the MAC. Not enabled by default, as arbitrary data gets
	// injected into most resources, causing the integrity check to fail.
	// Mostly kept around for feature completeness and documentation purposes.
	checkSopsMac bool

	// gnuPGHome is the absolute path of the GnuPG home directory used to
	// decrypt PGP data. When empty, the systems' GnuPG keyring is used.
	// When set, ImportKeys() imports found PGP keys into this keyring.
	gnuPGHome pgp.GnuPGHome
	// ageIdentities is the set of age identities available to the decryptor.
	ageIdentities age.ParsedIdentities
	// vaultToken is the Hashicorp Vault token used to authenticate towards
	// any Vault server.
	vaultToken string
	// awsCredsProvider is the AWS credentials provider object used to authenticate
	// towards any AWS KMS.
	awsCredsProvider *awskms.CredsProvider
	// azureToken is the Azure credential token used to authenticate towards
	// any Azure Key Vault.
	azureToken *azkv.Token
	// gcpCredsJSON is the JSON credential file of the service account used to
	// authenticate towards any GCP KMS.
	gcpCredsJSON []byte

	// keyServices are the SOPS keyservice.KeyServiceClient's available to the
	// decryptor.
	keyServices      []keyservice.KeyServiceClient
	localServiceOnce sync.Once
}

// NewDecryptor creates a new Decryptor for the given kustomization.
// gnuPGHome can be empty, in which case the systems' keyring is used.
func NewSOPSDecryptor(gnuPGHome string) *SOPSDecryptor {
	return &SOPSDecryptor{
		maxFileSize: maxEncryptedFileSize,
		gnuPGHome:   pgp.GnuPGHome(gnuPGHome),
	}
}

// NewTempDecryptor creates a new Decryptor, with a temporary GnuPG
// home directory to Decryptor.ImportKeys() into.
func NewSOPSTempDecryptor() (*SOPSDecryptor, func(), error) {
	gnuPGHome, err := pgp.NewGnuPGHome()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create keyring: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(gnuPGHome.String()) }
	return NewSOPSDecryptor(gnuPGHome.String()), cleanup, nil
}

// Only call this for Temporary Decryptors
func (d *SOPSDecryptor) RemoveKeyRing() error {
	return os.RemoveAll(string(d.gnuPGHome))
}

// IsEncrypted returns true if the given data is encrypted by SOPS.
func (d *SOPSDecryptor) IsEncrypted(data *sopsv1alpha1.SopsSecret) (bool, error) {
	sopsField := data.Sops
	if sopsField == nil {
		return false, nil
	}
	return true, nil
}

// Read reads the input data, decrypts it, and returns the decrypted data.
func (d *SOPSDecryptor) Decrypt(data *sopsv1alpha1.SopsSecret, secret *sopsv1alpha1.SopsSecretItem, log logr.Logger) error {
	// Loop over each secret item in the Spec.
	// We need to restore the origin reference
	entry := &sopsv1alpha1.SopsSecret{
		Spec: sopsv1alpha1.SopsSecretSpec{
			Secrets: []*sopsv1alpha1.SopsSecretItem{
				secret,
			},
		},
		Sops: data.Sops,
	}

	b, _ := json.Marshal(entry)

	inFormat := formats.Json
	outFormat := formats.Json

	// Decrypt using SopsDecryptWithFormat.
	decryptedBytes, err := d.SopsDecryptWithFormat(b, log, inFormat, outFormat)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret field: %w", err)
	}

	var target sopsv1alpha1.SopsSecret
	if err := json.Unmarshal(decryptedBytes, &target); err != nil {
		return err
	}
	// Rewrite Values
	secret.Data = target.Spec.Secrets[0].Data
	secret.StringData = target.Spec.Secrets[0].StringData

	return nil
}

//func (d *SOPSDecryptor) decryptSecretField(log logr.Logger, dataMap map[string]string, marker *api.SopsMetadata) error {
//	// Unmarshal marker bytes into a map to get the SOPS metadata.
//
//	minFile := sopsData{
//		data: dataMap,
//		Sops: marker,
//	}
//
//	wrappedBytes, err := yaml.Marshal(minFile)
//
//	log.V(1).Info(string(wrappedBytes))
//
//	// Here, we assume the input is YAML and we want YAML output.
//	inFormat := formats.Yaml
//	outFormat := formats.Yaml
//
//	// Decrypt using SopsDecryptWithFormat.
//	decryptedBytes, err := d.SopsDecryptWithFormat(wrappedBytes, log, inFormat, outFormat)
//	if err != nil {
//		return fmt.Errorf("failed to decrypt secret field: %w", err)
//	}
//
//	log.V(1).Info("DECYRPTED", decryptedBytes, "hi")
//
//	return nil
//}

//func (d *SOPSDecryptor) decryptSecretField(log logr.Logger, dataMap map[string]string, marker []byte) error {
//	for key, encodedValue := range dataMap {
//		decoded := []byte(encodedValue)
//
//		inFormat := detectFormatFromMarkerBytes(marker)
//		if inFormat == unsupportedFormat {
//			log.V(5).Info("unsupported format", "format", inFormat)
//			continue
//		}
//
//		// Determine the output format; (Differs for JSON content)
//		outFormat := formatForPath(key)
//
//		log.V(5).Info("formats", "in", inFormat, "out", outFormat)
//		decryptedBytes, err := d.SopsDecryptWithFormat(decoded, log, inFormat, outFormat)
//		if err != nil {
//			return fmt.Errorf("failed to decrypt secret field %q: %w", key, err)
//		}
//
//		dataMap[key] = base64.StdEncoding.EncodeToString(decryptedBytes)
//	}
//
//	log.V(10).Info("decrypted fields", "data", dataMap)
//
//	return nil
//}

// AddGPGKey adds given GPG key to the decryptor's keyring.
func (d *SOPSDecryptor) AddGPGKey(key []byte) error {
	return d.gnuPGHome.Import(key)
}

// AddAgeKey to the decryptor's identities.
func (d *SOPSDecryptor) AddAgeKey(key []byte) error {
	return d.ageIdentities.Import(string(key))
}

// SetVaultToken sets the Vault token for the decryptor.
func (d *SOPSDecryptor) SetVaultToken(token []byte) {
	vtoken := string(token)
	vtoken = strings.Trim(strings.TrimSpace(vtoken), "\n")
	d.vaultToken = vtoken
}

// SetAWSCredentials adds AWS credentials for the decryptor.
// Reference: https://github.com/getsops/sops#aws-kms-encryption-context
func (d *SOPSDecryptor) SetAWSCredentials(token []byte) (err error) {
	d.awsCredsProvider, err = awskms.LoadCredsProviderFromYaml(token)
	return err
}

// SetAzureAuthFile adds AWS credentials for the decryptor.
func (d *SOPSDecryptor) SetAzureCredentials(config []byte) (err error) {
	conf := azkv.AADConfig{}
	if err = azkv.LoadAADConfigFromBytes(config, &conf); err != nil {
		return err
	}
	if d.azureToken, err = azkv.TokenFromAADConfig(conf); err != nil {
		return err
	}

	return nil
}

// SetGCPCredentials adds GCP credentials for the decryptor.
func (d *SOPSDecryptor) SetGCPCredentials(config []byte) {
	d.gcpCredsJSON = bytes.Trim(config, "\n")
}

func (d *SOPSDecryptor) KeysFromSecret(ctx context.Context, c client.Client, secretName string, namespace string) (err error) {
	// Retrieve Secret
	var keySecret corev1.Secret
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, &keySecret); err != nil {
		if apierrors.IsNotFound(err) {
			return &MissingKubernetesSecret{Secret: secretName, Namespace: namespace}
		}
		return err
	}

	// Exract all keys from secret
	for name, value := range keySecret.Data {
		switch filepath.Ext(name) {
		case DecryptionPGPExt:
			if err = d.AddGPGKey(value); err != nil {
				return fmt.Errorf("failed to import data from %s decryption Secret '%s': %w", name, secretName, err)
			}
		case DecryptionAgeExt:
			if err = d.AddAgeKey(value); err != nil {
				return fmt.Errorf("failed to import data from %s decryption Secret '%s': %w", name, secretName, err)
			}
		case filepath.Ext(DecryptionVaultTokenFileName):
			// Make sure we have the absolute name
			if name == DecryptionVaultTokenFileName {
				d.SetVaultToken(value)
			}
		case filepath.Ext(DecryptionAWSKmsFile):
			if name == DecryptionAWSKmsFile {
				if d.SetAWSCredentials(value); err != nil {
					return fmt.Errorf("failed to import data from %s decryption Secret '%s': %w", name, secretName, err)
				}
			}
		case filepath.Ext(DecryptionAzureAuthFile):
			if name == DecryptionAzureAuthFile {
				if err = d.SetAzureCredentials(value); err != nil {
					return fmt.Errorf("failed to import data from %s decryption Secret '%s': %w", name, secretName, err)
				}
			}
		case filepath.Ext(DecryptionGCPCredsFile):
			if name == DecryptionGCPCredsFile {
				d.SetGCPCredentials(value)
			}
		}
	}

	return nil
}

// SopsDecryptWithFormat attempts to load a SOPS encrypted file using the store
// for the input format, gathers the data key for it from the key service,
// and then decrypts the file data with the retrieved data key.
// It returns the decrypted bytes in the provided output format, or an error.
func (d *SOPSDecryptor) SopsDecryptWithFormat(data []byte, log logr.Logger, inputFormat, outputFormat formats.Format) (_ []byte, err error) {
	defer func() {
		// It was discovered that malicious input and/or output instructions can
		// make SOPS panic. Recover from this panic and return as an error.
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to emit encrypted %s file as decrypted %s: %v",
				sopsFormatToString[inputFormat], sopsFormatToString[outputFormat], r)
		}
	}()

	store := common.StoreForFormat(inputFormat, config.NewStoresConfig())

	tree, err := store.LoadEncryptedFile(data)
	if err != nil {
		return nil, sopsUserErr(fmt.Sprintf("failed to load encrypted %s data", sopsFormatToString[inputFormat]), err)
	}

	if tree.Branches == nil {
		return nil, fmt.Errorf("tree.Branches is nil: invalid SOPS file structure")
	}

	keyService := d.keyServiceServer()
	if keyService == nil {
		return nil, fmt.Errorf("keyService is not initialized")
	}

	metadataKey, err := tree.Metadata.GetDataKeyWithKeyServices(d.keyServiceServer(), sops.DefaultDecryptionOrder)
	if err != nil {
		return nil, sopsUserErr("cannot get sops data key", err)
	}

	cipher := aes.NewCipher()
	mac, err := tree.Decrypt(metadataKey, cipher)
	if err != nil {
		return nil, sopsUserErr("error decrypting sops tree", err)
	}

	if d.checkSopsMac {
		// Compute the hash of the cleartext tree and compare it with
		// the one that was stored in the document. If they match,
		// integrity was preserved
		// Ref: github.com/getsops/sops/v3/decrypt/decrypt.go
		originalMac, err := cipher.Decrypt(
			tree.Metadata.MessageAuthenticationCode,
			metadataKey,
			tree.Metadata.LastModified.Format(time.RFC3339),
		)
		if err != nil {
			return nil, sopsUserErr("failed to verify sops data integrity", err)
		}
		if originalMac != mac {
			// If the file has an empty MAC, display "no MAC"
			if originalMac == "" {
				originalMac = "no MAC"
			}
			return nil, fmt.Errorf("failed to verify sops data integrity: expected mac '%s', got '%s'", originalMac, mac)
		}
	}

	outputStore := common.StoreForFormat(outputFormat, config.NewStoresConfig())
	out, err := outputStore.EmitPlainFile(tree.Branches)
	if err != nil {
		return nil, sopsUserErr(fmt.Sprintf("failed to emit encrypted %s file as decrypted %s",
			sopsFormatToString[inputFormat], sopsFormatToString[outputFormat]), err)
	}

	return out, err
}

// keyServiceServer returns the SOPS (local) key service clients used to serve
// decryption requests. loadKeyServiceServers() is only configured on the first
// call.
func (d *SOPSDecryptor) keyServiceServer() []keyservice.KeyServiceClient {
	d.localServiceOnce.Do(func() {
		d.loadKeyServiceServers()
	})
	return d.keyServices
}

// loadKeyServiceServers loads the SOPS (local) key service clients used to
// serve decryption requests for the current set of Decryptor
// credentials.
func (d *SOPSDecryptor) loadKeyServiceServers() {
	serverOpts := []intkeyservice.ServerOption{
		intkeyservice.WithGnuPGHome(d.gnuPGHome),
		intkeyservice.WithVaultToken(d.vaultToken),
		intkeyservice.WithAgeIdentities(d.ageIdentities),
		intkeyservice.WithGCPCredsJSON(d.gcpCredsJSON),
	}
	if d.azureToken != nil {
		serverOpts = append(serverOpts, intkeyservice.WithAzureToken{Token: d.azureToken})
	}
	serverOpts = append(serverOpts, intkeyservice.WithAWSKeys{CredsProvider: d.awsCredsProvider})
	server := intkeyservice.NewServer(serverOpts...)
	d.keyServices = append(make([]keyservice.KeyServiceClient, 0), keyservice.NewCustomLocalClient(server))
}

func sopsUserErr(msg string, err error) error {
	if userErr, ok := err.(sops.UserError); ok {
		err = fmt.Errorf(userErr.UserError())
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func formatForPath(path string) formats.Format {
	switch {
	case strings.HasSuffix(path, corev1.DockerConfigJsonKey):
		return formats.Json
	default:
		return formats.FormatForPath(path)
	}
}

func detectFormatFromMarkerBytes(b []byte) formats.Format {
	for k, v := range sopsFormatToMarkerBytes {
		if bytes.Contains(b, v) {
			return k
		}
	}
	return unsupportedFormat
}
