// Copyright 2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/meta"
)

var _ = Describe("Age SOPS Tests", Label("age"), func() {
	suiteLabelValue := "e2e-age"

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

	It("Age Encryption Tests", func() {
		provider1 := &sopsv1alpha1.SopsProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name: "age-provider-1",
				Labels: map[string]string{
					"e2e-test": suiteLabelValue,
				},
			},
			Spec: sopsv1alpha1.SopsProviderSpec{
				SOPSSelectors: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"secret-type": "age",
							},
						},
					},
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"secret-type": "age-1",
							},
						},
					},
				},

				ProviderSecrets: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"provider-age": "1",
							},
						},
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"customer": "age-1",
							},
						},
					},
				},
			},
		}

		provider2 := &sopsv1alpha1.SopsProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name: "age-provider-2",
				Labels: map[string]string{
					"e2e-test": suiteLabelValue,
				},
			},
			Spec: sopsv1alpha1.SopsProviderSpec{
				SOPSSelectors: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"secret-type": "age",
							},
						},
					},
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"secret-type": "age-2",
							},
						},
					},
				},

				ProviderSecrets: []*api.NamespacedSelector{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"customer": "age-2",
							},
						},
					},
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"customer": "age-1",
							},
						},
					},
				},
			},
		}

		By("Create the Provider", func() {
			err := k8sClient.Create(context.TODO(), provider1)
			Expect(err).Should(Succeed(), "Failed to create provider %s", provider1)

			err = k8sClient.Create(context.TODO(), provider2)
			Expect(err).Should(Succeed(), "Failed to create provider %s", provider2)
		})

		By("Get Applied revision", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: provider1.Name}, provider1)
			Expect(err).Should(Succeed())

			err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: provider2.Name}, provider2)
			Expect(err).Should(Succeed())
		})

		By("Create Namespaces, which where secrets can be sourced from", func() {
			ns1 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-age-provider-1",
					Labels: map[string]string{
						"e2e-test": suiteLabelValue,
						"customer": "age-1",
					},
				},
			}

			err := k8sClient.Create(context.TODO(), ns1)
			Expect(err).Should(Succeed())

			ns2 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-age-provider-2",
					Labels: map[string]string{
						"e2e-test": suiteLabelValue,
						"customer": "age-2",
					},
				},
			}

			err = k8sClient.Create(context.TODO(), ns2)
			Expect(err).Should(Succeed())
		})

		By("Create Encrypted SOPS Secret (Key-2)", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/age/secret-key-2.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-age-secret-2"
			secret.Namespace = "ns-age-provider-2"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "age-2",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			Expect(verifySecretToProviderAssociation(provider1, secret)).To(BeFalse())
			Expect(verifySecretToProviderAssociation(provider2, secret)).To(BeTrue())

			err = ValidateSopsSecretAbsence("testdata/age/secret-key-2.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Create Private Age-Keys", func() {
			secret, err := LoadFromYAMLFile[*corev1.Secret]("testdata/age/keys/key-1.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-age-secret-1"
			secret.Namespace = "ns-age-provider-1"
			secret.Labels = map[string]string{
				meta.KeySecretLabel: "true",
				"e2e-test":          suiteLabelValue,
				"provider-age":      "1",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			secret2, err := LoadFromYAMLFile[*corev1.Secret]("testdata/age/keys/key-2.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret2.Name = "test-age-secret-2"
			secret2.Namespace = "ns-age-provider-2"
			secret2.Labels = map[string]string{
				meta.KeySecretLabel: "true",
				"e2e-test":          suiteLabelValue,
			}

			Expect(k8sClient.Create(context.TODO(), secret2)).To(Succeed())

			secret3, err := LoadFromYAMLFile[*corev1.Secret]("testdata/age/keys/key-3.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret3.Name = "test-age-secret-3"
			secret3.Namespace = "ns-age-provider-2"
			secret3.Labels = map[string]string{
				meta.KeySecretLabel: "true",
				"e2e-test":          suiteLabelValue,
			}

			Expect(k8sClient.Create(context.TODO(), secret3)).To(Succeed())
		})

		By("Verify AGE-Provider allocation (Key-1)", func() {
			secret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test-age-secret-1", Namespace: "ns-age-provider-1"}, secret)
			Expect(err).ToNot(HaveOccurred())

			Expect(verifyKeyAssociationSuccess(provider1, secret)).To(BeTrue())
			Expect(verifyKeyAssociationSuccess(provider2, secret)).To(BeTrue())
		})

		By("Verify AGE-Provider allocation (Key-2)", func() {
			secret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test-age-secret-2", Namespace: "ns-age-provider-2"}, secret)
			Expect(err).ToNot(HaveOccurred())

			Expect(verifyKeyAssociation(provider1, secret)).To(BeFalse())
			Expect(verifyKeyAssociationSuccess(provider2, secret)).To(BeTrue())
		})

		By("Verify AGE-Provider allocation (Key-3)", func() {
			secret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test-age-secret-3", Namespace: "ns-age-provider-2"}, secret)
			Expect(err).ToNot(HaveOccurred())

			Expect(verifyKeyAssociationFailure(provider2, secret)).To(BeTrue())
		})

		By("Create Encrypted SOPS Secret (Key-1)", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/age/secret-key-1.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-age-secret-1"
			secret.Namespace = "ns-age-provider-1"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "age-1",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			Expect(verifySecretToProviderAssociation(provider1, secret)).To(BeTrue())
			Expect(verifySecretToProviderAssociation(provider2, secret)).To(BeFalse())

			err = ValidateSopsSecret("testdata/age/secret-key-1.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Reverify SOPS Secret (Key-2)", func() {
			secret := &sopsv1alpha1.SopsSecret{}
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test-age-secret-2", Namespace: "ns-age-provider-2"}, secret)
			Expect(err).Should(Succeed())

			Expect(verifySecretToProviderAssociation(provider1, secret)).To(BeFalse())
			Expect(verifySecretToProviderAssociation(provider2, secret)).To(BeTrue())

			err = ValidateSopsSecret("testdata/age/secret-key-2.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Create Multi-Secret (both Key-1 or Key-2)", func() {
			secret, err := LoadFromYAMLFile[*sopsv1alpha1.SopsSecret]("testdata/age/secret-multi.enc.yaml")
			Expect(err).ToNot(HaveOccurred())

			secret.Name = "test-age-multi-1"
			secret.Namespace = "ns-age-provider-1"
			secret.Labels = map[string]string{
				"e2e-test":    suiteLabelValue,
				"secret-type": "age",
			}

			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

			Expect(verifySecretToProviderAssociation(provider1, secret)).To(BeTrue())
			Expect(verifySecretToProviderAssociation(provider2, secret)).To(BeTrue())

			err = ValidateSopsSecret("testdata/age/secret-multi.yaml", secret.Name, secret.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
