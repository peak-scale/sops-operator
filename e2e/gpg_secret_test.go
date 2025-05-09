// Copyright 2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
)

var _ = Describe("SopsProvider Tests", func() {
	JustAfterEach(func() {
		Eventually(func() error {
			poolList := &sopsv1alpha1.SopsProviderList{}
			labelSelector := client.MatchingLabels{"e2e-gpg": "test"}
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
			labelSelector := client.MatchingLabels{"e2e-gpg": "test"}
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
			labelSelector := client.MatchingLabels{"e2e-gpg": "test"}
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

	It("Single Provider Workflow", func() {
		provider := &sopsv1alpha1.SopsProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name: "simple-provider",
				Labels: map[string]string{
					"e2e-gpg": "test",
				},
			},
			Spec: sopsv1alpha1.SopsProviderSpec{
				SOPSSelectors: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"sops-secret": "yeet",
							},
						},
					},
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"namespace-dev": "decryptable",
							},
						},
					},
				},

				ProviderSecrets: []*api.NamespacedSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"sops-source": "yeet",
							},
						},
					},
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"namespace-source": "sops",
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

		By("Get Applied revision", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: provider.Name}, provider)
			Expect(err).Should(Succeed())
		})

		By("Create Namespaces, which where secrets can be sourced from", func() {
			ns1 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-1-provider-solar",
					Labels: map[string]string{
						"e2e-gpg":                   "test",
						"capsule.clastix.io/tenant": "solar",
					},
				},
			}

			err := k8sClient.Create(context.TODO(), ns1)
			Expect(err).Should(Succeed())

			ns2 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-1-provider-secrets",
					Labels: map[string]string{
						"e2e-gpg":          "test",
						"namespace-source": "sops",
					},
				},
			}

			err = k8sClient.Create(context.TODO(), ns2)
			Expect(err).Should(Succeed())

			ns3 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-3-default-pool",
					Labels: map[string]string{
						"e2e-resourcepool":          "test",
						"capsule.clastix.io/tenant": "wind-quota",
					},
				},
			}

			err = k8sClient.Create(context.TODO(), ns3)
			Expect(err).Should(Succeed())
		})

		By("Verify Namespaces are shown as allowed targets", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())

			ok, msg := DeepCompare(namespaces, pool.Status.Namespaces)
			Expect(ok).To(BeTrue(), "Mismatch for expected namespaces: %s", msg)

			Expect(pool.Status.Size).To(Equal(uint(3)))
		})

		By("Verify ResourceQuotas for namespaces", func() {
			rqHardResources := corev1.ResourceList{
				corev1.ResourceLimitsCPU:      resource.MustParse("0"),
				corev1.ResourceLimitsMemory:   resource.MustParse("0"),
				corev1.ResourceRequestsCPU:    resource.MustParse("0"),
				corev1.ResourceRequestsMemory: resource.MustParse("0"),
			}

			quotaLabel, err := utils.GetTypeLabel(&capsulev1beta2.ResourcePool{})
			Expect(err).Should(Succeed())

			for _, ns := range namespaces {
				rq := &corev1.ResourceQuota{}

				err := k8sClient.Get(context.TODO(), client.ObjectKey{
					Name:      utils.PoolResourceQuotaName(pool),
					Namespace: ns},
					rq)
				Expect(err).Should(Succeed())

				Expect(rq.ObjectMeta.Labels[quotaLabel]).To(Equal(pool.Name), "Expected "+quotaLabel+" to be set to "+pool.Name)

				ok, msg := DeepCompare(rqHardResources, rq.Spec.Hard)
				Expect(ok).To(BeTrue(), "Mismatch for resources for resourcequota: %s", msg)

				found := false
				for _, ref := range rq.OwnerReferences {
					if ref.Kind == "ResourcePool" && ref.UID == pool.UID {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "Expected ResourcePool to be owner of ResourceQuota in namespace %s", ns)
			}
		})

		By("Update the ResourcePool", func() {
			pool.Spec.Defaults = corev1.ResourceList{
				corev1.ResourceLimitsCPU:   resource.MustParse("1"),
				corev1.ResourceRequestsCPU: resource.MustParse("1"),
			}

			err := k8sClient.Update(context.TODO(), pool)
			Expect(err).Should(Succeed(), "Failed to update ResourcePool %s", pool)
		})

		By("Get Applied revision", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())
		})

		By("Verify ResourceQuotas for namespaces", func() {
			rqHardResources := corev1.ResourceList{
				corev1.ResourceLimitsCPU:      resource.MustParse("1"),
				corev1.ResourceLimitsMemory:   resource.MustParse("0"),
				corev1.ResourceRequestsCPU:    resource.MustParse("1"),
				corev1.ResourceRequestsMemory: resource.MustParse("0"),
			}

			for _, ns := range namespaces {
				rq := &corev1.ResourceQuota{}

				err := k8sClient.Get(context.TODO(), client.ObjectKey{
					Name:      utils.PoolResourceQuotaName(pool),
					Namespace: ns},
					rq)
				Expect(err).Should(Succeed())

				ok, msg := DeepCompare(rqHardResources, rq.Spec.Hard)
				Expect(ok).To(BeTrue(), "Mismatch for resources for resourcequota: %s", msg)
			}
		})

		By("Remove namespace from being selected (Patch Labels)", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-2-default-pool",
				},
			}

			err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, ns)
			Expect(err).Should(Succeed())

			ns.ObjectMeta.Labels = map[string]string{
				"e2e-resourcepool": "test",
			}

			err = k8sClient.Update(context.TODO(), ns)
			Expect(err).Should(Succeed())

			pool.Spec.Defaults = corev1.ResourceList{
				corev1.ResourceLimitsCPU:   resource.MustParse("1"),
				corev1.ResourceRequestsCPU: resource.MustParse("1"),
			}
		})

		By("Get Applied revision", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())
		})

		By("Verify Namespaces was removed as allowed targets", func() {
			expected := []string{"ns-1-default-pool", "ns-3-default-pool"}

			ok, msg := DeepCompare(expected, pool.Status.Namespaces)
			Expect(ok).To(BeTrue(), "Mismatch for expected namespaces: %s", msg)

			Expect(pool.Status.Size).To(Equal(uint(2)))
		})

		By("Verify ResourceQuota was cleaned up", func() {
			rq := &corev1.ResourceQuota{}
			Eventually(func() error {
				return k8sClient.Get(context.TODO(), client.ObjectKey{
					Name:      utils.PoolResourceQuotaName(pool),
					Namespace: "ns-2-default-pool",
				}, rq)
			}, "30s", "1s").ShouldNot(Succeed(), "Expected ResourceQuota to be deleted from namespace %s", "ns-2-default-pool")
		})

		By("Remove namespace from being selected (Delete Namespace)", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-3-default-pool",
				},
			}

			err := k8sClient.Delete(context.TODO(), ns)
			Expect(err).Should(Succeed())
		})

		By("Get Applied revision", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())
		})

		By("Verify Namespaces was removed as allowed targets", func() {
			expected := []string{"ns-1-default-pool"}

			ok, msg := DeepCompare(expected, pool.Status.Namespaces)
			Expect(ok).To(BeTrue(), "Mismatch for expected namespaces: %s", msg)

			Expect(pool.Status.Size).To(Equal(uint(1)))
		})

		By("Delete Resourcepool", func() {
			err := k8sClient.Delete(context.TODO(), pool)
			Expect(err).Should(Succeed())
		})

		By("Ensure ResourceQuotas are cleaned up", func() {
			for _, ns := range namespaces {
				rq := &corev1.ResourceQuota{}
				Eventually(func() error {
					return k8sClient.Get(context.TODO(), client.ObjectKey{
						Name:      utils.PoolResourceQuotaName(pool),
						Namespace: ns,
					}, rq)
				}, "30s", "1s").ShouldNot(Succeed(), "Expected ResourceQuota to be deleted from namespace %s", ns)
			}
		})

	})

	It("Assigns Defaults correctly (empty)", func() {
		pool := &capsulev1beta2.ResourcePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "no-defaults-pool",
				Labels: map[string]string{
					"e2e-resourcepool": "test",
				},
			},
			Spec: capsulev1beta2.ResourcePoolSpec{
				Selectors: []api.NamespaceSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"capsule.clastix.io/tenant": "solar-quota",
							},
						},
					},
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"capsule.clastix.io/tenant": "wind-quota",
							},
						},
					},
				},
				Config: capsulev1beta2.ResourcePoolSpecConfiguration{
					DefaultsAssignZero: ptr.To(false),
				},
				Quota: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						corev1.ResourceLimitsCPU:      resource.MustParse("2"),
						corev1.ResourceLimitsMemory:   resource.MustParse("2Gi"),
						corev1.ResourceRequestsCPU:    resource.MustParse("2"),
						corev1.ResourceRequestsMemory: resource.MustParse("2Gi"),
					},
				},
			},
		}

		namespaces := []string{"ns-1-zero-pool", "ns-2-zero-pool"}

		By("Create the ResourcePool", func() {
			err := k8sClient.Create(context.TODO(), pool)
			Expect(err).Should(Succeed(), "Failed to create ResourcePool %s", pool)
		})

		By("Get Applied revision", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())
		})

		By("Verify Defaults are empty", func() {
			Expect(pool.Spec.Defaults).To(BeNil(), "Defaults should be empty")
		})

		By("Verify Status was correctly initialized", func() {
			expected := &capsulev1beta2.ResourcePoolQuotaStatus{
				Hard: pool.Spec.Quota.Hard,
				Claimed: corev1.ResourceList{
					corev1.ResourceLimitsCPU:      resource.MustParse("0"),
					corev1.ResourceLimitsMemory:   resource.MustParse("0"),
					corev1.ResourceRequestsCPU:    resource.MustParse("0"),
					corev1.ResourceRequestsMemory: resource.MustParse("0"),
				},
			}

			ok, msg := DeepCompare(*expected, pool.Status.Allocation)
			Expect(ok).To(BeTrue(), "Mismatch for expected status allocation: %s", msg)
		})

		By("Create Namespaces, which are selected by the pool", func() {
			ns1 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-1-zero-pool",
					Labels: map[string]string{
						"e2e-resourcepool":          "test",
						"capsule.clastix.io/tenant": "solar-quota",
					},
				},
			}

			err := k8sClient.Create(context.TODO(), ns1)
			Expect(err).Should(Succeed())

			ns2 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-2-zero-pool",
					Labels: map[string]string{
						"e2e-resourcepool":          "test",
						"capsule.clastix.io/tenant": "wind-quota",
					},
				},
			}

			err = k8sClient.Create(context.TODO(), ns2)
			Expect(err).Should(Succeed())
		})

		By("Verify Namespaces are shown as allowed targets", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())

			ok, msg := DeepCompare(namespaces, pool.Status.Namespaces)
			Expect(ok).To(BeTrue(), "Mismatch for expected namespaces: %s", msg)

			Expect(pool.Status.Size).To(Equal(uint(2)))
		})

		By("Verify ResourceQuotas for namespaces", func() {
			rqHardResources := corev1.ResourceList{}

			quotaLabel, err := utils.GetTypeLabel(&capsulev1beta2.ResourcePool{})
			Expect(err).Should(Succeed())

			for _, ns := range namespaces {
				rq := &corev1.ResourceQuota{}

				err := k8sClient.Get(context.TODO(), client.ObjectKey{
					Name:      utils.PoolResourceQuotaName(pool),
					Namespace: ns},
					rq)
				Expect(err).Should(Succeed())

				Expect(rq.ObjectMeta.Labels[quotaLabel]).To(Equal(pool.Name), "Expected "+quotaLabel+" to be set to "+pool.Name)

				ok, msg := DeepCompare(rqHardResources, rq.Spec.Hard)
				Expect(ok).To(BeTrue(), "Mismatch for resources for resourcequota: %s", msg)

				found := false
				for _, ref := range rq.OwnerReferences {
					if ref.Kind == "ResourcePool" && ref.UID == pool.UID {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "Expected ResourcePool to be owner of ResourceQuota in namespace %s", ns)

			}
		})

		By("Delete Resourcepool", func() {
			err := k8sClient.Delete(context.TODO(), pool)
			Expect(err).Should(Succeed())
		})

		By("Ensure ResourceQuotas are cleaned up", func() {
			for _, ns := range namespaces {
				rq := &corev1.ResourceQuota{}
				Eventually(func() error {
					return k8sClient.Get(context.TODO(), client.ObjectKey{
						Name:      utils.PoolResourceQuotaName(pool),
						Namespace: ns,
					}, rq)
				}, "30s", "1s").ShouldNot(Succeed(), "Expected ResourceQuota to be deleted from namespace %s", ns)
			}
		})

	})

	It("Admission Guards ", func() {
		pool := &capsulev1beta2.ResourcePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "admission-pool",
				Labels: map[string]string{
					"e2e-resourcepool": "test",
				},
			},
			Spec: capsulev1beta2.ResourcePoolSpec{
				Selectors: []api.NamespaceSelector{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"capsule.clastix.io/tenant": "solar-quota",
							},
						},
					},
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"capsule.clastix.io/tenant": "wind-quota",
							},
						},
					},
				},
				Config: capsulev1beta2.ResourcePoolSpecConfiguration{
					DefaultsAssignZero: ptr.To(true),
				},
				Quota: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						corev1.ResourceLimitsCPU:      resource.MustParse("2"),
						corev1.ResourceLimitsMemory:   resource.MustParse("2Gi"),
						corev1.ResourceRequestsCPU:    resource.MustParse("2"),
						corev1.ResourceRequestsMemory: resource.MustParse("2Gi"),
					},
				},
			},
		}

		claim := &capsulev1beta2.ResourcePoolClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "admission-pool-claim-1",
				Namespace: "ns-2-admission-pool",
			},
			Spec: capsulev1beta2.ResourcePoolClaimSpec{
				Pool: "admission-pool",
				ResourceClaims: corev1.ResourceList{
					corev1.ResourceLimitsCPU:      resource.MustParse("1"),
					corev1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
					corev1.ResourceRequestsCPU:    resource.MustParse("1"),
					corev1.ResourceRequestsMemory: resource.MustParse("1Gi"),
				},
			},
		}

		By("Create the ResourcePool", func() {
			err := k8sClient.Create(context.TODO(), pool)
			Expect(err).Should(Succeed(), "Failed to create ResourcePool %s", pool)
		})

		By("Create Namespaces, which are selected by the pool", func() {
			ns1 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-1-admission-pool",
					Labels: map[string]string{
						"e2e-resourcepool":          "test",
						"capsule.clastix.io/tenant": "solar-quota",
					},
				},
			}

			err := k8sClient.Create(context.TODO(), ns1)
			Expect(err).Should(Succeed())

			ns2 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-2-admission-pool",
					Labels: map[string]string{
						"e2e-resourcepool":          "test",
						"capsule.clastix.io/tenant": "wind-quota",
					},
				},
			}

			err = k8sClient.Create(context.TODO(), ns2)
			Expect(err).Should(Succeed())
		})

		By("Claim some Resources", func() {
			err := k8sClient.Create(context.TODO(), claim)
			Expect(err).Should(Succeed(), "Failed to create ResourcePoolClaim %s", claim)
		})

		By("Get Applied revisions", func() {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: pool.Name}, pool)
			Expect(err).Should(Succeed())

			err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: claim.Name, Namespace: claim.Namespace}, claim)
			Expect(err).Should(Succeed())
		})

		By("Verify ResourcePool Status Allocation", func() {
			expectedAllocation := capsulev1beta2.ResourcePoolQuotaStatus{
				Hard: corev1.ResourceList{
					corev1.ResourceLimitsCPU:      resource.MustParse("2"),
					corev1.ResourceLimitsMemory:   resource.MustParse("2Gi"),
					corev1.ResourceRequestsCPU:    resource.MustParse("2"),
					corev1.ResourceRequestsMemory: resource.MustParse("2Gi"),
				},
				Claimed: corev1.ResourceList{
					corev1.ResourceLimitsCPU:      resource.MustParse("1"),
					corev1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
					corev1.ResourceRequestsCPU:    resource.MustParse("1"),
					corev1.ResourceRequestsMemory: resource.MustParse("1Gi"),
				},
			}

			ok, msg := DeepCompare(pool.Status.Allocation, expectedAllocation)
			Expect(ok).To(BeTrue(), "Mismatch for resource allocation: %s", msg)

			//expectedClaims := map[string]capsulev1beta2.ResourcePoolClaimsList{}
			expectedClaims := map[string]capsulev1beta2.ResourcePoolClaimsList{
				claim.Namespace: {
					&capsulev1beta2.ResourcePoolClaimsItem{
						StatusNameUID: api.StatusNameUID{
							Name: api.Name(claim.GetName()),
							UID:  claim.GetUID(),
						},
						Claims: corev1.ResourceList{
							corev1.ResourceLimitsCPU:      resource.MustParse("1"),
							corev1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
							corev1.ResourceRequestsCPU:    resource.MustParse("1"),
							corev1.ResourceRequestsMemory: resource.MustParse("1Gi"),
						},
					},
				},
			}

			ok, msg = DeepCompare(expectedClaims, pool.Status.Claims)
			Expect(ok).To(BeTrue(), "Mismatch for expected claims: %s", msg)
		})

		By("Verify ResourceQuotas for namespaces", func() {

			quotaLabel, err := utils.GetTypeLabel(&capsulev1beta2.ResourcePool{})
			Expect(err).Should(Succeed())

			rq1 := &corev1.ResourceQuota{}
			err = k8sClient.Get(context.TODO(), client.ObjectKey{
				Name:      utils.PoolResourceQuotaName(pool),
				Namespace: "ns-1-admission-pool",
			}, rq1)
			Expect(err).Should(Succeed())

			Expect(rq1.ObjectMeta.Labels[quotaLabel]).To(Equal(pool.Name), "Expected "+quotaLabel+" to be set to "+pool.Name)

			rqHardResources1 := corev1.ResourceList{
				corev1.ResourceLimitsCPU:      resource.MustParse("0"),
				corev1.ResourceLimitsMemory:   resource.MustParse("0"),
				corev1.ResourceRequestsCPU:    resource.MustParse("0"),
				corev1.ResourceRequestsMemory: resource.MustParse("0"),
			}

			ok, msg := DeepCompare(rqHardResources1, rq1.Spec.Hard)
			Expect(ok).To(BeTrue(), "Mismatch for resources for resourcequota: %s", msg)

			found := false
			for _, ref := range rq1.OwnerReferences {
				if ref.Kind == "ResourcePool" && ref.UID == pool.UID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Expected ResourcePool to be owner of ResourceQuota in namespace %s", "ns-1-admission-pool")

			rq2 := &corev1.ResourceQuota{}
			err = k8sClient.Get(context.TODO(), client.ObjectKey{
				Name:      utils.PoolResourceQuotaName(pool),
				Namespace: "ns-2-admission-pool",
			}, rq2)
			Expect(err).Should(Succeed())

			Expect(rq2.ObjectMeta.Labels[quotaLabel]).To(Equal(pool.Name), "Expected "+quotaLabel+" to be set to "+pool.Name)

			rqHardResources2 := corev1.ResourceList{
				corev1.ResourceLimitsCPU:      resource.MustParse("1"),
				corev1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
				corev1.ResourceRequestsCPU:    resource.MustParse("1"),
				corev1.ResourceRequestsMemory: resource.MustParse("1Gi"),
			}

			ok, msg = DeepCompare(rqHardResources2, rq2.Spec.Hard)
			Expect(ok).To(BeTrue(), "Mismatch for resources for resourcequota: %s", msg)

			found = false
			for _, ref := range rq2.OwnerReferences {
				if ref.Kind == "ResourcePool" && ref.UID == pool.UID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Expected ResourcePool to be owner of ResourceQuota in namespace %s", "ns-2-admission-pool")
		})
	})
})
