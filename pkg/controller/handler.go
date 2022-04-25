package controller

import (
	"context"
	"strconv"
	"strings"

	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const SAS_PREFIX = "k8scc01covidacr.azurecr.io/sas:"
const FEATURES_LABEL = "state.aaw.statcan.gc.ca/employee-only-features"
const RB_LABEL = "state.aaw.statcan.gc.ca/non-employee-user"

func sasImage(pod *corev1.Pod) bool {

	for _, container := range pod.Spec.Containers {
		sasImage := strings.HasPrefix(container.Image, SAS_PREFIX)
		if sasImage {
			return true
		}
	}

	return false
}

func isEmployee(email string) bool {

	employeeDomains := [2]string{"cloud.statcan.ca", "statcan.gc.ca"} // Intialized with values
	for _, domain := range employeeDomains {
		if strings.HasSuffix(email, domain) {
			return true
		}
	}
	return false
}

func (c *Controller) handleProfile(profile *v1.Profile) error {
	namespace := profile.Name
	hasEmpOnlyFeatures := false

	// label set from rolebindings, if there are no rolebindings, then the default is that they are external
	nonEmployeeUser := true

	pods, err := c.podLister.Pods(namespace).List(labels.Everything())

	if err != nil {
		return err
	}

	roleBindings, err := c.roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		return err
	}

	for _, pod := range pods {
		if sasImage(pod) {
			hasEmpOnlyFeatures = true
			break
		}
	}

	for _, roleBindings := range roleBindings {
		for _, subject := range roleBindings.Subjects {
			email := subject.Name
			if strings.Contains(email, "@") {
				if isEmployee(email) {
					nonEmployeeUser = false
					break
				}
			}
		}
	}

	if profile.Labels == nil {
		profile.Labels = make(map[string]string)
	}

	profile.Labels[FEATURES_LABEL] = strconv.FormatBool(hasEmpOnlyFeatures)
	profile.Labels[RB_LABEL] = strconv.FormatBool(nonEmployeeUser)

	ctx := context.Background()
	_, err = c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, profile, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	log.Infof("Updated profile %v with labels", namespace)

	return nil
}
