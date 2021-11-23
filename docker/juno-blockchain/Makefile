
REGISTRY_PATH=ghcr.io/cosmoscontracts/juno-blockchain

TAG = latest

docker-deploy: docker-build docker-push

docker-pull:
	docker pull $(REGISTRY_PATH):$(TAG) || true

docker-build:
	docker build \
		-t $(REGISTRY_PATH):$(TAG) \
		-f ./Dockerfile \
		../..

docker-push:
	docker push $(REGISTRY_PATH):$(TAG)

