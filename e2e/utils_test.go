//nolint:all
package e2e_test

import (
	"context"
	"fmt"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultTimeoutInterval = 20 * time.Second
	defaultPollInterval    = time.Second
	e2eLabel               = "argo.addons.projectcapsule.dev/e2e"
	suiteLabel             = "e2e.argo.addons.projectcapsule.dev/suite"
)

func e2eConfigName() string {
	return "default"
}

// Returns labels to identify e2e resources.
func e2eLabels(suite string) (labels map[string]string) {
	labels = make(map[string]string)
	labels["argo.addons.projectcapsule.dev/e2e"] = "true"

	if suite != "" {
		labels["e2e.argo.addons.projectcapsule.dev/suite"] = suite
	}

	return
}

// Returns a label selector to filter e2e resources.
func e2eSelector(suite string) labels.Selector {
	return labels.SelectorFromSet(e2eLabels(suite))
}

// Pass objects which require cleanup and a label selector to filter them.
func cleanResources(res []client.Object, selector labels.Selector) (err error) {
	for _, resource := range res {
		err = k8sClient.DeleteAllOf(context.TODO(), resource, &client.MatchingLabels{"argo.addons.projectcapsule.dev/e2e": "true"})

		if err != nil {
			return err
		}
	}

	return nil
}

func CleanTranslators(selector labels.Selector) error {
	res := &v1alpha1.ArgoTranslatorList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list translators: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete translator %s: %w", app.GetName(), err)
		}
	}

	return nil
}

func CleanTenants(selector labels.Selector) error {
	res := &capsulev1beta2.TenantList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete tenant %s: %w", app.GetName(), err)
		}
	}

	return nil
}

func CleanAppProjects(selector labels.Selector, namespace string) error {
	res := &argocdv1alpha1.AppProjectList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// If a namespace is provided, set it in the list options
	if namespace != "" {
		listOptions.Namespace = namespace
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete resource %s: %w", app.GetName(), err)
		}
	}

	return nil
}
