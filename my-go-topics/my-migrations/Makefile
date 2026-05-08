DB_URL=postgres://mymigrationsuser:pass@localhost:5432/db?sslmode=disable

migrate-public:
	migrate -path db/migrations/public -database $(DB_URL) up

migrate-tenant:
	migrate -path db/migrations/tenant -database "$(DB_URL)&search_path=$(SCHEMA)" up

integration-test:
	go test -tags=integration ./tests/integration/...

# go test -v ./tests/integration/...        # verbose (print test names)
# go test -count=1 ./tests/integration/...  # bez cache (ważne przy integracyjnych!)
# go test -run TestUser ./tests/integration/...  # tylko wybrane testy



image-list:
	curl http://localhost:5000/v2/_catalog

tag-list:
	curl http://localhost:5000/v2/kkolcz/go-hello/tags/list


REGISTRY_NAME=go-hello-registry
REGISTRY_PORT=5000


####
# Helm install
####
# Start 
VERSION=0.0.34
IMAGE=kkolcz/go-migrations
HOST_REGISTRY=localhost:5000
CLUSTER_NAME=go-hello-cluster
RELEASE=rel001
DIR=charts
CHART_DIR=./$(DIR)
NAMESPACE=go-migrations-ns
CLUSTER_REGISTRY=go-hello-registry:5000

create-helm-charts:
	helm create $(DIR)

ns:
	kubectl create namespace $(NAMESPACE)

k3d-image:
	docker build -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(HOST_REGISTRY)/$(IMAGE):$(VERSION)
	docker tag $(IMAGE):$(VERSION) $(CLUSTER_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(HOST_REGISTRY)/$(IMAGE):$(VERSION)

k3d-helm-up:
	helm upgrade --install $(RELEASE) $(CHART_DIR) \
	  -n $(NAMESPACE) \
	  -f $(CHART_DIR)/values.yaml \
	  --wait --timeout 120s --debug

k3d-image-helm-up: k3d-image k3d-helm-up
	echo "ok"

k3d-template:
	helm template $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml	

k3d-dry-run:
	helm install $(RELEASE) $(CHART_DIR) \
	-f $(CHART_DIR)/values.yaml \
	--dry-run \
	--debug

k3d-retry-failed:
	helm upgrade $(RELEASE) $(CHART_DIR) --set migrateRetry.enabled=true

k-retry-failed:
	kubectl apply -f charts/templates/migrate-retry-failed-job.yaml

# obserwuj logi na bieżąco
k-retry-failed-logs:
	kubectl logs -f job/migrate-retry-failed-job -n $(NAMESPACE)

# po zakończeniu — usuń żeby następnym razem móc apply znowu
k-retry-failed-logs:
	kubectl delete -f job/migrate-retry-failed-job -n $(NAMESPACE)

pf:
	kubectl port-forward svc/rel001-charts 8080:80
	# kubectl port-forward svc/rel001-go-migrations 8080:80

psql-pf:
	kubectl port-forward svc/postgresql 5434:5432 -n database


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

####
# Sealed secret
####
sealed-create-ns:
	kubectl create namespace sealed-secrets

sealed-add-repo:
	helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
	helm repo update

sealed-install:
	helm install sealed-secrets sealed-secrets/sealed-secrets \
	-n sealed-secrets

sealed-test:
	kubectl get pods -n sealed-secrets

sealed-create-ns-database:
	kubectl create namespace database

# Tworzy plik ./go-hello/templates/sealedsecret-postgres.yaml
# musi istniec ./postgresql/postgres-secret.yaml - w nim definiuje namespace 
sealed-generate:
	kubeseal \
	--controller-name sealed-secrets \
	--controller-namespace sealed-secrets \
	--format yaml \
	< ./charts/postgres/postgres-secret.yaml \
	> ./charts/templates/sealedsecret-postgres.yaml

test-single: docker-host-mackbook
	go test ./tests/integration -run TestPublicMigrations -v


sqlc-generate:
	sqlc generate

docker-host-mackbook:
	export DOCKER_HOST=unix:///Users/C5383717/.colima/default/docker.sock