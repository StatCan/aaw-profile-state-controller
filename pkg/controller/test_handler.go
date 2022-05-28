/*
These tests make use of hasEmployeeOnlyFeatures and isNonEmployeeUser with mocked rolebindings and pods.

Can parse Pod/RoleBinding specs from YAML files in tests/ folder and use the same YAML files to do an E2E
test against a local k8s cluster like k3d.
*/

package controller
