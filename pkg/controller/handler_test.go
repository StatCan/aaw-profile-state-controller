/*
These tests make use of hasEmployeeOnlyFeatures and isNonEmployeeUser with mocked rolebindings and pods.

Can parse Pod/RoleBinding specs from YAML files in tests/ folder and use the same YAML files to do an E2E
test against a local k8s cluster like k3d.
*/

package controller

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

//        _   _ _
//  _   _| |_(_) |___
// | | | | __| | / __|
// | |_| | |_| | \__ \
// \__,_|\__|_|_|___/

const TEST_DIRECTORY = "../../tests/"

var mockController = Controller{
	nonEmployeeExceptions: UnmarshalConf(filepath.Join(TEST_DIRECTORY, "non-employee-exceptions.yaml")),
}

// Load kubernetes object from YAML file
func loadObjectFromYaml(filePath string) (runtime.Object, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlFile), nil, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
		return nil, err
	}
	return obj, err
}

// Load pod spec from yaml file
func getPod(filePath string) (*corev1.Pod, error) {
	obj, _ := loadObjectFromYaml(filePath)
	pod := obj.(*corev1.Pod)
	return pod, nil
}

// Load all pods in a test folder
func getPods(filePath string) ([]*corev1.Pod, error) {
	// Initialize list of pods
	pods := []*corev1.Pod{}
	items, _ := ioutil.ReadDir(filePath)
	// Load each pod in
	for _, item := range items {
		pod, _ := getPod(filepath.Join(filePath, item.Name()))
		pods = append(pods, pod)
	}
	return pods, nil
}

// Load rolebinding spec from yaml file
func getRolebinding(filePath string) (*rbacv1.RoleBinding, error) {
	obj, _ := loadObjectFromYaml(filePath)
	rolebinding := obj.(*rbacv1.RoleBinding)
	return rolebinding, nil
}

// Load all rolebindings in a test folder
func getRolebindings(filePath string) ([]*rbacv1.RoleBinding, error) {
	// Initialize list of pods
	rolebindings := []*rbacv1.RoleBinding{}
	items, _ := ioutil.ReadDir(filePath)
	// Load each pod in
	for _, item := range items {
		rolebinding, _ := getRolebinding(filepath.Join(filePath, item.Name()))
		rolebindings = append(rolebindings, rolebinding)
	}
	return rolebindings, nil
}

//  _            _
// | |_ ___  ___| |_ ___
// | __/ _ \/ __| __/ __|
// | ||  __/\__ \ |_\__ \
//  \__\___||___/\__|___/

// If any pod in the list uses a SAS image, hasEmployeeOnlyFeatures should return true
func TestAnyPodWithSASImageReturnsTrue(t *testing.T) {
	pods, _ := getPods(filepath.Join(TEST_DIRECTORY, "1"))
	result := mockController.hasEmployeeOnlyFeatures(pods)
	if !result {
		t.Fatalf("Expected hasEmployeeOnlyFeatures to return true because at least one pod contains a SAS image.")
	}
}

func TestNoPodWithSASImageReturnsFalse(t *testing.T) {
	pods, _ := getPods(filepath.Join(TEST_DIRECTORY, "2"))
	result := mockController.hasEmployeeOnlyFeatures(pods)
	if result {
		t.Fatalf("Expected hasEmployeeOnlyFeatures to return false because no pods contain a SAS image.")
	}
}

func TestEmptyPodListReturnsFalse(t *testing.T) {
	pods := []*corev1.Pod{}
	result := mockController.hasEmployeeOnlyFeatures(pods)
	if result {
		t.Fatalf("Expected hasEmployeeOnlyFeatures to return false because empty list of pods can't contain SAS image.")
	}
}

func TestAnyRolebindingWithNonEmployeeReturnsTrue(t *testing.T) {
	rolebindings, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "4"))
	result := mockController.hasNonEmployeeUser(rolebindings)
	if !result {
		t.Fatalf("Expected hasNonEmployeeUser to return true because at least one rolebinding contains a non-employee user.")
	}
}

func TestNoRolebindingWithNonEmployeeReturnsFalse(t *testing.T) {
	rolebindings, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "5"))
	result := mockController.hasNonEmployeeUser(rolebindings)
	if result {
		t.Fatalf("Expected hasNonEmployeeUser to return false because no rolebindings contain a non-employee user.")
	}
}

func TestEmptyRolebindingListReturnsFalse(t *testing.T) {
	rolebindings := []*rbacv1.RoleBinding{}
	result := mockController.hasNonEmployeeUser(rolebindings)
	if result {
		t.Fatalf("Expected hasNonEmployeeUser to return false because empty list of rolebindings contain a non-employee user.")
	}
}
