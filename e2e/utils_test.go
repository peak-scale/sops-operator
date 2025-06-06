//nolint:all
package e2e_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"time"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/meta"
)

const (
	defaultTimeoutInterval = 40 * time.Second
	defaultPollInterval    = time.Second
)

// All Secrets are expected to be decrypted
func ValidateSopsSecret(filePath string, name string, namespace string) error {
	expectedSecret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret](filePath)
	if err != nil {
		return err
	}
	if expectedSecret == nil {
		return fmt.Errorf("no SopsSecret was loaded from file: %s", filePath)
	}

	secretCount := len(expectedSecret.Spec.Secrets)
	if secretCount == 0 {
		return fmt.Errorf("expected at least one secret in .spec.secrets, but got 0 from file: %s", filePath)
	}

	fmt.Printf("✅ Loaded %d secrets from file: %s\n", secretCount, filePath)

	sops := &sopsv1alpha1.SopsSecret{}
	if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, sops); err != nil {
		return fmt.Errorf("failed to get sopssecret from cluster: %w", err)
	}

	for _, specSecret := range expectedSecret.Spec.Secrets {
		key := types.NamespacedName{Name: specSecret.Name, Namespace: namespace}
		secret := &corev1.Secret{}
		if err := k8sClient.Get(context.TODO(), key, secret); err != nil {
			return fmt.Errorf("failed to get secret %s/%s: %w", key.Namespace, key.Name, err)
		}

		// Compare stringData (declared) to data (base64) in Kubernetes
		for k, v := range specSecret.StringData {
			encoded := base64.StdEncoding.EncodeToString([]byte(v))
			actual, ok := secret.Data[k]
			if !ok {
				return fmt.Errorf("secret %s/%s missing key %q", key.Namespace, key.Name, k)
			}

			if encoded != base64.StdEncoding.EncodeToString(actual) {
				return fmt.Errorf("secret %s/%s key %q does not match expected content", key.Namespace, key.Name, k)
			}
		}

		found := false
		for _, statusSecret := range sops.Status.Secrets {
			if statusSecret.UID != secret.UID {
				continue
			}
			found = true

			Expect(statusSecret.Name).To(Equal(secret.Name))
			Expect(statusSecret.Namespace).To(Equal(secret.Namespace))

			break
		}

		Expect(found).To(BeTrue())
	}

	Expect(sops.Status.Size).To(Equal(uint(len(expectedSecret.Spec.Secrets))))

	return nil
}

func ValidateSopsSecretAbsence(filePath string, name string, namespace string) error {
	expectedSecret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret](filePath)
	if err != nil {
		return fmt.Errorf("failed to load SopsSecret from file: %w", err)
	}
	if expectedSecret == nil {
		return fmt.Errorf("no SopsSecret loaded from file: %s", filePath)
	}

	secretCount := len(expectedSecret.Spec.Secrets)
	if secretCount == 0 {
		return fmt.Errorf("expected at least one secret in .spec.secrets, but got 0 from file: %s", filePath)
	}

	// Attempt to fetch the SopsSecret from the cluster
	sops := &sopsv1alpha1.SopsSecret{}
	err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, sops)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("unexpected error getting SopsSecret: %w", err)
	}

	for _, specSecret := range expectedSecret.Spec.Secrets {
		key := types.NamespacedName{Name: specSecret.Name, Namespace: namespace}
		secret := &corev1.Secret{}
		err := k8sClient.Get(context.TODO(), key, secret)
		if err == nil {
			return fmt.Errorf("secret %s/%s unexpectedly still exists in the cluster", key.Namespace, key.Name)
		}
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking secret %s/%s: %w", key.Namespace, key.Name, err)
		}

		// If the SopsSecret resource was found, check if it still references the secret
		if err == nil {
			for _, statusSecret := range sops.Status.Secrets {
				if statusSecret.Name == specSecret.Name && statusSecret.Namespace == namespace {
					return fmt.Errorf("secret %s/%s is still referenced in SopsSecret.status", namespace, specSecret.Name)
				}
			}
		}
	}

	fmt.Printf("✅ Verified: All secrets from %s are absent from the cluster and SopsSecret.status\n", filePath)
	return nil
}

func verifySecretToProviderAssociation(
	provider *sopsv1alpha1.SopsProvider,
	secret *sopsv1alpha1.SopsSecret,
) bool {
	fetched := &sopsv1alpha1.SopsSecret{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, fetched)
	Expect(err).Should(Succeed())

	for _, prov := range fetched.Status.Providers {
		if prov.Name == provider.Name && prov.UID == provider.GetUID() {
			return true
		}
	}

	return false
}

func verifyKeyAssociation(
	provider *sopsv1alpha1.SopsProvider,
	key *corev1.Secret,
) bool {
	fetched := &sopsv1alpha1.SopsProvider{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: provider.Name}, fetched)
	Expect(err).Should(Succeed())

	for _, sec := range fetched.Status.Providers {
		if sec.Origin != *api.NewOrigin(key) {
			continue
		}

		return true
	}

	return false
}

func verifyKeyAssociationSuccess(
	provider *sopsv1alpha1.SopsProvider,
	key *corev1.Secret,
) bool {
	fetched := &sopsv1alpha1.SopsProvider{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: provider.Name}, fetched)
	Expect(err).Should(Succeed())

	for _, sec := range fetched.Status.Providers {
		if sec.Origin != *api.NewOrigin(key) {
			continue
		}

		if sec.Condition.Reason != meta.SucceededReason {
			continue
		}

		if sec.Condition.Status != metav1.ConditionTrue {
			continue
		}

		if sec.Condition.Type != meta.ReadyCondition {
			continue
		}

		return true
	}

	return false
}

func verifyKeyAssociationFailure(
	provider *sopsv1alpha1.SopsProvider,
	key *corev1.Secret,
) bool {
	fetched := &sopsv1alpha1.SopsProvider{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: provider.Name}, fetched)
	Expect(err).Should(Succeed())

	for _, sec := range fetched.Status.Providers {
		if sec.Origin != *api.NewOrigin(key) {
			continue
		}

		if sec.Condition.Reason != meta.FailedReason {
			continue
		}

		if sec.Condition.Status != metav1.ConditionFalse {
			continue
		}

		if sec.Condition.Type != meta.NotReadyCondition {
			continue
		}

		return true
	}

	return false
}

func LoadFromYAMLFile[T client.Object](path string) (T, error) {
	var zero T

	data, err := os.ReadFile(path)
	if err != nil {
		return zero, fmt.Errorf("failed to read YAML file: %w", err)
	}

	obj := new(T)

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	if err := decoder.Decode(obj); err != nil {
		return zero, fmt.Errorf("failed to decode YAML into object: %w", err)
	}

	return *obj, nil
}

func EventuallyCreation(f interface{}) AsyncAssertion {
	return Eventually(f, defaultTimeoutInterval, defaultPollInterval)
}

func DeepCompare(expected, actual interface{}) (bool, string) {
	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	// If the kinds differ, they are not equal.
	if expVal.Kind() != actVal.Kind() {
		return false, fmt.Sprintf("kind mismatch: %v vs %v", expVal.Kind(), actVal.Kind())
	}

	switch expVal.Kind() {
	case reflect.Slice, reflect.Array:
		// Convert slices to []interface{} for ElementsMatch.
		expSlice := make([]interface{}, expVal.Len())
		actSlice := make([]interface{}, actVal.Len())
		for i := 0; i < expVal.Len(); i++ {
			expSlice[i] = expVal.Index(i).Interface()
		}
		for i := 0; i < actVal.Len(); i++ {
			actSlice[i] = actVal.Index(i).Interface()
		}
		// Use a dummy tester to capture error messages.
		dummy := &dummyT{}
		if !assert.ElementsMatch(dummy, expSlice, actSlice) {
			return false, fmt.Sprintf("slice mismatch: %v", dummy.errors)
		}
		return true, ""
	case reflect.Struct:
		// Iterate over fields and compare recursively.
		for i := 0; i < expVal.NumField(); i++ {
			fieldName := expVal.Type().Field(i).Name
			ok, msg := DeepCompare(expVal.Field(i).Interface(), actVal.Field(i).Interface())
			if !ok {
				return false, fmt.Sprintf("field %s mismatch: %s", fieldName, msg)
			}
		}
		return true, ""
	default:
		// Fallback to reflect.DeepEqual for other types.
		if !reflect.DeepEqual(expected, actual) {
			return false, fmt.Sprintf("expected %v but got %v", expected, actual)
		}
		return true, ""
	}
}

// dummyT implements a minimal TestingT for testify.
type dummyT struct {
	errors []string
}

func (d *dummyT) Errorf(format string, args ...interface{}) {
	d.errors = append(d.errors, fmt.Sprintf(format, args...))
}
