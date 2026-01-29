
VERSION=0.1.0
IMAGE=kkolcz/my-go-rest-foundation
TAG=$(IMAGE):$(VERSION)
RELEASE=my-go-rest-foundation-release

CHART=my-go-rest-foundation-chart
CHART_DIR=./$(CHART)

HOST=localhost
REGISTRY_NAME=go-hello-registry
REGISTRY_PORT=5000

HOST_REGISTRY=$(HOST):$(REGISTRY_PORT)
CLUSTER_REGISTRY=$(REGISTRY_NAME):$(REGISTRY_PORT)

NAMESPACE=myrfns

build-image:
	docker build -t $(TAG) .

tag-image:
	docker tag $(TAG) localhost:$(REGISTRY_PORT)/$(TAG)
	docker tag $(TAG) $(REGISTRY_NAME):$(REGISTRY_PORT)/$(TAG)

k3d-import-image:
	@echo "Importing docker image into k3d"
	docker push localhost:$(REGISTRY_PORT)/$(TAG)

create-helm-charts:
	helm create $(CHART)

# upgrade --install - można zamiast helm install
# Zmieniasz chart → robisz helm upgrade --install → K8s robi resztę.
k3d-upgrade-helm:
	docker build -t $(TAG) .
	docker tag $(TAG) $(HOST_REGISTRY)/$(TAG)
	docker tag $(TAG) $(CLUSTER_REGISTRY)/$(TAG)
	docker push $(HOST_REGISTRY)/$(TAG)
	helm upgrade --install $(RELEASE) $(CHART_DIR) \
	  -n $(NAMESPACE) \
	  -f $(CHART_DIR)/values.yaml \
	  --create-namespace

# kubectl get services -n myrfns
# NAME                                                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
# my-go-rest-foundation-release-my-go-rest-foundation-chart   ClusterIP   10.43.12.83   <none>        80/TCP    75m

port-forward:
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE)-$(CHART) 8080:80

helm-template:
	helm template $(RELEASE) $(CHART_DIR)

helm-diff:
	helm diff upgrade $(RELEASE) $(CHART_DIR) -n $(NAMESPACE)



