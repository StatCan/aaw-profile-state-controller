package controller

import (
	"context"
	"strconv"
	"strings"

	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const SAS_PREFIX = "k8scc01covidacr.azurecr.io/sas:"
const FEATURES_LABEL = "state.aaw.statcan.gc.ca/employee-only-features"
const RB_LABEL = "state.aaw.statcan.gc.ca/non-employee-users"

var employeeDomains [2]string = [2]string{"cloud.statcan.ca", "statcan.gc.ca"}

func sasImage(pod *corev1.Pod) bool {

	for _, container := range pod.Spec.Containers {
		sasImage := strings.HasPrefix(container.Image, SAS_PREFIX)
		if sasImage {
			return true
		}
	}

	return false
}

func internalUser(email string) bool {
	for _, domain := range employeeDomains {
		if strings.HasSuffix(email, domain) {
			return true
		}
	}
	return false
}

func isEmployee(rolebinding *rbacv1.RoleBinding) bool {
	for _, subject := range rolebinding.Subjects {
		email := subject.Name
		if strings.Contains(email, "@") {
			if !internalUser(email) {
				return false
			}
		}
	}
	return true
}

func (c *Controller) hasEmployeeOnlyFeatures(pods []*corev1.Pod) bool {
	// label to set
	hasEmpOnlyFeatures := false

	for _, pod := range pods {
		if sasImage(pod) {
			hasEmpOnlyFeatures = true
			break
		}
	}

	return hasEmpOnlyFeatures
}

func (c *Controller) hasNonEmployeeUser(roleBindings []*rbacv1.RoleBinding) bool {
	// label to set
	nonEmployeeUser := false

	for _, roleBindings := range roleBindings {
		if !isEmployee(roleBindings) {
			nonEmployeeUser = true
			break
		}
	}

	return nonEmployeeUser
}

func (c *Controller) handleProfileAndNamespace(profile *v1.Profile, namespace *corev1.Namespace, hasEmployeeOnlyFeatures bool, isNonEmployeeUser bool) error {
	// set namespace labels
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string)
	}
	// set Profile labels
	if profile.Labels == nil {
		profile.Labels = make(map[string]string)
	}
	// Update profile and namespace labels
	profile.Labels[FEATURES_LABEL] = strconv.FormatBool(hasEmployeeOnlyFeatures)
	profile.Labels[RB_LABEL] = strconv.FormatBool(isNonEmployeeUser)
	namespace.Labels[FEATURES_LABEL] = strconv.FormatBool(hasEmployeeOnlyFeatures)
	namespace.Labels[RB_LABEL] = strconv.FormatBool(isNonEmployeeUser)

	ctx := context.Background()
	// Update profile and namespace resources
	_, err := c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, profile, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	log.Infof("Updated profile %v with labels", namespace.Name)

	_, err = c.kubeclientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	log.Infof("Updated namespace %v with labels", namespace.Name)

	return nil
}
