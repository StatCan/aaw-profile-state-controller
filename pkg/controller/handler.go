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
	"k8s.io/apimachinery/pkg/labels"
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

func (c *Controller) handleProfile(profile *v1.Profile) error {
	namespace := profile.Name

	// labels to set in Profile
	hasEmpOnlyFeatures := false
	nonEmployeeUser := false

	// check Pods
	pods, err := c.podLister.Pods(namespace).List(labels.Everything())

	if err != nil {
		return err
	}

	for _, pod := range pods {
		if sasImage(pod) {
			hasEmpOnlyFeatures = true
			break
		}
	}

	// check RoleBindings
	roleBindings, err := c.roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		return err
	}

	for _, roleBindings := range roleBindings {
		if !isEmployee(roleBindings) {
			nonEmployeeUser = true
			break
		}
	}

	// set Profile labels
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
