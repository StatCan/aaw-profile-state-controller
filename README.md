# Profile State Controller

The Profile State Controller applies labels to Kubeflow Profiles and Kubernetes namespaces that indicate certain properties of users in those namespaces.

The purpose of this controller is to ensure that profiles/namespaces are correctly labelled so that other applications in the cluster can make allow/deny decisions based on these labels.

The sections below detail specific labels and the logic they use.

The initial implementation involved the bundling of features into either employee or non-employee features, but the addition of exception lists introduced the requirement for more fine-grain discretionary access control. This means we have a label for each feature.
## SAS notebook feature

This controller adds the `state.aaw.statcan.gc.ca/has-sas-notebook-feature` and `state.aaw.statcan.gc.ca/exists-non-sas-notebook-user` labels to the Profiles based on the Pods and users in that namespace.
If any Pods in that namespace are using a SAS image, it will set the `state.aaw.statcan.gc.ca/has-sas-notebook-feature` label in the Profile to `true`, otherwise `false`.
If any users in that namespace are not permitted to use SAS (i.e external users and not in exception list), it will set the `state.aaw.statcan.gc.ca/exists-non-sas-notebook-user` label in the Profile to `true`, otherwise `false`.

### Interaction with Gatekeeper

The Gatekeeper Policies check for these labels in the Profile and allow objects to be created or denied accordingly. More information about these policies and how the objects interact can be found in their README in the Gatekeeper Policies repository (linked below).
- [Deny External Users Policy](https://github.com/StatCan/gatekeeper-policies/tree/master/general/deny-external-users)
  - `state.aaw.statcan.gc.ca/has-sas-notebook-feature` affects RoleBinding and AuthorizationPolicy objects - which allows or denies adding external contributors
  - Checks to see if there are any pods with the SAS image running. If there are, then it will only allow the RoleBinding and AuthorizationPolicy to be created for internal users.
- [Employee-Only Features Policy](https://github.com/StatCan/gatekeeper-policies/tree/master/pod-security-policy/deny-employee-only-features)
  - `state.aaw.statcan.gc.ca/exists-non-sas-notebook-user` affects Pod and Notebook objects - which allows or denies creation of SAS Notebook Servers
  - Checks to see if there are any external users in a namespace through a Profile label. If there are and they aren't in the [exception list](https://github.com/StatCan/aaw-kubeflow-profiles/blob/main/non-employee-exceptions-config.jsonnet), then it will not allow the SAS Pod and Notebook to be created.

### Unit Test Cases

1. If **any pod** in a list of pods contains a SAS image, `hasSasNotebookFeature` should return `true`.
2. If **no pod** in a list of pods contains a SAS image,  `hasSasNotebookFeature` should return `false`.
3. If an empty list is passed to `hasSasNotebookFeature`, it should return `false`.

4. If **any rolebinding** in a list of rolebindings contains a non-employee user, `existsNonSasUser` should return `true`.
5. If **no rolebinding** in a list of rolebindings contains a non-employee user, `existsNonSasUser` should return `false`.
6. If an empty list is passed to `existsNonSasUser`, it should return `false`.

## Non-Employee Users

This controller also adds the `state.aaw.statcan.gc.ca/non-employee-users` label to the Profiles based on the RoleBindings in that namespace.

When the owner of a Kubeflow Profile adds a contributor to their namespace, a rolebinding is [created in that namespace](https://www.kubeflow.org/docs/components/multi-tenancy/getting-started/#managing-contributors-manually), binding the new contributor to the `kubeflow-edit` role in that namespace.

Specifically, contributors are named by their email address (emails with a StatCan domain are considered employees and all other domains are considered non-employees).

If at any point a contributor with an external email address is added to a Kubeflow Profile, the Profile State Controller sets the label `state.aaw.statcan.gc.ca/non-employee-users=true` on both the Kubeflow Profile and the corresponding namespace.

If the RoleBinding shows that the user is external (email does not end with an accepted domain), it will set the label in the Profile to `true` otherwise `false`.

Other applications on the cluster can use this label to make decisions based on whether namespaces contain non-employee users.

## FDI Storage Feature
The FDI storage can be either an external or internal storage container. The controller uses the naming convention of the PVC to determine whether or not the containers are internal. If they are, it will set the label `state.aaw.statcan.gc.ca/exists-internal-blob-storage` to true. This label is used by the KFAM module in Kubeflow and prevents the addition of external users to the profile. 

### Unit Test Cases
See https://github.com/StatCan/aaw-profile-state-controller/blob/dd4f2944013f6d39709f80d0354fb8c09fddd131/pkg/controller/handler_test.go#L154 for the unit tests and relavent documentation.

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

**Try to add an external users when you have an internal DAS common storage container mounted to your namespace**
- check that the internal storage is mounted
- If you try to add a contributor with an email domain not in `statcan.gc.ca` or `cloud.statcan.ca`, you will receive an error.


### How to Contribute

See [CONTRIBUTING.md](CONTRIBUTING.md)

### License

Unless otherwise noted, the source code of this project is covered under Crown Copyright, Government of Canada, and is distributed under the [MIT License](LICENSE).

The Canada wordmark and related graphics associated with this distribution are protected under trademark law and copyright law.
No permission is granted to use them outside the parameters of the Government of Canada's corporate identity program.
For more information, see [Federal identity requirements](https://www.canada.ca/en/treasury-board-secretariat/topics/government-communications/federal-identity-requirements.html).
