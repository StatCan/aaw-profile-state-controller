# Profile State Controller

The Profile State Controller applies labels to Kubeflow Profiles and Kubernetes namespaces that indicate certain properties of users in those namespaces.

The purpose of this controller is to ensure that profiles/namespaces are correctly labelled so that other applications in the cluster can make allow/deny decisions based on these labels.

The sections below detail specific labels and the logic they use.

## Employee Only Features

This controller adds the `state.aaw.statcan.gc.ca/employee-only-features` label to the Profiles based on the Pods in that namespace.
If any Pods in that namespace are using a SAS image, it will set the label in the Profile to `true`, otherwise `false`.

### Unit Test Cases

1. If **any pod** in a list of pods contains a SAS image, `hasEmployeeOnlyFeatures` should return `true`.
2. If **no pod** in a list of pods contains a SAS image,  `hasEmployeeOnlyFeatures` should return `false`.
3. If an empty list is passed to `hasEmployeeOnlyFeatures`, it should return `false`.

## Non-Employee Users

When the owner of a Kubeflow Profile adds a contributor to their namespace, a rolebinding is [created in that namespace](https://www.kubeflow.org/docs/components/multi-tenancy/getting-started/#managing-contributors-manually), binding the new contributor to the `kubeflow-edit` role in that namespace.

Specifically, contributors are named by their email address (emails with a StatCan domain are considered employees and all other domains are considered non-employees).

If at any point a contributor with an external email address is added to a Kubeflow Profile, the Profile State Controller sets the label `state.aaw.statcan.gc.ca/non-employee-users=true` on both the Kubeflow Profile and the corresponding namespace.

Other applications on the cluster can use this label to make decisions based on whether namespaces contain non-employee users.

### Unit Test Cases

1. If **any rolebinding** in a list of rolebindings contains a non-employee user, `hasNonEmployeeUser` should return `true`.
2. If **no rolebinding** in a list of rolebindings contains a non-employee user, `hasNonEmployeeUser` should return `false`.
3. If an empty list is passed to `hasNonEmployeeUser`, it should return `false`.


### How to Contribute

See [CONTRIBUTING.md](CONTRIBUTING.md)

### License

Unless otherwise noted, the source code of this project is covered under Crown Copyright, Government of Canada, and is distributed under the [MIT License](LICENSE).

The Canada wordmark and related graphics associated with this distribution are protected under trademark law and copyright law.
No permission is granted to use them outside the parameters of the Government of Canada's corporate identity program.
For more information, see [Federal identity requirements](https://www.canada.ca/en/treasury-board-secretariat/topics/government-communications/federal-identity-requirements.html).
