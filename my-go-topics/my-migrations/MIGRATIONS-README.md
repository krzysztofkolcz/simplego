```
project/
├── db/
│   ├── migrations/
│   │   ├── public/
│   │   │   ├── 001_init.up.sql
│   │   │   └── 001_init.down.sql
│   │   │
│   │   └── tenant/
│   │       ├── 001_init.up.sql
│   │       └── 001_init.down.sql
│   │
│   └── migrate.go   👈 helper w Go
│
├── cmd/
│   └── migrate/
│       └── main.go  👈 CLI do migracji
│
├── go.mod
└── Makefile
```
```
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Uruchomienie migracji public:
```
migrate \
  -path db/migrations/public \
  -database "postgres://mymigrationsuser:mypassword@localhost:5432/mymigrationsdb?sslmode=disable" \
  up
  ```
Uruchomienie migracji tenant:
```
migrate \
  -path db/migrations/tenant \
  -database "postgres://mymigrationsuser:mypassword@localhost:5432/mymigrationdb?sslmode=disable&search_path=tenant_123" \
  up
```

```
make migrate-public
```
1. INSERT INTO tenants
2. CREATE SCHEMA tenant_x
3. migrate tenant schema

# Automatyzacja
Start aplikacji:
1. migrate public
2. migrate wszystkie tenanty

Testy:
1. start postgres (testcontainers)
2. migrate public
3. test tworzy tenant → migrate tenant
```
db/
├── migrations/
│   ├── public/
│   └── tenant/
│
├── migrate.go
└── migrate_all.go   👈 NOWE
```

## Czy robić migracje przy każdym starcie?

👉 TAK (bezpieczne)

golang-migrate:

nie wykona ponownie tych samych migracji
sprawdza schema_migrations

👉 więc:

✔ idempotent
✔ bezpieczne
✔ standard w SaaS

# Testcontainers + migracje (KLUCZOWE)


# TODO
jak przetestować tylko utworzenie schematu i tabeli w bazie?
embed + ifos
Jak robić migracje na klastrze dev / prod
jak robić zero-downtime migrations (expand/contract)
jak zrobić rollback-safe deploy
jak powiązać migracje z deploy pipeline
jak versionować DB razem z kodem


# Deploy 
## Wersjonowanie migracji
build → test → migrate DB → deploy app
migracje są wersjonowane razem z kodem
migracje są idempotentne
rollback = forward fix (najczęściej)

Versioning migracji

golang-migrate używa:

001_init.up.sql
002_add_column.up.sql
003_add_index.up.sql

👉 to jest Twoja „historia DB”

🔥 Zasada:

1 commit = 1 migracja (jeśli zmienia DB)

## Gdzie trzymasz migracje
repo/
├── db/
│   └── migrations/
│       ├── public/
│       └── tenant/
├── internal/
└── cmd/

👉 migracje są w repo = versioned razem z kodem

## Deploy
3. Jak robić migracje na DEV / PROD

Masz 2 podejścia:

🥇 OPCJA A (REKOMENDOWANA): migracje jako osobny krok deployu
pipeline:
1. build
2. test
3. migrate DB
4. deploy app


go run cmd/migrate/main.go

albo:

migrate -path db/migrations/public -database $DB_URL up

👉 dla tenantów:

go run cmd/migrate_all/main.go

🥈 OPCJA B: migracje przy starcie aplikacji
db.MigrateAll(...)
👍 plusy:
prostota
zero infra
👎 minusy:
trudniej kontrolować rollout
ryzyko przy wielu instancjach

👉 moja rekomendacja:

DEV:   auto-migrate przy starcie ✔
PROD:  osobny krok w CI/CD ✔

4. Multitenancy w deployu
krok deployu:
1. migrate public
2. migrate all tenants

👉 czyli:

MigratePublic()
MigrateAllTenants()
## ⚠️ 5. Co z nowymi tenantami?

👉 w runtime:

CreateTenant()
→ CREATE SCHEMA
→ migrate tenant

## 🔙 6. Rollback — brutalna prawda

👉 w 90% przypadków:

❌ NIE używasz down.sql w produkcji

dlaczego?

Bo:

DROP COLUMN name;

= utrata danych ❌

🧠 Prawidłowe podejście: forward-only

Zamiast rollback:

deploy v2 → bug
→ deploy v3 (fix)

## 🔥 7. Expand / Contract pattern

Zamiast:

ALTER TABLE users DROP COLUMN name;

robisz:

krok 1 (expand)
ALTER TABLE users ADD COLUMN full_name TEXT;
krok 2 (deploy code)
krok 3 (contract)
ALTER TABLE users DROP COLUMN name;

👉 zero downtime, zero rollback

## ⚙️ 8. Kiedy używać down.sql?

👉 tylko:

dev
testy
lokalnie

## 🔗 9. Jak powiązać migracje z wersją kodu

Masz 2 opcje:

🥇 Git jako source of truth
commit:
- kod
- migracja

👉 zawsze razem

🥈 wersja aplikacji

np:

v1.2.0 → migracje do 005
v1.3.0 → migracje do 007

## 🧠 10. Jak DB wie, co już było?

golang-migrate tworzy:

schema_migrations

👉 tam są wersje

## 🧪 11. Testowanie w pipeline
1. start testcontainers
2. migrate
3. run tests

👉 jeśli migracja się wywali → deploy STOP


⚠️ 12. Największe błędy (ważne)
❌ brak migracji w repo

→ chaos

❌ ręczne zmiany w DB

→ brak reprodukowalności

❌ rollback przez down.sql

→ utrata danych

❌ migracje po deployu

→ app crash

## 🧭 13. Idealny setup dla Ciebie
✔ golang-migrate + embed
✔ migracje w repo
✔ CI step: migrate
✔ forward-only migrations
✔ expand/contract
✔ tenant migrations batch
✔ CreateTenant → migrate

## 🔥 14. Bonus: zero-downtime deploy
1. deploy migracje (safe)
2. deploy app
3. cleanup (later)

# Migracja bez dostępu githubrunner do postgresql
Jeszcze jedno pytanie. Mam pipeline na github runner. Jendak nie mam dostępu z githubrunnera do postgresql. Do postgresql mam tylko dostęp z klastra. Jak w takim razie mogę uruchomić migrację? 

To jest bardzo realny problem — i dobrze, że go łapiesz teraz, bo to decyzja architektoniczna DevOps, nie tylko kod.

Masz sytuację:

GitHub Runner ❌ brak dostępu do DB
Kubernetes cluster ✅ ma dostęp do DB

👉 Wniosek:

❌ NIE uruchamiasz migracji w GitHub Actions
✅ uruchamiasz migracje wewnątrz klastra

🧠 3 poprawne podejścia (production-grade)
🥇 OPCJA A — Kubernetes Job (REKOMENDOWANE)

👉 pipeline:

GitHub Actions:
  build image
  push image
  kubectl apply job (migration)

Kubernetes:
  uruchamia job
  job robi migracje
🧱 Jak to wygląda?
1. Tworzysz image (z embed migracjami)

Twój binary:

func main() {
    db.MigrateAll(...)
}
2. Job w Kubernetes
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migrate
spec:
  template:
    spec:
      containers:
        - name: migrate
          image: your-app:latest
          command: ["./app", "migrate"]
          env:
            - name: DB_URL
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: url
      restartPolicy: Never
3. GitHub Actions
- name: Run DB migration
  run: |
    kubectl apply -f k8s/migrate-job.yaml
    kubectl wait --for=condition=complete job/db-migrate --timeout=60s
🔥 Zalety:
działa w tej samej sieci co DB ✔
bezpieczne ✔
standard SaaS ✔ 

🧠 Co JA bym zrobił u Ciebie

Biorąc pod uwagę:

SaaS
Kubernetes
multitenancy
Stripe (w przyszłości)

👉 wybór:

✔ Kubernetes Job do migracji
✔ embed migrations
✔ osobny krok w pipeline
✔ app bez auto-migrate w prod
🔁 Flow deployu (idealny)
1. build image
2. push image
3. run migration job  🔥
4. deploy app
⚠️ Co jeśli migracja padnie?

👉 wtedy:

❌ NIE deployujesz aplikacji

Pipeline stop.

🔥 1 ważna rzecz (często pomijana)

👉 Job musi być idempotentny

czyli:

m.Up() // + ErrNoChange
🧪 Testowanie lokalne

Możesz odpalić:

kubectl apply -f migrate-job.yaml
⚡ BONUS — Helm (lepsze)

Jeśli używasz Helm:

hooks:
  - pre-install
  - pre-upgrade

👉 migracja jako hook

🧠 Najczęstsze błędy

❌ migracje z GitHub runnera (brak dostępu)
❌ migracje po deployu
❌ brak wait na job
❌ brak retry

🔚 TL;DR

Masz 3 opcje:

🥇 Kubernetes Job        ← wybierz to
🥈 App startup migrate
🥉 kubectl exec
Jeśli chcesz dalej

Kolejny krok logiczny:

👉 pokażę Ci:

gotowy Helm chart z migracją
albo retry + safe migrations (zero downtime)

Powiedz 👍


iDobra — zróbmy production-ready Helm chart z migracją DB pod Twój setup (Go + embed + multitenancy).

Pokażę Ci najlepszy wzorzec używany w SaaS:

🔥 Helm hook + Kubernetes Job (pre-install / pre-upgrade)

🧠 Co chcemy osiągnąć
helm upgrade:
  1. 🔥 migracja DB (job)
  2. deploy aplikacji

👉 jeśli migracja padnie → deploy NIE idzie dalej

🧱 1. Struktura Helm chart
helm/
└── myapp/
    ├── Chart.yaml
    ├── values.yaml
    └── templates/
        ├── deployment.yaml
        ├── service.yaml
        └── migrate-job.yaml   👈 kluczowe
⚙️ 2. values.yaml
image:
  repository: myapp
  tag: latest

db:
  url: postgres://user:pass@postgres:5432/db?sslmode=disable

migrate:
  enabled: true
🚀 3. Migration Job (najważniejsze)
templates/migrate-job.yaml
{{- if .Values.migrate.enabled }}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "myapp.fullname" . }}-migrate
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": hook-succeeded
spec:
  backoffLimit: 3
  template:
    spec:
      restartPolicy: Never

      containers:
        - name: migrate
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command: ["./app", "migrate"]   # 👈 ważne

          env:
            - name: DB_URL
              value: "{{ .Values.db.url }}"
{{- end }}
🧠 4. Co robi ./app migrate

W Twoim Go:

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations()
		return
	}

	runApp()
}
migracje:
func runMigrations() {
	dsn := os.Getenv("DB_URL")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.MigrateAll(ctx, pool, dsn); err != nil {
		log.Fatal(err)
	}
}
🔥 5. Deployment (aplikacja)
templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp.fullname" . }}
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: app
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command: ["./app"]

          env:
            - name: DB_URL
              value: "{{ .Values.db.url }}"
🧠 6. Jak to działa w praktyce
helm upgrade:
  ↓
Helm:
  → odpala JOB (hook)
  → czeka aż się zakończy
  → dopiero wtedy Deployment
⚠️ 7. CRITICAL: hook-delete-policy
"helm.sh/hook-delete-policy": hook-succeeded

👉 usuwa job po sukcesie
👉 nie zostawia śmieci

🔁 8. Retry
backoffLimit: 3

👉 3 próby

🔒 9. Bezpieczeństwo (PROD)

Zamiast:

value: "postgres://..."

użyj:

env:
  - name: DB_URL
    valueFrom:
      secretKeyRef:
        name: db-secret
        key: url
⚡ 10. Multitenancy — co się stanie?

Twoje:

db.MigrateAll(...)

zrobi:

✔ migrate public
✔ migrate all tenants

👉 więc job ogarnia wszystko

🧪 11. Debug
kubectl logs job/myapp-migrate
⚠️ 12. Najczęstsze błędy

❌ brak hook → migracja nie odpala
❌ migracja w deployment → race condition
❌ brak wait → deploy leci mimo błędu
❌ brak ErrNoChange

🧭 13. Advanced (polecam później)

Możesz dodać:

🔒 advisory lock
SELECT pg_advisory_lock(123);

👉 tylko jedna migracja naraz

⚡ podział:
job:
  migrate-public
  migrate-tenants
🔚 TL;DR

👉 najlepszy setup:

✔ Helm hook (pre-install, pre-upgrade)
✔ Kubernetes Job
✔ embed migrations
✔ app z komendą "migrate"
✔ brak migracji w runtime (prod)
Jeśli chcesz dalej

Mogę teraz pokazać:

🔥 advisory lock + safe migrations (uniknięcie race condition)
⚡ zero-downtime migrations (expand/contract)
🧪 testy migracji w CI + Helm

Powiedz kierunek 👍

testy migracji w CI + Helm

Świetnie — to jest moment, gdzie łączysz CI (testy) + Helm (deploy) w jeden sensowny pipeline. Zrobimy to tak, żeby:

✔ migracje są testowane PRZED deployem
✔ Helm odpala migracje w klastrze
✔ deploy NIE idzie dalej, jeśli migracje padną

Bez hacków.

🧠 Finalny flow (docelowy)
GitHub Actions:
  1. build
  2. test (testcontainers + migracje)
  3. push image
  4. helm upgrade (→ job migracyjny)

Kubernetes:
  5. migration job (hook)
  6. deploy app
🧪 1. TESTY MIGRACJI W CI (klucz)

👉 tu używasz testcontainers — NIE klastra

co testujesz:
✔ czy migracje się wykonują
✔ czy tworzą schematy
✔ czy można zrobić INSERT
przykładowy test (skrót)
func TestMigrations(t *testing.T) {
	ctx := context.Background()

	// 1. start postgres
	container, _ := postgres.RunContainer(ctx)

	dsn, _ := container.ConnectionString(ctx, "sslmode=disable")

	pool, _ := pgxpool.New(ctx, dsn)

	// 2. migrate public
	err := db.MigratePublic(dsn)
	require.NoError(t, err)

	// 3. create tenant
	err = db.CreateTenant(ctx, pool, dsn, "t1")
	require.NoError(t, err)

	// 4. insert test
	_, err = pool.Exec(ctx, `
		SET search_path TO tenant_t1;
		INSERT INTO todos (id, title) VALUES (gen_random_uuid(), 'ok')
	`)
	require.NoError(t, err)
}
🚀 2. GitHub Actions (CI)
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: 1.22

      - name: Run tests (with migrations)
        run: go test ./...

👉 to gwarantuje:

❌ jeśli migracja jest zła → PR nie przejdzie
🐳 3. Build + push image
  build:
    runs-on: ubuntu-latest
    needs: test

    steps:
      - uses: actions/checkout@v4

      - name: Build Docker
        run: docker build -t myapp:${{ github.sha }} .

      - name: Push
        run: docker push myapp:${{ github.sha }}
🚀 4. Deploy + Helm
  deploy:
    runs-on: ubuntu-latest
    needs: build

    steps:
      - name: Helm upgrade
        run: |
          helm upgrade --install myapp ./helm/myapp \
            --set image.tag=${{ github.sha }}
🔥 5. Co robi Helm (ważne)

Twój hook:

annotations:
  "helm.sh/hook": pre-install,pre-upgrade

👉 oznacza:

Helm:
  1. odpala migration job
  2. czeka na sukces
  3. dopiero potem deploy
⚠️ 6. CRITICAL: Helm musi czekać

Dodaj:

helm upgrade ... --wait --timeout 120s

👉 bez tego:

Helm nie poczeka
deploy pójdzie dalej mimo błędu
🧪 7. Czy testujemy Helm w CI?

👉 NIE bezpośrednio

Zamiast tego:

✔ testujesz migracje (Go + testcontainers)
✔ Helm tylko odpala to samo w klastrze
🔁 8. Spójność: CI vs PROD

Masz:

środowisko	migracje
CI	testcontainers
PROD	Helm job

👉 ale:

ten sam kod migracji ✔
ten sam binary ✔
⚠️ 9. Co jeśli migracja padnie w Helm?
✔ job FAIL
✔ helm FAIL
✔ deploy STOP

👉 dokładnie o to chodzi

🔍 10. Debug w prod
kubectl logs job/myapp-migrate
🧠 11. Najważniejsze zasady
🔥 1. testy migracji w CI
zawsze
🔥 2. migracje w klastrze
nie z GitHub runnera
🔥 3. jeden binary
embed migrations
🔥 4. brak różnic CI vs PROD
tylko środowisko
⚠️ 12. Najczęstsze błędy

❌ brak testów migracji
❌ migracje tylko w Helm (bez CI testów)
❌ różne migracje lokalnie vs prod
❌ brak --wait w Helm

🧭 13. Idealny setup (dla Ciebie)
✔ sqlc + golang-migrate (embed)
✔ testcontainers w CI
✔ Helm job (hook)
✔ brak dostępu DB z CI (OK)
✔ multitenancy (MigrateAll)
🔚 Co dalej?

Teraz masz:

migracje ✔
testy ✔
deploy ✔

👉 kolejny naturalny krok:

transakcje + Unit of Work (gdzie BEGIN/COMMIT)

albo:

👉 sqlc + tenant context + repo (pełny flow request → DB)

To są dwa najważniejsze elementy dalej 👍


## hook policy:
```
"helm.sh/hook-delete-policy": hook-succeeded

👉 usuwa job po sukcesie
👉 nie zostawia śmieci
```

## job log
```
kubectl logs job/myapp-migrate
```

## Flow
GitHub Actions:
  1. build
  2. test (testcontainers + migracje)
  3. push image
  4. helm upgrade (→ job migracyjny)

Kubernetes:
  5. migration job (hook)
  6. deploy app

### 1. TESTY MIGRACJI W CI (klucz)

👉 tu używasz testcontainers — NIE klastra

co testujesz:
✔ czy migracje się wykonują
✔ czy tworzą schematy
✔ czy można zrobić INSERT


### Żeby zrobić migracje, chyba muszę skompilować kod?
Dodałem migrację 002.init.up.sql do db/tenant, ale nie poszła migracja

# TODO
Jak powtórzyć migracje które padły? Zapisywać listę tenantów z errorami?
