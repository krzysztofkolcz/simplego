```Makefile
DBUSERNAME      := simplegouser
DBPASS          := pass
DBNAME          := simplego
DB_ADMIN_PASS   := admin-pass
PSQL_RELEASE_NAME := simplegodb
PG_SECRET_NAME  := pg-auth

# psql-001 secret z haslami
psql-secret:
	@kubectl create secret generic $(PG_SECRET_NAME) \
	  --from-literal=password='$(DBPASS)' \
	  --from-literal=postgres-password='$(DB_ADMIN_PASS)' \
	  --namespace $(NAMESPACE) \
	  --dry-run=client -o yaml | kubectl apply -f -

# psql-002 
# UWAGA! Musi istnieć configmap postgres-custom-config, najpier trzeba więc uruchomic helm.
psql-add-to-cluster:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	helm upgrade --install $(PSQL_RELEASE_NAME) bitnami/postgresql \
	  --namespace $(NAMESPACE) \
	  --create-namespace \
	  --set global.postgresql.auth.username=$(DBUSERNAME) \
	  --set global.postgresql.auth.database=$(DBNAME) \
	  --set global.postgresql.auth.existingSecret=$(PG_SECRET_NAME) \
	  --set primary.existingConfigmap=postgres-custom-config \
	  --set primary.persistence.enabled=true

# PostgreSQL can be accessed via port 5432 on the following DNS names from within your cluster:
#    simplegodb-postgresql.go-hello-ns.svc.cluster.local - Read/Write connection

# psql-003-helper
wait-for-psql:
	@kubectl wait \
		--for=condition=Ready pod \
		-l app.kubernetes.io/instance=$(PSQL_RELEASE_NAME) \
		-n $(NAMESPACE) \
		--timeout=180s


# psql-003-helper
psql-port-forward: wait-for-psql
	@if ! lsof -i :5432 >/dev/null; then \
		echo "Start 5433:5432 port-forward"; \
		kubectl port-forward svc/$(PSQL_RELEASE_NAME)-postgresql 5433:5432 -n $(NAMESPACE) & \
	else \
		echo "Port 5433:5432 already forwarded"; \
	fi

# psql-003
# psql-port-forward
# psql-cli: 
# @PGPASSWORD="$(DBPASS)" \
# psql -h 127.0.0.1 -p 5432 -U $(DBUSERNAME) -d $(DBNAME)
psql-cli:
	kubectl port-forward svc/$(PSQL_RELEASE_NAME)-postgresql 5432:5432 -n $(NAMESPACE) &
	sleep 2
	PGPASSWORD="$(DBPASS)" psql -h 127.0.0.1 -p 5432 -U simplegouser -d simplego


#psql-helper
enter-psql:
	kubectl exec -it simplegodb-postgresql-0 -n go-hello-ns -- bash
	# psql -U postgres
	# Password: admin-pass
	# \du

# TODO- haslo. Możliwość utworzenia dodatkowej bazy danych?
create-simplego-db: psql-port-forward wait-for-psql
	PGPASSWORD=$(DBPASS) psql -h localhost -p 5432 -U $(DBUSERNAME) -f ./db/db.sql

# psql-helper
get-postgres-info:
	@kubectl get secret $(PG_SECRET_NAME) -n $(NAMESPACE) -o yaml

# psql-helper
psql-password:
	@kubectl get secret $(PG_SECRET_NAME) -n $(NAMESPACE) \
	-o jsonpath="{.data.password}" | base64 --decode && echo

# psql-helper - reset
psql-reset:
	helm uninstall $(PSQL_RELEASE_NAME) -n $(NAMESPACE) || true
	kubectl delete pvc -n $(NAMESPACE) -l app.kubernetes.io/name=postgresql || true
	kubectl delete secret $(PG_SECRET_NAME) -n $(NAMESPACE) || true

```
# Postgresql
Co robi ta komenda?:

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
?
Czy zakłada bazę danych?
Czy zakłada użytkownika z dostepem do danych?
Czy muszę tworzyć bazę danych skryptem?
Czy tworzy użytkownika postrgesql? 
Jeżeli tworzy użytkownika postgresql to z jakim hasłem?

Świetne pytania — dokładnie te, które trzeba zadać przy Bitnami PostgreSQL.
Odpowiem konkretnie, punkt po punkcie, bez lania wody.

Co robi ta komenda w całości
helm upgrade --install simplegodb bitnami/postgresql ...


➡️ Instaluje (lub aktualizuje) PostgreSQL w klastrze Kubernetes
➡️ Tworzy:

StatefulSet

Service (ClusterIP, headless)

PersistentVolumeClaim

Secrets (hasła)

Użytkownika + bazę danych (jeśli podasz parametry)

1️⃣ Czy zakłada bazę danych?
✅ TAK
--set global.postgresql.auth.database=$(DBNAME)


➡️ Bitnami tworzy bazę danych o nazwie $(DBNAME) przy pierwszym starcie.

Nie musisz robić:

CREATE DATABASE ...

2️⃣ Czy zakłada użytkownika z dostępem do tej bazy?
✅ TAK
--set global.postgresql.auth.username=$(DBUSERNAME)
--set global.postgresql.auth.password=$(DBPASS)


➡️ Tworzy użytkownika:

$(DBUSERNAME)


➡️ Nadaje mu:

dostęp do bazy $(DBNAME)

ownership tej bazy

pełne prawa w tej bazie

💡 To jest użytkownik aplikacyjny — dokładnie ten, którego powinna używać Twoja aplikacja Go.

3️⃣ Czy muszę tworzyć bazę danych skryptem SQL?
❌ NIE

Bitnami robi to za Ciebie:

Element	Kto tworzy
baza danych	Bitnami Helm Chart
użytkownik	Bitnami Helm Chart
hasło	Ty (DBPASS)

Ty musisz tworzyć tylko:

tabele

indeksy

dane początkowe

Czyli:

CREATE TABLE ...

4️⃣ Czy tworzy użytkownika postgres?
✅ TAK (ZAWSZE)

PostgreSQL zawsze ma superusera postgres.

Bitnami:

tworzy go automatycznie

ustawia MU INNE hasło

5️⃣ Jakie hasło ma użytkownik postgres?
🔐 NIE $(DBPASS)

Hasło postgres:

jest losowo generowane

zapisywane w Secret

Możesz je odczytać tak:

kubectl get secret simplegodb-postgresql \
  -n go-hello-ns \
  -o jsonpath="{.data.postgres-password}" | base64 -d

6️⃣ Co robi ta linia (ważne!)
--set global.postgresql.auth.secretKeys.adminPasswordKey=$(DB_ADMIN_PASS_KEY)


➡️ Mówi Bitnami:

„hasło admina (postgres) będzie zapisane w tym kluczu secreta”

Domyślnie:

postgres-password


Jeśli zmienisz DB_ADMIN_PASS_KEY, to musisz wiedzieć gdzie go potem szukać.

7️⃣ Czy muszę w ogóle używać użytkownika postgres?
❌ NIE (i nie powinieneś)

Dobre praktyki:

aplikacja → $(DBUSERNAME)

migracje → $(DBUSERNAME)

postgres → tylko debug / admin

TL;DR (najkrócej jak się da)
Pytanie	Odpowiedź
Czy tworzy bazę danych?	✅ TAK
Czy tworzy użytkownika app?	✅ TAK
Czy nadaje prawa do DB?	✅ TAK
Czy muszę robić CREATE DATABASE?	❌ NIE
Czy tworzy postgres?	✅ TAK
Czy hasło postgres = DBPASS?	❌ NIE
Gdzie jest hasło postgres?	Secret K8s
🔥 Pro tip (polecam na przyszłość)

Jeśli chcesz 100% kontroli, zamiast --set zrób:

auth:
  username: app
  password: secret
  database: appdb


i:

helm upgrade --install -f values-postgres.yaml


Jeśli chcesz:

pokażę jak zrobić migracje DB w Kubernetes

albo jak podpiąć to idealnie pod Go (env + secrets)

Daj znać 👍
