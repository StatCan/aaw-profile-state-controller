/*
These tests make use of hasSasNotebookFeature and existsNonSasUser with mocked rolebindings and pods.

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
		log.Fatalf(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
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

func getPVC(filePath string) (*corev1.PersistentVolumeClaim, error) {
	obj, err := loadObjectFromYaml(filePath)
	if err != nil {
		return nil, err
	}
	pvc := obj.(*corev1.PersistentVolumeClaim)
	return pvc, nil
}

//  _            _
// | |_ ___  ___| |_ ___
// | __/ _ \/ __| __/ __|
// | ||  __/\__ \ |_\__ \
//  \__\___||___/\__|___/

// If any pod in the list uses a SAS image, hasSasNotebookFeature should return true
func TestAnyPodWithSASImageReturnsTrue(t *testing.T) {
	pods, _ := getPods(filepath.Join(TEST_DIRECTORY, "1"))
	result := mockController.hasSasNotebookFeature(pods)
	if !result {
		t.Fatalf("Expected hasSasNotebookFeature to return true because at least one pod contains a SAS image.")
	}
}

func TestNoPodWithSASImageReturnsFalse(t *testing.T) {
	pods, _ := getPods(filepath.Join(TEST_DIRECTORY, "2"))
	result := mockController.hasSasNotebookFeature(pods)
	if result {
		t.Fatalf("Expected hasSasNotebookFeature to return false because no pods contain a SAS image.")
	}
}

func TestEmptyPodListReturnsFalse(t *testing.T) {
	pods := []*corev1.Pod{}
	result := mockController.hasSasNotebookFeature(pods)
	if result {
		t.Fatalf("Expected hasSasNotebookFeature to return false because empty list of pods can't contain SAS image.")
	}
}

func TestAnyRolebindingWithNonEmployeeReturnsTrue(t *testing.T) {
	rolebindings, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "4"))
	result := mockController.existsNonSasUser(rolebindings)
	if !result {
		t.Fatalf("Expected existsNonSasUser to return true because at least one rolebinding contains a non-employee user.")
	}
}

func TestNoRolebindingWithNonEmployeeReturnsFalse(t *testing.T) {
	rolebindings, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "5"))
	result := mockController.existsNonSasUser(rolebindings)
	if result {
		t.Fatalf("Expected existsNonSasUser to return false because no rolebindings contain a non-employee user.")
	}
}

func TestEmptyRolebindingListReturnsFalse(t *testing.T) {
	rolebindings := []*rbacv1.RoleBinding{}
	result := mockController.existsNonSasUser(rolebindings)
	if result {
		t.Fatalf("Expected existsNonSasUser to return false because empty list of rolebindings contain a non-employee user.")
	}
}

// expect true for existsNonEmployee return value
func TestExistsNonEmployeeTrue(t *testing.T) {
	rolebindings, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "4"))
	result := mockController.existsNonEmployee(rolebindings)
	if !result {
		t.Fatalf("Expected existsNonEmployee to return true because at least one rolebinding contains a non-employee user.")
	}
}

// expect false for existsNonEmployee return value
func TestExistsNonEmployeeFalse(t *testing.T) {
	rolebindings, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "5"))
	result := mockController.existsNonEmployee(rolebindings)
	if result {
		t.Fatalf("Expected existsNonEmployee to return false because no rolebindings contain a non-employee user.")
	}
}

func TestExistsNonEmployeeEmpty(t *testing.T) {
	rolebindings := []*rbacv1.RoleBinding{}
	result := mockController.existsNonEmployee(rolebindings)
	if result {
		t.Fatalf("Expected existsNonEmployee to return false because empty list of rolebindings contain a non-employee user.")
	}
}

// can't actually test the existsInternalCommonStorage func since these tests simply use the yaml object
// Instead we test the internalPVC() func which contains the selection logic
func TestExistsInternalPVC(t *testing.T) {
	// test for both "iunc" and "iprotb"
	iuncPVC, _ := getPVC(filepath.Join(TEST_DIRECTORY, "blob/1/iunc_pvc_exists.yaml"))
	result1 := mockController.internalPVC(iuncPVC.Name)
	iprotBPVC, _ := getPVC(filepath.Join(TEST_DIRECTORY, "blob/1/iprotb_pvc_exists.yaml"))
	result2 := mockController.internalPVC(iprotBPVC.Name)
	if !(result1 && result2) {
		t.Fatalf("Expected internalPVC to return true because the PVCs %s and %s contains substring `iunc`", iuncPVC.Name, iprotBPVC)
	}
}

func TestNotExistsInternalPVC(t *testing.T) {
	pvc, _ := getPVC(filepath.Join(TEST_DIRECTORY, "blob/2/internal_pvc_not_exists.yaml"))
	result := mockController.internalPVC(pvc.Name)
	if result {
		t.Fatalf("Expected internalPVC to return false because the PVC %s does not contain the needed substring", pvc.Name)
	}
}

//    					    _   _
//   _____  _____ ___ _ __ | |_(_) ___  _ __     ___ __ _ ___  ___  ___
//  / _ \ \/ / __/ _ \ '_ \| __| |/ _ \| '_ \   / __/ _` / __|/ _ \/ __|
// |  __/>  < (_|  __/ |_) | |_| | (_) | | | | | (_| (_| \__ \  __/\__ \
//  \___/_/\_\___\___| .__/ \__|_|\___/|_| |_|  \___\__,_|___/\___||___/
// 				    |_|

// If rolebinding contains only statcan domains --> profiles-state-controller should produce the labels
// state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: false and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: false

func TestStatCanEmployeeOnlyReturnsFalseAndFalse(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_1"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == False AND cloudMainResult == False, we are OK. Otherwise the test failed
	if !sasResult && !cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: false and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: false")
}

// If rolebinding contains statcan domains AND a user with only the SAS exception --> profiles-state-controller should
// produce the labels state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: false and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true

func TestStatCanEmployeeAndSasExceptionReturnsFalseAndTrue(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_2"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == False AND cloudMainResult == True, we are OK. Otherwise the test failed
	if !sasResult && cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: false and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true")
}

// If rolebinding contains statcan domains AND a user with only the cloud main exception --> profiles-state-controller
// should produce the labels state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: false

func TestStatCanEmployeeAndCloudMainExceptionReturnsTrueAndFalse(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_3"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == True AND cloudMainResult == False, we are OK. Otherwise the test failed
	if sasResult && !cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: false")

}

// If rolebinding contains statcan domains AND a user with only the cloud main exception AND a user with only the SAS
// exception --> profiles-state-controller should produce the labels
// state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true

func TestStatCanEmployeeAndCloudMainExceptionAndSasExceptionReturnsTrueAndTrue(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_4"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == True AND cloudMainResult == True, we are OK. Otherwise the test failed
	if sasResult && cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true")
}

// If rolebinding contains external employees who all have both the cloud main exception and
// the SAS exception --> profiles-state-controller should produce the labels
// state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: false and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: false

func TestExternalEmployeesWithBothExceptionsReturnsFalseAndFalse(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_5"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == False AND cloudMainResult == False, we are OK. Otherwise the test failed
	if !sasResult && !cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: false and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: false")
}

// If rolebinding contains only external employees who have no exceptions --> profiles-state-controller
// should produce the labels state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true

func TestExternalEmployeesWithNoExceptionsReturnsTrueAndTrue(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_6"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == True AND cloudMainResult == True, we are OK. Otherwise the test failed
	if sasResult && cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true")
}

// If a rolebinding contains StatCan employees AND a single external employee with no
// exceptions --> profiles-state-controller should produce the labels
// state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and
// state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true

func TestStatCanEmployeeAndUserWithNoExceptionsReturnsTrueAndTrue(t *testing.T) {
	rolebinding, _ := getRolebindings(filepath.Join(TEST_DIRECTORY, "exception_7"))
	sasResult := mockController.existsNonSasUser(rolebinding)
	cloudMainResult := mockController.existsNonCloudMainUser(rolebinding)
	// if sasResult == True AND cloudMainResult == True, we are OK. Otherwise the test failed
	if sasResult && cloudMainResult {
		return
	}
	t.Fatalf("Expected state.aaw.statcan.gc.ca/exists-non-sas-notebook-user: true and state.aaw.statcan.gc.ca/exists-non-cloud-main-user: true")
}
