IMAGE := profile-state-controller
TAG := 0.0.1

.PHONY: build run

build:
	docker build . -t $(IMAGE):$(TAG)

run: build
	docker run $(IMAGE):$(TAG)
