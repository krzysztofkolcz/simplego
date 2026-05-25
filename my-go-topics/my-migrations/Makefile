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
VERSION=0.0.44
IMAGE=kkolcz/go-migrations
HOST_REGISTRY=localhost:5000
CLUSTER_NAME=go-hello-cluster
RELEASE=rel001
DIR=charts
CHART_DIR=./$(DIR)
NAMESPACE=go-migrations-ns
CLUSTER_REGISTRY=go-hello-registry:5000
ENV ?= local

create-helm-charts:
	helm create $(DIR)

ns:
	kubectl create namespace $(NAMESPACE)

# Zbudowanie nowej wersji image z aplikacją i dodanie jej do rejestru
k3d-image:
	docker build -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(HOST_REGISTRY)/$(IMAGE):$(VERSION)
	docker tag $(IMAGE):$(VERSION) $(CLUSTER_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(HOST_REGISTRY)/$(IMAGE):$(VERSION)

# Release wersji na klaster
# Użycie: make k3d-helm-up           → values.local.yaml
#         make k3d-helm-up ENV=dev   → values.dev.yaml
#         make k3d-helm-up ENV=prod  → values.prod.yaml
k3d-helm-up:
	helm upgrade --install $(RELEASE) $(CHART_DIR) \
	  -n $(NAMESPACE) \
	  -f $(CHART_DIR)/values.yaml \
	  -f $(CHART_DIR)/values.$(ENV).yaml \
	  --set image.tag=$(VERSION) \
	  --wait --timeout 120s --debug

# Zbudowanie nowej wersji i release w jednym
k3d-image-helm-up: k3d-image k3d-helm-up
	echo "ok"

# Chyba wypisanie jak będą wyglądały pliki, po uruchomieniu.
# W celu przejrzenia przed releasem
# TODO - potwierdzić
k3d-template:
	helm template $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml	

k3d-template-2:
	helm template $(RELEASE) $(CHART_DIR) \
	-n go-migrations-ns \
	-f charts/values.yaml \
	-f charts/values.local.yaml \
	| grep -A 20 "kind: Deployment"

# j.w.
# TODO - potwierdzić
k3d-dry-run:
	helm install $(RELEASE) $(CHART_DIR) \
	-f $(CHART_DIR)/values.yaml \
	--dry-run \
	--debug

# Uruchomienie retry, dla tenantów, których migracja się nie udała
# TODO - potwierdzić
k3d-retry-failed:
	helm upgrade $(RELEASE) $(CHART_DIR) --set migrateRetry.enabled=true

# Uruchomienie retry, dla tenantów, których migracja się nie udała
# TODO - potwierdzić
k-retry-failed:
	kubectl apply -f charts/templates/migrate-retry-failed-job.yaml

# obserwuj logi na bieżąco
k-retry-failed-logs:
	kubectl logs -f job/migrate-retry-failed-job -n $(NAMESPACE)

# po zakończeniu — usuń żeby następnym razem móc apply znowu
k-delete-retry-job:
	kubectl delete -f job/migrate-retry-failed-job -n $(NAMESPACE)

pf:
	kubectl port-forward svc/rel001-charts 8080:80
	# kubectl port-forward svc/rel001-go-migrations 8080:80



####
# psql-2
####
DBNAMESPACE=database
psql2-add-bitnami:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update

psql2-install-psql:
	helm install postgresql bitnami/postgresql \
	-n $(DBNAMESPACE) \
	-f charts/postgres/values.yaml

####
# Restart postgresql psql database
####
psql-uninstall:
	helm uninstall postgresql -n $(DBNAMESPACE)

# after this run: make psql2-install-psql
psql-delete-pvc:
	kubectl delete pvc data-postgresql-0 -n $(DBNAMESPACE)

# Info from bitnami
# postgresql.database.svc.cluster.local - Read/Write connection
psql-postgres-pass:
	export POSTGRES_ADMIN_PASSWORD=$(kubectl get secret --namespace database postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)

psql-appuser-pass:
	export POSTGRES_PASSWORD=$(kubectl get secret --namespace database postgresql -o jsonpath="{.data.password}" | base64 -d)

psql-connect-1:
    kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace database --image registry-1.docker.io/bitnami/postgresql:17.5.0 --env="PGPASSWORD=$POSTGRES_PASSWORD" \
      --command -- psql --host postgresql -U appuser -d appdb -p 5432

psql-connect-2:
	kubectl port-forward --namespace database svc/postgresql 5434:5432 & PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U appuser -d appdb -p 5434

psql-connect-3:
	PGPASSWORD=strongpassword psql --host 127.0.0.1 -U appuser -d appdb -p 5434

psql-connect-4:
	psql "postgres://appuser:strongpassword@127.0.0.1:5434/appdb?sslmode=disable"

psql-pf:
	kubectl port-forward --address 0.0.0.0 svc/postgresql 5434:5432 -n database

psql-pf-kill:
	lsof -ti :5434 | xargs kill -9

psql-connect-superuser-1:
	kubectl exec -it postgresql-0 -n database -- psql -U postgres

# potrzebny make psql-pf
#   PGPASSWORD=$(kubectl get secret postgresql -n database -o jsonpath='{.data.postgres-password}' | base64 -d)
psql-connect-superuser-2:
	PGPASSWORD=$$(kubectl get secret postgresql -n database \
	  -o jsonpath='{.data.postgres-password}' | base64 -d) \
	psql --host 127.0.0.1 -U postgres -d postgres -p 5434

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

xxx:
	echo 'xxx'

codegen:
	oapi-codegen --config=./apis/config.yaml ./apis/mymigrations.yaml

# Postgres problems
# Dlaczego postgres nie wstaje
describe-psql:
	kubectl describe pod postgresql-0 -n database

# Wyczyszczenie obrazow dockera, jeżeli mało miejsca na dysku
docker-prune:
	docker system prune -a

# Wymuszenie restartu
psql-force:
	kubectl taint nodes --all node.kubernetes.io/disk-pressure:NoSchedule-