KUBECTL_CONFIG=${HOME}/.config/k3d/kubeconfig-$(CLUSTER_NAME).yaml
NAMESPACE=go-hello-ns
CLUSTER_NAME=go-hello-cluster
REGISTRY_NAME=go-hello-registry
REGISTRY_PORT=5000

####
# K3D Cluster
####
default: 
	@echo "ok"

create-k3d-registry:
	k3d registry create go-hello-registry --port 5000

delete-k3d-registry:
	k3d registry delete go-hello-registry

# 0
create-k3d-cluster-with-registry:
	k3d cluster create $(CLUSTER_NAME) -p "30083:30083@server:0" --api-port 127.0.0.1:6445 --registry-create $(REGISTRY_NAME):$(REGISTRY_PORT); \
	k3d kubeconfig merge $(CLUSTER_NAME) --kubeconfig-switch-context; \
	kubectl create namespace $(NAMESPACE)

delete-cluster:
	@echo "Deleting k3d cluster '$(CLUSTER_NAME)'."
	@if k3d cluster list | grep -q '$(CLUSTER_NAME)'; then \
	   k3d cluster delete $(CLUSTER_NAME); \
	else \
	   echo "k3d cluster '$(CLUSTER_NAME)' does not exist."; \
	fi

clean-k3d:
	@echo "Cleaning everything in the namespace in k3d."
	@if kubectl --kubeconfig=${KUBECTL_CONFIG} get namespace $(NAMESPACE) > /dev/null 2>&1; then \
	   kubectl --kubeconfig=${KUBECTL_CONFIG} delete namespace $(NAMESPACE) --ignore-not-found=true; \
	else \
	   echo "Namespace $(NAMESPACE) does not exist."; \
	fi

####
# Postgresql
####


wait-for-pod:
	@echo "Waiting for pod with label $(LABEL) in namespace $(NAMESPACE) to be Running..."
	@while [ -z "$$(kubectl get pod -n $(NAMESPACE) -l $(LABEL) -o jsonpath='{.items[*].metadata.name}')" ]; do \
		echo "No pods found, waiting for pod creation..."; \
		sleep 2; \
	done
	@while [ "$$(kubectl get pod -n $(NAMESPACE) -l $(LABEL) -o jsonpath='{.items[0].status.phase}' 2>/dev/null)" != "Running" ]; do \
		echo "Pod not ready, waiting..."; \
		sleep 2; \
	done
	@echo "Pod is Running!"


####
# Helm install
####
# Start 
TAG=kkolcz/go-hello:0.1.0
RELEASE=demo
CHART_DIR=./go-hello
# 1
build-image:
	docker build -t $(TAG) .

# 2
tag-image:
	docker tag $(TAG) localhost:$(REGISTRY_PORT)/$(TAG)
	docker tag $(TAG) $(REGISTRY_NAME):$(REGISTRY_PORT)/$(TAG)


# Target to build Docker imgage within k3d 
# 3
k3d-import-image:
	@echo "Importing docker image into k3d"
	docker push localhost:$(REGISTRY_PORT)/$(TAG)

cluster-list:
	k3d cluster list

image-list:
	curl http://localhost:5000/v2/_catalog

tag-list:
	curl http://localhost:5000/v2/kkolcz/go-hello/tags/list

# 3
k3d-install-helm:
	helm install $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml -n $(NAMESPACE)

k3d-dry-run:
	helm install --dry-run --debug $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml

# k3d-build-helm:
# 	helm dependency build ./charts

VERSION=0.1.1
IMAGE=kkolcz/go-hello
HOST_REGISTRY=localhost:5000
# Musze tak otagować - registry z hosta dostępne pod localhost
CLUSTER_REGISTRY=go-hello-registry:5000
# Musze tak otagować - registry z klastra dostępne pod tym tagiem. Ustawione w values.yaml

k3d-upgrade-helm:
	docker build -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(HOST_REGISTRY)/$(IMAGE):$(VERSION)
	docker tag $(IMAGE):$(VERSION) $(CLUSTER_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(HOST_REGISTRY)/$(IMAGE):$(VERSION)
	helm upgrade --install $(RELEASE) ./go-hello \
	  -n go-hello-ns \
	  -f ./go-hello/values.yaml

port-forward:
	kubectl port-forward -n $(NAMESPACE) svc/demo-go-hello 8080:80



####
# psql-2
####
psql2-add-bitnami:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update

psql2-install-psql:
	helm install postgresql bitnami/postgresql \
	-n $(NAMESPACE) \
	-f postgresql/values.yaml