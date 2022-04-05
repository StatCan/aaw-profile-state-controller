muchk8s.wow

## Requirements: 
The following need to be installed
- kubectl
- helm
- kind installed

Highly recommand at **least** 8gb or RAM. 

## To start the env:
`make create`

This will create the deployment of 3 namespaces/
- Alice - protected-b
- Bob - unclassidied
- carla - no label of classification

`make jupyter-alice` or similar: creates the pod for jupyter-alice, which will be protected-b as defined above.

`make chromium`: allows you to open a kiali console with http://kiali.muchk8s.wow/kiali/
or check your working pods with http://kubeflow.muchk8s.wow/notebook/alice/jupyter/lab or similar urls.


## To delete the env:
`make delete`

## To test the controller 
### Direct way
`make controller-deploy` will create the daaasinternal namespace and deploy the controller. Once it is, any notebook created will have their authorization policy created.

### Alternatice way
`go build -o profile-state-controller .`
`./profile-state-controller -kubeconfig=istio-test-config`

Once those two lines are executed, and it is working, for any notebook you create or delete, the corresponding authorization policy should be created or deleted.

## Test the controller - other way
`make controller-deploy`
This command might need a bit of love if you don't have enough cpu. It will need to remove the resources from the generated yaml. Another possibility is needing to remove `imagePullSecrets` also from the generated yaml. Then apply said generated yaml.

## Known Issues:
- If the deployments don't start because of insufficient CPUs.
`kubectl edit deployment gatekeeper-controller-manager -n gatekeeper-system`
Go to the line about resources, delete the lines and have `resources: {}`.  
  - This might also happen in the deploy.yaml

- Another possibilities, is that sometimes, the pods don't get the right classifications, then it needs to be deleted using something like `make jupyter-delete-alice` before re-creating it `make jupyter-alice` 

- Some times there will be an error in the make file, it happens if something isn't ready before the next step. 
The solution is to track down which step failed, and redo it manually, it and any following step.
