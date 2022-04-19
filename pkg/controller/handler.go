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
const PROFILE_LABEL = "state.aaw.statcan.gc.ca/employee-only-features"

func sasImage(pod *corev1.Pod) bool {
	// TODO: how do we know the 0th container is the SAS pod?
	// should range over all containers and do this check.
	image := pod.Spec.Containers[0].Image
	// TODO: should this string be moved to a global constant in case
	// it needs to be changed?
	sasImage := strings.HasPrefix(image, SAS_PREFIX)
	return sasImage
}

func (c *Controller) handleProfile(profile *v1.Profile) error {
	namespace := profile.Name
	hasEmpOnlyFeatures := false

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

	if profile.Labels == nil {
		profile.Labels = make(map[string]string)
	}

	profile.Labels[PROFILE_LABEL] = strconv.FormatBool(hasEmpOnlyFeatures)
	// TODO: should we get the context closer to where it is used in the code?
	ctx := context.Background()
	_, err = c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, profile, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	log.Infof("Updated profile %v with label", namespace)

	return nil
}
