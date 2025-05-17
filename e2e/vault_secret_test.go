// Copyright 2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/meta"
)

var _ = Describe("Vault SOPS Tests", func() {
	suiteLabelValue := "e2e-vault"

	JustAfterEach(func() {
		Eventually(func() error {
			poolList := &sopsv1alpha1.SopsProviderList{}
			labelSelector := client.MatchingLabels{"e2e-test": suiteLabelValue}
			if err := k8sClient.List(context.TODO(), poolList, labelSelector); err != nil {
				return err
			}

			for _, pool := range poolList.Items {
				if err := k8sClient.Delete(context.TODO(), &pool); err != nil {
					return err
				}
			}

			return nil
		}, "30s", "5s").Should(Succeed())

		Eventually(func() error {
			poolList := &sopsv1alpha1.SopsSecretList{}
			labelSelector := client.MatchingLabels{"e2e-test": suiteLabelValue}
			if err := k8sClient.List(context.TODO(), poolList, labelSelector); err != nil {
				return err
			}

			for _, pool := range poolList.Items {
				if err := k8sClient.Delete(context.TODO(), &pool); err != nil {
					return err
				}
			}

			return nil
		}, "30s", "5s").Should(Succeed())

		Eventually(func() error {
			poolList := &corev1.NamespaceList{}
			labelSelector := client.MatchingLabels{"e2e-test": suiteLabelValue}
			if err := k8sClient.List(context.TODO(), poolList, labelSelector); err != nil {
				return err
			}

			for _, pool := range poolList.Items {
				if err := k8sClient.Delete(context.TODO(), &pool); err != nil {
					return err
				}
			}

			return nil
		}, "30s", "5s").Should(Succeed())

		Eventually(func() error {
			poolList := &corev1.SecretList{}
			labelSelector := client.MatchingLabels{"e2e-test": suiteLabelValue}
			if err := k8sClient.List(context.TODO(), poolList, labelSelector); err != nil {
				return err
			}

			for _, pool := range poolList.Items {
				if err := k8sClient.Delete(context.TODO(), &pool); err != nil {
					return err
				}
			}

			return nil
		}, "30s", "5s").Should(Succeed())

	})

	It("Vault Encryption Tests", func() {
		provider := &sopsv1alpha1.SopsProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name: "vault-provider-1",
				Labels: map[string]string{
					"e2e-test": suiteLabelValue,
				},
			},
			Spec: sopsv1alpha1.SopsProviderSpec{
				SOPSSelectors: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"secret-type": "vault",
							},
						},
					},
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"secret-type": "vault-1",
							},
						},
					},
				},

				ProviderSecrets: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"provider-vault": "1",
							},
						},
					},
				},
			},
		}

		By("Create the Provider", func() {
			err := k8sClient.Create(context.TODO(), provider)
			Expect(err).Should(Succeed(), "Failed to create provider %s", provider)
		})

		By("Create Namespaces, which where secrets can be sourced from", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-vault-provider-1",
					Labels: map[string]string{
						"e2e-test": suiteLabelValue,
						"customer": "vault-1",
					},
				},
			}

			err := k8sClient.Create(context.TODO(), ns)
			Expect(err).Should(Succeed())
		})

		By("Create Encrypted SOPS Secret (Key-2)", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/openbao/secret-key-2.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-vault-secret-2"
			secret.Namespace = "ns-vault-provider-1"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "vault-1",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			time.Sleep(10000 * time.Millisecond)

			Expect(verifySecretToProviderAssociation(provider, secret)).To(BeTrue())

			err = ValidateSopsSecretAbsence("testdata/openbao/secret-key-2.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Create Vault Token", func() {
			secret, err := LoadFromYAMLFile[*corev1.Secret]("testdata/openbao/token.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-vault-token-1"
			secret.Namespace = "ns-vault-provider-1"
			secret.Labels = map[string]string{
				meta.KeySecretLabel: "true",
				"e2e-test":          suiteLabelValue,
				"provider-vault":    "1",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
		})

		By("Verify Vault-Provider allocation", func() {
			secret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test-vault-token-1", Namespace: "ns-vault-provider-1"}, secret)
			Expect(err).ToNot(HaveOccurred())

			Expect(verifyKeyAssociation(provider, secret)).To(BeTrue())
		})

		By("Create Encrypted Vault Secret", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/openbao/secret-key-1.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-vault-secret-1"
			secret.Namespace = "ns-vault-provider-1"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "vault-1",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			time.Sleep(10000 * time.Millisecond)

			Expect(verifySecretToProviderAssociation(provider, secret)).To(BeTrue())

			err = ValidateSopsSecret("testdata/openbao/secret-key-1.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Reverify SOPS Secret (Key-2)", func() {
			secret := &sopsv1alpha1.SopsSecret{}
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test-vault-secret-2", Namespace: "ns-vault-provider-1"}, secret)
			Expect(err).Should(Succeed())

			time.Sleep(10000 * time.Millisecond)

			Expect(verifySecretToProviderAssociation(provider, secret)).To(BeTrue())

			err = ValidateSopsSecret("testdata/openbao/secret-key-2.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Create Multi-Secret (one of Key-1 or Key-2)", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/openbao/secret-multi.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-vault-multi-1"
			secret.Namespace = "ns-vault-provider-1"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "vault",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			time.Sleep(20000 * time.Millisecond)

			Expect(verifySecretToProviderAssociation(provider, secret)).To(BeTrue())

			err = ValidateSopsSecret("testdata/openbao/secret-multi.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Create Quorum-Secret (both of Key-1 or Key-2)", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/openbao/secret-quorum.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-vault-quorum-1"
			secret.Namespace = "ns-vault-provider-1"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "vault",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			time.Sleep(30000 * time.Millisecond)

			Expect(verifySecretToProviderAssociation(provider, secret)).To(BeTrue())

			err = ValidateSopsSecret("testdata/openbao/secret-quorum.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
