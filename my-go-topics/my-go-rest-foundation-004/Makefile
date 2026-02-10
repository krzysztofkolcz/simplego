# Upgrade
VERSION=0.1.0
IMAGE=kkolcz/my-go-rest-foundation-004
TAG=$(IMAGE):$(VERSION)
RELEASE=my-go-rest-foundation-004-release

CHART=my-go-rest-foundation-004-chart
CHART_SHORT=my-go-rest-foundation-004-cha
CHART_DIR=./$(CHART)

HOST=localhost
REGISTRY_NAME=go-hello-registry
REGISTRY_PORT=5000

HOST_REGISTRY=$(HOST):$(REGISTRY_PORT)
CLUSTER_REGISTRY=$(REGISTRY_NAME):$(REGISTRY_PORT)

NAMESPACE=myrf004ns

build-image:
	docker build -t $(TAG) .

tag-image:
	docker tag $(TAG) $(HOST):$(REGISTRY_PORT)/$(TAG)
	docker tag $(TAG) $(REGISTRY_NAME):$(REGISTRY_PORT)/$(TAG)

push-image:
	@echo "Importing docker image into k3d"
	docker push localhost:$(REGISTRY_PORT)/$(TAG)


create-helm-charts:
	helm create $(CHART)

k3d-upgrade-helm: build-image tag-image push-image
	helm upgrade --install $(RELEASE) $(CHART_DIR) \
	  -n $(NAMESPACE) \
	  -f $(CHART_DIR)/values.yaml \
	  --create-namespace

image-list:
	curl http://localhost:5000/v2/_catalog

tag-list:
	curl http://localhost:5000/v2/kkolcz/my-go-rest-foundation-004/tags/list

helm-template:
	helm template $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml

port-forward:
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE)-$(CHART_SHORT) 8080:8080