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

func sasImage(pod *corev1.Pod) bool {
	image := pod.Spec.Containers[0].Image
	sasImage := strings.HasPrefix(image, "k8scc01covidacr.azurecr.io/sas:")
	return sasImage
}

func (c *Controller) handleProfile(profile *v1.Profile) error {
	ctx := context.Background()

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

	profile.Labels["state.aaw.statcan.gc.ca/employee-only-features"] = strconv.FormatBool(hasEmpOnlyFeatures)

	_, err = c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, profile, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	log.Infof("Updated profile %v with label", namespace)

	return nil
}
