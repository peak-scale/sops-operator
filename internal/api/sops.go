// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package api

// Metadata is stored in SOPS encrypted files, and it contains the information necessary to decrypt the file.
// This struct is just used for serialization, and SOPS uses another struct internally, sops.Metadata. It exists
// in order to allow the binary format to stay backwards compatible over time, but at the same time allow the internal
// representation SOPS uses to change over time.
// +kubebuilder:object:generate=true
type Metadata struct {
	ShamirThreshold           int         `json:"shamir_threshold,omitempty"`
	KeyGroups                 []Keygroup  `json:"key_groups,omitempty"`
	Kmskeys                   []Kmskey    `json:"kms,omitempty"`
	GcpKmskeys                []GcpKmskey `json:"gcp_kms,omitempty"`
	AzureKeyVaultkeys         []Azkvkey   `json:"azure_kv,omitempty"`
	Vaultkeys                 []Vaultkey  `json:"hc_vault,omitempty"`
	Agekeys                   []Agekey    `json:"age,omitempty"`
	LastModified              string      `json:"lastmodified"`
	MessageAuthenticationCode string      `json:"mac"`
	Pgpkeys                   []Pgpkey    `json:"pgp,omitempty"`
	UnencryptedSuffix         string      `json:"unencrypted_suffix,omitempty"`
	EncryptedSuffix           string      `json:"encrypted_suffix,omitempty"`
	UnencryptedRegex          string      `json:"unencrypted_regex,omitempty"`
	EncryptedRegex            string      `json:"encrypted_regex,omitempty"`
	UnencryptedCommentRegex   string      `json:"unencrypted_comment_regex,omitempty"`
	EncryptedCommentRegex     string      `json:"encrypted_comment_regex,omitempty"`
	MACOnlyEncrypted          bool        `json:"mac_only_encrypted,omitempty"`
	Version                   string      `json:"version,omitempty"`
}

// +kubebuilder:object:generate=true
type Keygroup struct {
	Pgpkeys           []Pgpkey    `json:"pgp,omitempty"`
	Kmskeys           []Kmskey    `json:"kms,omitempty"`
	GcpKmskeys        []GcpKmskey `json:"gcp_kms,omitempty"`
	AzureKeyVaultkeys []Azkvkey   `json:"azure_kv,omitempty"`
	Vaultkeys         []Vaultkey  `json:"hc_vault,omitempty"`
	Agekeys           []Agekey    `json:"age,omitempty"`
}

// +kubebuilder:object:generate=true
type Pgpkey struct {
	CreatedAt        string `json:"created_at,omitempty"`
	EncryptedDataKey string `json:"enc,omitempty"`
	Fingerprint      string `json:"fp,omitempty"`
}

// +kubebuilder:object:generate=true
type Kmskey struct {
	Arn              string             `json:"arn"`
	Role             string             `json:"role,omitempty"`
	Context          map[string]*string `json:"context,omitempty"`
	CreatedAt        string             `json:"created_at"`
	EncryptedDataKey string             `json:"enc"`
	AwsProfile       string             `json:"aws_profile"`
}

// +kubebuilder:object:generate=true
type GcpKmskey struct {
	ResourceID       string `json:"resource_id"`
	CreatedAt        string `json:"created_at"`
	EncryptedDataKey string `json:"enc"`
}

// +kubebuilder:object:generate=true
type Vaultkey struct {
	VaultAddress     string `json:"vault_address"`
	EnginePath       string `json:"engine_path"`
	KeyName          string `json:"key_name"`
	CreatedAt        string `json:"created_at"`
	EncryptedDataKey string `json:"enc"`
}

// +kubebuilder:object:generate=true
type Azkvkey struct {
	VaultURL         string `json:"vault_url"`
	Name             string `json:"name"`
	Version          string `json:"version"`
	CreatedAt        string `json:"created_at"`
	EncryptedDataKey string `json:"enc"`
}

// +kubebuilder:object:generate=true
type Agekey struct {
	Recipient        string `json:"recipient"`
	EncryptedDataKey string `json:"enc"`
}
