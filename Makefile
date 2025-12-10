SIMPLEGO_APP_NAME := simplego-api-server
SIMPLEGO_DEV_TARGET := dev
TAG := latest
IMAGE_NAME := $(SIMPLEGO_APP_NAME)-$(SIMPLEGO_DEV_TARGET):$(TAG)
DOCKERFILE_DIR := .
DOCKERFILE_NAME := Dockerfile
CONTEXT_DIR := .
SIMPLEGO_HELM_CHART := /Users/C5383717/GolandProjects/helm-tutorial/simpleGo-helm

# Target to build Docker image
docker-dev-build:
	docker build -f $(DOCKERFILE_DIR)/$(DOCKERFILE_NAME) -t $(IMAGE_NAME) $(CONTEXT_DIR)

docker-dev-run:
	docker run -p 8080:8080 --env-file ./.env $(IMAGE_NAME)
##K3D

.PHONY: install-k3d start-k3d k3d-build-image k3d-build-otel-image k3d-build-audit-server-image \
k3d-build-producer-% k3d-deploy-% clean-k3d delete-cluster

KUBECTL_CONFIG=${HOME}/.config/k3d/kubeconfig-$(CLUSTER_NAME).yaml
CLUSTER_NAME=simplegocluster
NAMESPACE=simplego
PATH_TO_INIT_VOLUME=$(pwd)/local_env/volume

# Target to install k3d using wget
install-k3d:
	@echo "Installing k3d using wget."
	@if ! command -v k3d >/dev/null 2>&1; then \
		wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash; \
	else \
		echo "k3d is already installed."; \
	fi

start-k3d-colima: install-k3d delete-cluster
	@echo "Starting k3d."
	@if ! k3d cluster list | grep -q '$(CLUSTER_NAME)'; then \
	   K3D_FIX_DNS=0 k3d cluster create $(CLUSTER_NAME) -p "30084:30083@server:0" --api-port 127.0.0.1:6444; \
	   k3d kubeconfig merge $(CLUSTER_NAME) --kubeconfig-switch-context; \
	fi

# Target to start k3d
start-k3d: install-k3d delete-cluster
	@echo "Starting k3d."
	@if ! k3d cluster list | grep -q '$(CLUSTER_NAME)'; then \
	   k3d cluster create $(CLUSTER_NAME) -p "30084:30083@server:0" --api-port 127.0.0.1:6444; \
	   k3d kubeconfig merge $(CLUSTER_NAME) --kubeconfig-switch-context; \
	fi

# Target to build Docker image within k3d
k3d-build-image: docker-dev-build
	@echo "Importing Docker image into k3d."
	k3d image import $(APPLY_IMAGE_NAME) -c $(CLUSTER_NAME)

# Target to build the CMK image within k3d
k3d-build-simplego-image:
	@echo "Building the simplego image within k3d."
	@$(MAKE) k3d-build-image APPLY_IMAGE_NAME=$(IMAGE_NAME)

k3d-apply-helm-chart:
	@echo "Applying Helm chart."
	helm upgrade --install $(CHART_NAME) $(CHART_DIR) --namespace $(APPLY_NAMESPACE) --create-namespace --set volumePath=$(PATH_TO_INIT_VOLUME)

k3d-apply-simplego-helm-chart: clean-k3d start-k3d k3d-build-simplego-image
	@echo "Applying CMK Helm chart."
	$(MAKE) k3d-apply-helm-chart CHART_NAME=simplego CHART_DIR=$(SIMPLEGO_HELM_CHART) APPLY_NAMESPACE=$(NAMESPACE)

k3d-apply-simplego-helm-chart-colima: clean-k3d start-k3d-colima k3d-build-simplego-image
	@echo "Applying CMK Helm chart."
	$(MAKE) k3d-apply-helm-chart CHART_NAME=simplego CHART_DIR=$(SIMPLEGO_HELM_CHART) APPLY_NAMESPACE=$(NAMESPACE)

k3d-upgrade-simplego-helm-chart:
	@echo "Applying CMK Helm chart."
	$(MAKE) k3d-apply-helm-chart CHART_NAME=simplego CHART_DIR=$(SIMPLEGO_HELM_CHART) APPLY_NAMESPACE=$(NAMESPACE)

# Target to clean everything in the namespace
clean-k3d:
	@echo "Cleaning everything in the simplego namespace in k3d."
	@if kubectl --kubeconfig=${KUBECTL_CONFIG} get namespace $(NAMESPACE) > /dev/null 2>&1; then \
	   kubectl --kubeconfig=${KUBECTL_CONFIG} delete namespace $(NAMESPACE) --ignore-not-found=true; \
	else \
	   echo "Namespace $(NAMESPACE) does not exist."; \
	fi

# Target to delete the k3d cluster
delete-cluster:
	@echo "Deleting k3d cluster '$(CLUSTER_NAME)'."
	@if k3d cluster list | grep -q '$(CLUSTER_NAME)'; then \
	   k3d cluster delete $(CLUSTER_NAME); \
	else \
	   echo "k3d cluster '$(CLUSTER_NAME)' does not exist."; \
	fi

# Kolejność
# make k3d-apply-simplego-helm-chart-colima
# clean-k3d
# start-k3d-colima
	# install-k3d
	# delete-cluster
# k3d-build-simplego-image
	# k3d-build-image:
		# docker-dev-build

# make clean-k3d
# make start-k3d-colima
# make k3d-build-simplego-image
# k create namespace simplego
# make add-postgresql-to-cluster-2
# make k3d-upgrade-simplego-helm-chart


GOLANGCI_VERSION = v2.5.0
GOLANGCI_LINT := $(shell go env GOPATH)/bin/golangci-lint

lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

lint:
	$(GOLANGCI_LINT) run -v --fix

gotestsum:
	go install gotest.tools/gotestsum@latest
	export PATH=$PATH:$HOME/go/bin

DBUSERNAME := postgres
DBPASS := secret
DBNAME := simplego
DB_ADMIN_PASS_KEY := secret
PSQL_RELEASE_NAME := simplegodb

psql-add-to-cluster:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f helm/go-hello/go-hello/charts/configmap.yaml -n $(NAMESPACE)
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	helm upgrade --install $(PSQL_RELEASE_NAME) bitnami/postgresql \
	  --set global.postgresql.auth.username=$(DBUSERNAME) \
	  --set global.postgresql.auth.password=$(DBPASS) \
	  --set global.postgresql.auth.database=$(DBNAME) \
	  --set global.postgresql.auth.secretKeys.adminPasswordKey=$(DB_ADMIN_PASS_KEY) \
	  --set primary.existingConfigmap=postgres-custom-config \
	  --namespace $(NAMESPACE) \
	  --create-namespace
# 	  --set image.repository=bitnamilegacy/postgresql \

create-simplego-db: psql-port-forward wait-for-psql
	PGPASSWORD=$(DBPASS) psql -h localhost -p 5432 -U $(DBUSERNAME) -f ./db/db.sql

psql-port-forward: wait-for-psql
	@if ! lsof -i :5432 >/dev/null; then \
		echo "Start 5432 port-forward"; \
		kubectl port-forward svc/$(PSQL_RELEASE_NAME) 5432:5432 -n $(NAMESPACE) & \
	else \
		echo "Port 5432 already forwarded"; \
	fi

psql-secret:
	kubectl create secret generic '$(PSQL_RELEASE_NAME)' --from-literal=password='$(DBPASS)' --namespace '$(NAMESPACE)'

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

# wait-for-psql:
# 	@$(MAKE) wait-for-pod LABEL=app.kubernetes.io/name=postgresql

wait-for-psql:
	@kubectl wait \
		--for=condition=Ready pod \
		-l app.kubernetes.io/instance=$(PSQL_RELEASE_NAME) \
		-n $(NAMESPACE) \
		--timeout=120s

psql-cli: psql-port-forward
	@PGPASSWORD="$(DBPASS)" \
	psql -h 127.0.0.1 -p 5432 -U postgres -d $(DBNAME)