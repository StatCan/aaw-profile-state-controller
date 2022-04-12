# NOTE: You will need kubectl, helm, and kind installed.

REGISTRY := internaluser
CONTROLLER := controller
TAG := latest

CHROMIUM := chromium
#CHROMIUM := flatpak run org.chromium.Chromium

.DEFAULT: create

create: create-users configure-users
	@echo "Cluster is all set up for users!"

go-deploy:
	cd .. && go build .
	cd .. && ./profile-state-controller --kubeconfig ~/.kube/config

controller:
	cd .. && docker build . -t $(REGISTRY)/$(CONTROLLER):$(TAG)
	k3d image import --cluster sas-policy-cluster $(REGISTRY)/$(CONTROLLER):$(TAG)

###  _   _
### | | | |
### | | | |___  ___ _ __ ___
### | | | / __|/ _ \ '__/ __|
### | |_| \__ \  __/ |  \__ \
###  \___/|___/\___|_|  |___/
###
create-namespaces: alice bob
alice alice2 bob:
	kubectl create namespace $@ || true

create-pods:
	kubectl apply -f ./examples/alice-pod-sas.yaml || true
	kubectl apply -f ./examples/alice-pod.yaml || true
	kubectl apply -f ./examples/bob-pod.yaml || true

create-profiles:
	kubectl apply -f ./examples/alice-profile.yaml || true
	kubectl apply -f ./examples/bob-profile.yaml || true

delete-pods:
	kubectl delete -f ./examples/alice-pod-sas.yaml || true
	kubectl delete -f ./examples/alice-pod.yaml || true
	kubectl delete -f ./examples/bob-pod.yaml || true

