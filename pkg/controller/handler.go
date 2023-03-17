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

// Declare capability labels
const HAS_SAS_NOTEBOOK_FEATURE_LABEL = "state.aaw.statcan.gc.ca/has-sas-notebook-feature"
const EXISTS_NON_SAS_NOTEBOOK_USER_LABEL = "state.aaw.statcan.gc.ca/exists-non-sas-notebook-user"
const EXISTS_NON_CLOUD_MAIN_USER_LABEL = "state.aaw.statcan.gc.ca/exists-non-cloud-main-user"

var employeeDomains [2]string = [2]string{"cloud.statcan.ca", "statcan.gc.ca"}

//  ____    _    ____    _   _       _       _                 _
// / ___|  / \  / ___|  | \ | | ___ | |_ ___| |__   ___   ___ | | __
// \___ \ / _ \ \___ \  |  \| |/ _ \| __/ _ \ '_ \ / _ \ / _ \| |/ /
//  ___) / ___ \ ___) | | |\  | (_) | ||  __/ |_) | (_) | (_) |   <
// |____/_/   \_\____/  |_| \_|\___/ \__\___|_.__/ \___/ \___/|_|\_\

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

func (c *Controller) subjectInSasNotebookExceptionList(subject string) bool {
	for _, exceptionCase := range c.nonEmployeeExceptions["sasNotebookExceptions"] {
		if subject == exceptionCase {
			return true
		}
	}
	log.Infof("Found unexcepted SAS user %v", subject)
	return false
}

func (c *Controller) rolebindingContainsNonSasUser(rolebinding *rbacv1.RoleBinding) bool {
	for _, subject := range rolebinding.Subjects {
		// If subject.Kind is not a user, then nothing below applies
		if subject.Kind != "User" {
			continue
		}
		// If the subject contains a Statcan employee email, there is nothing more to check.
		// Continue to the next iteration
		email := subject.Name
		if strings.Contains(email, "@") {
			if internalUser(email) {
				continue
			}
		}
		// If the subject is in the exception list for SAS users, then we can continue to the next
		// iteration
		if c.subjectInSasNotebookExceptionList(subject.Name) {
			continue
		}
		// If we get to this point, the user is not a statcan employee and the user has not
		// been granted an exception to use the SAS feeature. This is a sufficient condition
		// for rolebindingContainsNonSasUser to return true.
		return true
	}
	return false
}

func (c *Controller) hasSasNotebookFeature(pods []*corev1.Pod) bool {
	// label to set
	hasSasNotebookFeature := false

	for _, pod := range pods {
		if sasImage(pod) {
			hasSasNotebookFeature = true
			break
		}
	}

	return hasSasNotebookFeature
}

func (c *Controller) existsNonSasUser(roleBindings []*rbacv1.RoleBinding) bool {
	// label to set
	nonSasUser := false

	for _, roleBinding := range roleBindings {
		if c.rolebindingContainsNonSasUser(roleBinding) {
			nonSasUser = true
			break
		}
	}

	return nonSasUser
}

//       _                 _                   _
//   ___| | ___  _   _  __| |  _ __ ___   __ _(_)_ __
//  / __| |/ _ \| | | |/ _` | | '_ ` _ \ / _` | | '_ \
// | (__| | (_) | |_| | (_| | | | | | | | (_| | | | | |
//  \___|_|\___/ \__,_|\__,_| |_| |_| |_|\__,_|_|_| |_|

func (c *Controller) subjectInCloudMainExceptionList(subject string) bool {
	for _, exceptionCase := range c.nonEmployeeExceptions["cloudMainExceptions"] {
		if subject == exceptionCase {
			return true
		}
	}
	log.Infof("Found unexcepted cloudmain user %v", subject)
	return false
}

func (c *Controller) rolebindingContainsNonCloudMainUser(rolebinding *rbacv1.RoleBinding) bool {
	for _, subject := range rolebinding.Subjects {
		// If subject.Kind is not a user, then nothing below applies
		if subject.Kind != "User" {
			continue
		}
		// If the subject contains a Statcan employee email, there is nothing more to check.
		// Continue to the next iteration
		email := subject.Name
		if strings.Contains(email, "@") {
			if internalUser(email) {
				continue
			}
		}
		// If the subject is in the exception list for SAS users, then we can continue to the next
		// iteration
		if c.subjectInCloudMainExceptionList(subject.Name) {
			continue
		}
		// If we get to this point, the user is not a statcan employee and the user has not
		// been granted an exception to use the SAS feeature. This is a sufficient condition
		// for rolebindingContainsNonSasUser to return true.
		return true
	}
	return false
}

func (c *Controller) existsNonCloudMainUser(roleBindings []*rbacv1.RoleBinding) bool {
	// label to set
	nonCloudMainUser := false

	for _, roleBinding := range roleBindings {
		if c.rolebindingContainsNonCloudMainUser(roleBinding) {
			nonCloudMainUser = true
			break
		}
	}

	return nonCloudMainUser
}

//              _                     _ _
//  _ __  ___  | |__   __ _ _ __   __| | | ___ _ __
// | '_ \/ __| | '_ \ / _` | '_ \ / _` | |/ _ \ '__|
// | | | \__ \ | | | | (_| | | | | (_| | |  __/ |
// |_| |_|___/ |_| |_|\__,_|_| |_|\__,_|_|\___|_|

func (c *Controller) handleProfileAndNamespace(profile *v1.Profile, namespace *corev1.Namespace, hasSasNotebookFeature bool, existsNonSasUser bool, existsNonCloudMainUser bool) error {
	// set namespace labels
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string)
	}
	// set Profile labels
	if profile.Labels == nil {
		profile.Labels = make(map[string]string)
	}
	// Update profile labels
	profile.Labels[HAS_SAS_NOTEBOOK_FEATURE_LABEL] = strconv.FormatBool(hasSasNotebookFeature)
	profile.Labels[EXISTS_NON_SAS_NOTEBOOK_USER_LABEL] = strconv.FormatBool(existsNonSasUser)
	profile.Labels[EXISTS_NON_CLOUD_MAIN_USER_LABEL] = strconv.FormatBool(existsNonCloudMainUser)

	// Update namespace labels
	namespace.Labels[HAS_SAS_NOTEBOOK_FEATURE_LABEL] = strconv.FormatBool(hasSasNotebookFeature)
	namespace.Labels[EXISTS_NON_SAS_NOTEBOOK_USER_LABEL] = strconv.FormatBool(existsNonSasUser)
	namespace.Labels[EXISTS_NON_CLOUD_MAIN_USER_LABEL] = strconv.FormatBool(existsNonCloudMainUser)

	ctx := context.Background()
	// Update profile and namespace resources
	_, err := c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, profile, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	log.Infof("Updated profile %v with labels hasSasNotebookFeature=%b existsNonSasUser=%b existsNonCloudMainUser=%b",
		namespace.Name, hasSasNotebookFeature, existsNonSasUser, existsNonCloudMainUser)

	_, err = c.kubeclientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	log.Infof("Updated namespace %v with labels hasSasNotebookFeature=%b existsNonSasUser=%b existsNonCloudMainUser=%b",
		namespace.Name, hasSasNotebookFeature, existsNonSasUser, existsNonCloudMainUser)

	return nil
}
