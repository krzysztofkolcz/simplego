# Organization
DZIENNIK.md - dziennik aktywności

## 2 podejścia:
### Każdy temat osobno, minimalny deployment, za to ogarniam inne tematy
my-helm/
    01-basic-deploy
    02-config-secret

### Wszystkie tematy razem
helm/
    go-hello
    ...
    TUTORIAL.md - w tutorialu każdy temat opisany osobno.

### Ćwiczenia - usuwam
helm/
    go-hello-exercise

lub

my-helm/
    01-basic-deploy-exercise

# Go topics
https://www.youtube.com/watch?v=wr8gJMj3ODw

Linter
Authentication topics - should I create MVP for ‘Supertokens’, and other auth libraries
HTTP server + openapi, middleware
Openapi templates for generating Domain driven design files
DB topics, Transaction topics, isolation levels etc.
GORM
Go routines
Grpc

## 4 Go books
https://www.youtube.com/shorts/YNuhuo4QpJw
Learning Go - O'Reilly - https://www.oreilly.com/library/view/learning-go-2nd/9781098139285/
gRPC Microservices in Go - https://www.oreilly.com/library/view/grpc-microservices-in/9781633439207/
(Microservices patterns - concepts)

Domain Driven Design with Golang - https://www.packtpub.com/en-us/product/domain-driven-design-with-golang-9781804619261?srsltid=AfmBOoqJV4TTT7ndtjIa4wubImBmnHUN4kkfqfQOFZQB2h0exhxcY08W

Concurrency in Go O'Reilly -  https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/


## List of books on github
https://github.com/dariubs/GoBooks?tab=readme-ov-file#2025---pro-go-patterns-advanced-function-design-concurrency-models-and-clean-code

https://github.com/irezaul/go-life

## Libraries
slogctx


## Books
https://go.dev/doc/effective_go
100 go mistakes - 1 topic/week
Domain Driven Design - 1 topic/week
https://lets-go-further.alexedwards.net/
https://quii.gitbook.io/learn-go-with-tests
https://interpreterbook.com/

https://www.oreilly.com/library/view/designing-data-intensive-applications/9781098119058/

## VS Code
VS Code - keyboard shortcuts
Code snippets
Plugins

Plugin VS Code: "golangci-lint" by golangci

## Devops
Grafana topics

Kubernetes topics
helm charts
K9s

## Program ideas
Simple programs:
Creating folders
Processing nef to jpeg with go routines



# Tematy na Saas
## Skrót
### DDD, architektura
### Autoryzacja, biblioteka do logowania/rejestracji
### Baza danych
### sqlx, sqlc, gorm
### Migracje bazy danych
    automatyczne migracje
    rollback
    backup bazy
    migracje zero downtime
### Wersjonowanie backend, frontend, infra, testy
### Testy jednostkowe
### Testy react
### Testy E2E, API i UI
### load testing
### Prometheus, grafana, loki - logi, 
### monitoring
### opentelemetry
### CI/CD (GitHub Actions)
### retry dla webhooków
### security scanning - co to?
### logowanie strukturalne - co to?
    structured logs (slog + slogctx)
    request ID
    trace ID
### helm/kubernetes
### kafka/rabbit - inne mniej zasobożerne biblioteki
### env - staging, prod
### alerting - jak?

##
https://chatgpt.com/g/g-p-6974db69d3dc819196dfb45bdb7bd10a-saas/c/69a55e7b-40a8-8332-82ec-786f23a05fc6

🧠 1️⃣ Architektura – Fundament
Backend (Go)
✅ Struktura projektu – DDD + Clean Architecture

Proponowana struktura:

/cmd/api
/internal
    /domain
        user.go
        subscription.go
    /application
        user_service.go
        subscription_service.go
    /infrastructure
        /db
        /stripe
        /auth
    /interfaces
        /http
            handlers
            middleware
/pkg
Warstwy
domain → czyste modele + reguły biznesowe
application → use-case’y
infrastructure → GORM, Stripe, SuperTokens
interfaces → HTTP, middleware

To jest skalowalne i gotowe pod zespół.

🔐 Autoryzacja

Masz 3 opcje:

Opcja	Kiedy
SuperTokens (self-hosted)	SaaS, wiele tenantów
Keycloak	duży system
Własne JWT	mniejszy projekt

Jeśli budujesz SaaS produkcyjny → SuperTokens jest sensowny kompromis.

💳 Stripe

Dobrze, że już myślisz o billing logic (ostatnie rozmowy 😉)

Musisz mieć:

webhook handler (idempotentny!)
mapowanie:
Stripe customer → user_id
subscription → plan
retry policy
obsługę:
upgrade
downgrade
cancel
failed payment
🗄 PostgreSQL – GORM czy sqlx?
GORM
szybciej startujesz
migracje
– mniej kontroli
sqlx
pełna kontrola
lepsza wydajność
– więcej kodu

👉 Jeśli budujesz SaaS, który ma rosnąć → ja bym:

używał sqlc albo sqlx
unikał ciężkiego ORM

Ale GORM nie jest błędem.

🧪 2️⃣ Testy – tutaj większość SaaS odpada
Backend
1️⃣ Testy jednostkowe
testujesz domain i application
mockujesz repozytoria
2️⃣ Testy API

Najprościej:

net/http + httptest

Lub:

uruchamiasz testowy Postgres (Docker)
seedujesz dane
testujesz prawdziwe endpointy

Możesz też użyć:

testcontainers-go
🎭 3️⃣ Frontend – React
Testy frontendowe
Jednostkowe
Vitest albo Jest
React Testing Library

Testujesz:

komponent
formularz
walidację
Integracyjne
mock API
sprawdzasz flow
🎬 4️⃣ Testy E2E (bardzo ważne)

Najlepsze narzędzie:

🔥 Playwright

Testujesz:

rejestrację
logowanie
zakup planu
zmianę planu
anulowanie
🌍 5️⃣ Środowiska

Minimum:

Środowisko	Cel
local	development
staging	testy E2E
production	produkcja
Jak zarządzać?

W Kubernetes:

/helm
   /values-dev.yaml
   /values-staging.yaml
   /values-prod.yaml

Oddzielne:

namespace
baza danych
Stripe (test vs live)
☸️ 6️⃣ Infrastruktura

Masz już dobre kierunki.

Kubernetes
deployment
service
ingress
secrets
HPA
Monitoring

Minimum:

Prometheus
Grafana
Loki (logi)
📊 7️⃣ Obserwowalność (tu SaaS robi się „enterprise”)

Backend:

structured logs (slog + slogctx)
request ID
trace ID

Warto dodać:

OpenTelemetry
distributed tracing
📈 8️⃣ Co często brakuje w SaaS

To są rzeczy, o których mało kto myśli:

🔄 Migracje bazy
zero downtime
backward compatible
🔐 Rate limiting
📧 Email system
resend / SES
retry
template system
🧾 Audit log
🧑‍💻 Feature flags
📦 Background jobs
asynq / redis
kolejka do maili
retry
🏗 9️⃣ Organizacja repo – monorepo czy multi?

Dla SaaS solo dev:

👉 Monorepo

/backend
/frontend
/infra

Wersjonowanie:

tagujesz całość jako jeden release
frontend i backend kompatybilne
🧠 10️⃣ Twój stack – czy jest kompletny?

Patrząc na Ciebie (Go + K8s + Stripe + SaaS mindset):

Brakuje jeszcze:

CI/CD (GitHub Actions)
automatyczne migracje
rollback
backup bazy
load testing
security scanning
🎯 11️⃣ Docelowy poziom SaaS „A-klasa”

Jeśli chcesz, żeby aplikacja była „na wysokim poziomie”, powinna mieć:

✅ testy jednostkowe
✅ testy API
✅ testy E2E
✅ monitoring
✅ logowanie strukturalne
✅ retry dla webhooków
✅ migracje zero downtime
✅ CI/CD
✅ staging
✅ backup
✅ alerting

💬 Odpowiedzi na Twoje konkretne pytania
„Testy API (jak?)”

→ httptest + testcontainers-go + prawdziwa baza

„Jak testować front?”

→ React Testing Library + Playwright

„Czy potrzebuję dwóch środowisk?”

→ Tak. Minimum staging + prod.

„Jak tym zarządzać?”

→ Helm values + osobne namespace + osobne bazy



## Implementacja serwera produkcyjnego - tematy chagGPT
https://chatgpt.com/g/g-p-6979069f038081918633e25bb9943f89-nauka-golanga/c/698df9c8-0778-8329-a605-e1e6921db781

🧱 1. Warstwa HTTP / Transport
✅ Routing i kontrakt API
OpenAPI Generator / oapi-codegen
Strict request/response validation
Versioning API (/v1, /v2)
Content negotiation
Idempotency support (np. header Idempotency-Key)
Pagination standard
Sorting/filtering standard
Error response standard (RFC7807 / Problem Details)
✅ Middleware (bardzo ważne)

Masz już część, ale pełna lista production:

Core
request logging (structured)
request ID / correlation ID
panic recovery
timeout middleware
rate limiting
authentication
authorization
CORS
compression (gzip / brotli)
request size limiting
body replay / buffering (do logów)
Observability middleware
metrics
tracing
request context propagation
✅ JSON handling
Strict decoding
Unknown fields rejection
Validation (np. go-playground/validator)
Custom error mapper validation → OpenAPI response
🧠 2. Context i Request Lifecycle

Production serwer MUSI mieć spójny lifecycle requestu:

✅ Context propagation
request-scoped logger
trace/span injection
tenant context (SaaS!)
user identity context
✅ Deadline / cancellation
respect ctx.Done()
DB queries z context
external calls z context
🗄 3. Warstwa persistence (PostgreSQL)
✅ Connection management
pool tuning
connection health checks
retry logic
read/write split (opcjonalnie)
✅ Migration strategy
goose / atlas / migrate
backward compatible migrations
schema version monitoring
✅ Transaction management
Unit of Work pattern
Retry for serialization failures
Context-aware transactions
✅ Query safety
prepared statements
timeout per query
slow query logging
🔐 4. Security

Production absolutnie kluczowe:

✅ Transport security
TLS / mTLS (jeśli internal services)
HSTS
secure headers
✅ Auth
JWT / OAuth2 / OIDC
token refresh
RBAC / ABAC
scope validation
✅ Input security
validation
sanitization
SQL injection protection
JSON schema validation
✅ Abuse protection
rate limiting
brute force protection
request quotas per tenant
📊 5. Observability (SRE ready)

Masz już OTel logs — dodaj resztę:

✅ Metrics
Prometheus metrics
RED metrics (Rate / Errors / Duration)
Business metrics (!)
✅ Tracing
OpenTelemetry tracing
DB spans
external HTTP spans
message queue spans
✅ Logging
structured logs
correlation ID
log levels
sampling
PII redaction
✅ Health checks
liveness probe
readiness probe
startup probe
dependency checks
⚙️ 6. Configuration

Production Go server zawsze potrzebuje:

✅ Config system
env vars
config validation
hot reload (opcjonalnie)
feature flags
secrets manager integration
🧪 7. Testing

Często pomijane, a krytyczne:

✅ Unit tests
handler tests
service tests
repository tests
✅ Contract tests
OpenAPI contract validation
backward compatibility tests
✅ Integration tests
testcontainers PostgreSQL
full HTTP stack tests
✅ Load tests
k6 / vegeta
latency SLA
📦 8. Error Handling Strategy

Production system zawsze ma:

✅ Error taxonomy
domain errors
infrastructure errors
HTTP mapping
✅ Error wrapping
errors.Is / errors.As
stack traces
error codes
🔄 9. Resilience / Reliability

Mega ważne w SaaS:

✅ Retry policies
exponential backoff
circuit breakers
✅ Outbox / Inbox pattern

(do eventów)

✅ Idempotency

(do payments / webhooks)

✅ Background workers
job queue
retry + DLQ
🧩 10. Multi-tenancy (jeśli SaaS)

Ty często robisz SaaSy więc dorzucam:

tenant isolation
tenant quotas
tenant config
tenant logging separation
🚀 11. Kubernetes Readiness
✅ Graceful shutdown (masz 👍)

Dodatkowo:

preStop hook
drain connections
finish inflight requests
✅ Resource management
memory limits awareness
GC tuning
pprof endpoint
✅ Scaling readiness
stateless handlers
sticky session avoidance
📚 12. Documentation / Developer Experience

Production = łatwo rozwijać:

OpenAPI docs
architecture ADR
runbooks
error catalog
onboarding docs
🧰 13. Dependency Management
dependency injection
module boundaries
internal packages
clean architecture / hexagonal
🔄 14. CI/CD readiness
lint
static analysis
vulnerability scanning
contract diff
migration verification
container scanning
⭐ BONUS – elementy enterprise

Jeśli celujesz wysoko:

audit trail
feature toggles
distributed locking
schema registry
replay engine
SLA monitoring
canary deployment support
📌 Minimalny Production Blueprint (moja rekomendacja)

Jeśli miałbym zrobić "must-have":

HTTP Layer
 ├ OpenAPI strict handlers
 ├ Middleware stack
 ├ Validation
 └ Error mapping

Service Layer
 ├ Business logic
 └ Domain errors

Persistence
 ├ Repository pattern
 ├ Transaction manager
 └ Migrations

Observability
 ├ Logs
 ├ Metrics
 └ Tracing

Security
 ├ AuthN/AuthZ
 └ Rate limiting

Resilience
 ├ Retries
 └ Idempotency

Infra
 ├ Config
 ├ Graceful shutdown
 └ Health checks
💬 Moja opinia (z doświadczenia Go + SaaS + K8s)

Najczęściej pomijane, a potem bolą:

Error taxonomy
Idempotency
Tenant context
Contract tests
Slow query observability
Retry / circuit breakers
Business metrics
👉 Jeśli chcesz iść poziom wyżej

Mogę Ci zaprojektować:

🧱 gotowy production template Go servera:
OpenAPI strict server
OTel
PostgreSQL
middleware stack
SaaS ready
Kubernetes ready
Clean architecture
testcontainers
outbox pattern

Jeśli chcesz — napisz tylko:

👉 czy chcesz template:

minimalistyczny
production SaaS
enterprise event-driven

Mogę Ci rozpisać architekturę + strukturę katalogów + przykładowe implementacje.