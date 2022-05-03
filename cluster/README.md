## Profile State Controller

Checks images in Pods and email in RoleBindings to set `state.aaw.statcan.gc.ca/employee-only-features` and `state.aaw.statcan.gc.ca/non-employee-user` labels in the Profile.

### Test Cases (in examples folder)

alice:
- Pod has a SAS image 
- `state.aaw.statcan.gc.ca/employee-only-features` will be set to `true` in Profile
- RoleBinding shows internal user (email ends in accepted domain). 
- `state.aaw.statcan.gc.ca/non-employee-user` will be set to `false` in Profile

bob:
- Pods have no SAS image
- `state.aaw.statcan.gc.ca/employee-only-features` will be set to `false` in Profile
- RoleBinding shows user is external (email ends in non-accepted domains)
- `state.aaw.statcan.gc.ca/non-employee-user` will be set to `true` in Profile

sam:
- Pod has SAS image
- `state.aaw.statcan.gc.ca/employee-only-features` will be set to `true` in Profile
- Two RoleBindings in that namespace. One shows internal user (email ends in accepted domain) and another is external.
- `state.aaw.statcan.gc.ca/non-employee-user` will be set to `true` in Profile
