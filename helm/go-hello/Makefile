KUBECTL_CONFIG=${HOME}/.config/k3d/kubeconfig-$(CLUSTER_NAME).yaml
NAMESPACE=go-hello-ns
CLUSTER_NAME=go-hello-cluster
REGISTRY_NAME=go-hello-registry
REGISTRY_PORT=5000

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


DBUSERNAME := postgres
DBPASS := secret
DBNAME := simplego
DB_ADMIN_PASS_KEY := secret
PSQL_RELEASE_NAME := simplegodb

psql-add-to-cluster:
# 	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
# 	kubectl apply -f helm/go-hello/go-hello/charts/configmap.yaml -n $(NAMESPACE)
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
	docker tag $(TAG) k3d-$(REGISTRY_NAME):$(REGISTRY_PORT)/$(TAG)


# Target to build Docker imgage within k3d 
# 3
k3d-import-image: build-image
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
	helm install $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml

k3d-dry-run:
	helm install --dry-run --debug $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml


# k3d-apply-helm-chart:
# 	@echo "Applying Helm chart."
# 	helm upgrade --install $(CHART_NAME) $(CHART_DIR) --namespace $(APPLY_NAMESPACE) \
# 		--set volumePath=$(PATH_TO_INIT_VOLUME) \
# 		--set config.activePlugins=$(ACTIVE_PLUGINS) \
# 		-f $(CHART_DIR)/values-dev.yaml

# k3d-apply-cmk-helm-chart:
# 	@echo "Applying CMK Helm chart."
# 	$(MAKE) k3d-apply-helm-chart CHART_NAME=cmk CHART_DIR=$(pwd)/charts APPLY_NAMESPACE=$(NAMESPACE)

# k3d-build-helm:
# 	helm dependency build ./charts