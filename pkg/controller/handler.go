package controller

import (
	"context"
	"fmt"
	//"reflect"
	"strings"
	"strconv"
	//"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	//rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func podProfileName(pod *corev1.Pod) string {
	return fmt.Sprintf("%s", pod.Namespace)
}

func sasImage(pod *corev1.Pod) bool {
	image := pod.Spec.Containers[0].Image
	sasImage := strings.HasPrefix(image, "k8scc01covidacr.azurecr.io/sas:")
	return sasImage
}

func (c *Controller) handleProfile(pod *corev1.Pod) error {
	ctx := context.Background()
	
	namespace := pod.GetNamespace()
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

	existingProfiles, err := c.profileInformerLister.Lister().Get(namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if existingProfiles != nil {
	
		if existingProfiles != nil {
			newProfile, err := c.generateProfile(pod, existingProfiles, hasEmpOnlyFeatures)
			if err != nil {
				return err
			}

			// Copy the new spec
			existingProfiles.ObjectMeta = newProfile.ObjectMeta

			_, err =  c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, existingProfiles, metav1.UpdateOptions{})
			if err != nil {
			return err
		}
			//existingLabels["state.aaw.statcan.gc.ca"] = strconv.FormatBool(hasEmpOnlyFeatures)
			//existingProfiles.SetLabels(existingLabels)
		}
	}

	return nil
}


func (c *Controller) generateProfile(pod *corev1.Pod, profile *v1.Profile, hasEmpOnlyFeatures bool)(*v1.Profile, error){

	existingProfiles := &v1.Profile{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"state.aaw.statcan.gc.ca": strconv.FormatBool(hasEmpOnlyFeatures),
			},
			Name: profile.ObjectMeta.Name,
			ResourceVersion: profile.ObjectMeta.ResourceVersion,
		},
	}
	return existingProfiles, nil
}
