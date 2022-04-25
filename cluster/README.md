## Profile State Controller

Checks images in Pods and email in RoleBindings to set `state.aaw.statcan.gc.ca/employee-only-features` and `state.aaw.statcan.gc.ca/non-employee-user` labels in the Profile.

### Test Cases (in examples folder)

alice:
- Pod has a SAS image 
- `state.aaw.statcan.gc.ca/employee-only-features` will be set to `true` in Profile
- Based on previous [Deny External Users](https://github.com/StatCan/gatekeeper-policies/blob/master/general/deny-external-users/README.md) policy, will not be allowed to create a RoleBinding. No RoleBindings exist in namespace. 
- `state.aaw.statcan.gc.ca/non-employee-user` will be set to `true` in Profile by default

bob:
- Pods have no SAS image
- `state.aaw.statcan.gc.ca/employee-only-features` will be set to `false` in Profile
- RoleBinding shows user is external (email ends in non-accepted domains)
- `state.aaw.statcan.gc.ca/non-employee-user` will be set to `true` in Profile

sam:
- Pod has SAS image
- `state.aaw.statcan.gc.ca/employee-only-features` will be set to `true` in Profile
- RoleBinding shows internal user (email ends in accepted domain)
- `state.aaw.statcan.gc.ca/non-employee-user` will be set to `false` in Profile
