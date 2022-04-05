package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"strconv"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func podProfileName(pod *corev1.Pod) string {
	return fmt.Sprintf("%s", pod.Namespace)
}

func sasImage(pod *corev1.Pod) string {
	image := pod.Spec.Containers[0].Image
	sasImage := strings.HasPrefix(image, "k8scc01covidacr.azurecr.io/sas:")
	return strconv.FormatBool(sasImage)
}

func (c *Controller) handleProfile(pod *corev1.Pod) error {
	ctx := context.Background()
	
	//Find any profile with the same name

	existingProfiles, err := c.profileInformerLister.Lister().Get(podProfileName(pod))
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	//Check that we own this profile
	/*if ap != nil {
		if !metav1.IsControlledBy(ap, pod) {
			msg := fmt.Sprintf("Profile \"%s/%s\" already exists and is not managed by the pod", ap.Namespace, ap.Name)
			c.recorder.Event(pod, corev1.EventTypeWarning, "ErrResourceExists", msg)
			return fmt.Errorf("%s", msg)
		}
	}*/

	//New Profile
	newProfile, err := c.generateProfile(pod)
	if err != nil {
		return err
	}

	// If we don't have a profile, then let's make one
	if existingProfiles == nil {
		log.Infof("create profile \"%s\"", newProfile.Name)

		_, err = c.kubeflowClientset.KubeflowV1().Profiles().Create(ctx, newProfile, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !reflect.DeepEqual(existingProfiles.Spec, newProfile.Spec) { //We have a profile, but is it the same
		log.Infof("updated profile \"%s\"", existingProfiles.Name)

		// Copy the new spec
		existingProfiles.Spec = newProfile.Spec

		_, err =  c.kubeflowClientset.KubeflowV1().Profiles().Update(ctx, existingProfiles, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}


func (c *Controller) generateProfile(pod *corev1.Pod)(*v1.Profile, error){

	existingProfiles := &v1.Profile{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"state.aaw.statcan.gc.ca": sasImage(pod),
			},
			Name: podProfileName(pod),
			
		},
		Spec: v1.ProfileSpec{
			Owner: rbacv1.Subject{
				Kind: "User",
				Name: "test",
			},
			ResourceQuotaSpec: corev1.ResourceQuotaSpec{},
		},
	}
	return existingProfiles, nil
}
