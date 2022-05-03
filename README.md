## Profile State Controller

This controller adds the `state.aaw.statcan.gc.ca/employee-only-features` label to the Profiles based on the Pods in that namespace. 
If any Pods in that namespace are using a SAS image, it will set the label in the Profile to `true`, otherwise `false`.

This controller also adds the `state.aaw.statcan.gc.ca/non-employee-user` label to the Profiles based on the RoleBindings in that namespace.
If the RoleBinding shows that the user is external (email does not end with an accepted domain), it will set the label in the Profile to `true` otherwise `false`.

### How to Contribute

See [CONTRIBUTING.md](CONTRIBUTING.md)

### License

Unless otherwise noted, the source code of this project is covered under Crown Copyright, Government of Canada, and is distributed under the [MIT License](LICENSE).

The Canada wordmark and related graphics associated with this distribution are protected under trademark law and copyright law. 
No permission is granted to use them outside the parameters of the Government of Canada's corporate identity program. 
For more information, see [Federal identity requirements](https://www.canada.ca/en/treasury-board-secretariat/topics/government-communications/federal-identity-requirements.html).
