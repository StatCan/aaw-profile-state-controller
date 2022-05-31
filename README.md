## Profile State Controller

This controller adds the `state.aaw.statcan.gc.ca/employee-only-features` label to the Profiles based on the Pods in that namespace. 
If any Pods in that namespace are using a SAS image, it will set the label in the Profile to `true`, otherwise `false`.

This controller also adds the `state.aaw.statcan.gc.ca/non-employee-users` label to the Profiles based on the RoleBindings in that namespace.
If the RoleBinding shows that the user is external (email does not end with an accepted domain), it will set the label in the Profile to `true` otherwise `false`.

The Gatekeeper Policies check for these labels in the Profile and allow objects to be created or denied accordingly. More information about these policies and how the objects interact can be found in their README in the Gatekeeper Policies repository (linked below).
- [Deny External Users Policy](https://github.com/StatCan/gatekeeper-policies/tree/master/general/deny-external-users)
  - `state.aaw.statcan.gc.ca/employee-only-features` affects RoleBinding and AuthorizationPolicy objects - which allows or denies adding external contributors
  - Checks to see if there are any employee only features (like SAS images) in the namespace through a profile label. If there are, then it will only allow the RoleBinding and AuthorizationPolicy to be created for internal users.
- [Employee-Only Features Policy](https://github.com/StatCan/gatekeeper-policies/tree/master/pod-security-policy/deny-employee-only-features) 
  - `state.aaw.statcan.gc.ca/non-employee-users` affects Pod and Notebook objects - which allows or denies creation of SAS Notebook Servers
  - Checks to see if there are any external users in a namespace through a Profile label. If there are, then it will not allow the SAS Pod and Notebook to be created.

### Implementation in Kubeflow
**Add an External Contributor with a SAS image on your namespace**
- Create a new notebook server with a custom SAS image
- If you go to Manage Contributors and try to add an external user then it will give you an error so you will not be allowed to add them as a contributor to your namespace.
![image](https://user-images.githubusercontent.com/57377830/171225642-4291de80-9e33-4f3f-9192-d0d608269198.png)
*If you want to add the external contributor, you will have to delete the notebook server with the SAS image first and then add retry adding them a contributor.

**Try to create a notebook with a SAS image when you have an external contributor on your namespace**
- Add external user as a contributor (make sure you have no SAS images to do this).
![image](https://user-images.githubusercontent.com/57377830/171225762-69ec2ac1-425b-4d9f-975b-dc6bc251833d.png)
- If you try to create a new notebook server with a custom SAS image (k8scc01covidacr.azurecr.io/sas:latest), you will get the following error, stopping you from creating it:
![image](https://user-images.githubusercontent.com/57377830/171225844-5f1e9da6-f304-4c82-8d26-81f0be0aa577.png)
![image](https://user-images.githubusercontent.com/57377830/171225864-cfd8fa61-b8a8-41ed-a5f8-3d316d7bf35c.png)


*If you want to create the SAS notebook, you will have to remove the external contributor from your namespace first and then retry creating the notebook.

**Sometimes it will take a few seconds in between removing/adding a contributor and notebook for the policies to take effect



### How to Contribute

See [CONTRIBUTING.md](CONTRIBUTING.md)

### License

Unless otherwise noted, the source code of this project is covered under Crown Copyright, Government of Canada, and is distributed under the [MIT License](LICENSE).

The Canada wordmark and related graphics associated with this distribution are protected under trademark law and copyright law. 
No permission is granted to use them outside the parameters of the Government of Canada's corporate identity program. 
For more information, see [Federal identity requirements](https://www.canada.ca/en/treasury-board-secretariat/topics/government-communications/federal-identity-requirements.html).
