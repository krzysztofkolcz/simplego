Tak. I zrobiłbym Ci kurs przygotowany konkretnie pod tę rekrutację, a nie ogólny kurs „Go przed rozmową”.

Patrząc na wymagania Coveron oraz na to, czym już się zajmowałeś w Go, Kubernetes, RabbitMQ, PostgreSQL, Dockerze i mikroserwisach, widzę tu bardzo sensowny cel: przesunąć Cię z „znam technologie i potrafię budować backend” do „potrafię jako Mid/Senior uzasadniać decyzje architektoniczne i prowadzić rozmowę techniczną”.

Oferta jest szczególnie mocno nastawiona na microservices + system design + DDD + event-driven architecture + bezpieczeństwo danych. Oficjalna oferta wymienia Go, MySQL, RabbitMQ, Kubernetes, OpenSearch/Elasticsearch, Redis i Debezium.

Co więcej, Coveron to nie jest przypadkowy CRUD-owy backend: produkt monitoruje m.in. e-maile, telefony, dane finansowe i inne informacje użytkownika, generuje alerty o wyciekach i podejrzanej aktywności oraz obsługuje proces odzyskiwania tożsamości. To bardzo mocno sugeruje pytania o security, consistency, privacy, event processing, idempotency i niezawodność.

🎯 Cel kursu

Po kursie chcę, żebyś potrafił odpowiedzieć na pytania typu:

„Zaprojektuj system monitorowania wycieków danych dla 10 milionów użytkowników.”

i nie tylko narysować pudełka, ale powiedzieć:

dlaczego REST, a dlaczego gRPC,
gdzie użyć RabbitMQ,
gdzie potrzebna jest transakcja,
jak zagwarantować idempotency,
co zrobić z duplikatem eventu,
jak obsłużyć retry,
co zrobić z poison message,
jak zapewnić ordering,
jak skalować consumerów,
kiedy Redis, a kiedy MySQL,
jak użyć Debezium,
dlaczego CDC zamiast publikowania eventu bezpośrednio z aplikacji,
jak zaprojektować DDD,
gdzie przebiega granica bounded context,
jak zabezpieczyć PII,
jak nie wycieknąć danych do logów,
jak działać po awarii RabbitMQ,
jak wdrożyć to w Kubernetes,
jak monitorować system,
i dlaczego właśnie takie rozwiązanie wybrałeś.

To ostatnie jest kluczowe na poziomie Mid–Senior.

Kurs „Coveron Backend Engineer — Go Mid/Senior”

Proponuję 8 modułów, które będziemy przechodzić razem.

Moduł 1 — Go na poziomie Senior

Nie będziemy robić „co to jest slice”.

Skupimy się na rzeczach, które mogą zostać wykorzystane w rozmowie:

Concurrency

goroutines
channels
worker pools
fan-in / fan-out
cancellation
context.Context
backpressure
race conditions
deadlocks
graceful shutdown

Memory / runtime

stack vs heap
escape analysis
GC
allocations
pointers
interfaces
nil interface
generics
copying vs referencing

Go backend

error handling
wrapping errors
sentinel errors
custom errors
dependency injection
interfaces
package design
testing
benchmarks
profiling

Pytania rekrutacyjne

Np.:

Dlaczego ten kod może powodować race condition?

Co się stanie, jeżeli zamkniemy channel dwa razy?

Dlaczego context.Context powinien być pierwszym argumentem funkcji?

Co dokładnie dzieje się z goroutine po anulowaniu contextu?

Dlaczego nie powinno się używać context.Background() wszędzie?

Moduł 2 — Distributed Systems

To będzie jeden z najważniejszych modułów.

Musisz bardzo dobrze rozumieć:

Reliability
retries
exponential backoff
jitter
timeout
circuit breaker
bulkhead
graceful degradation
Distributed consistency
strong consistency
eventual consistency
read-after-write
optimistic locking
distributed transactions
Saga
Idempotency

Np.:

POST /identity-monitoring/check

Request dochodzi do serwera.

Serwer wykonuje operację.

Klient nie dostaje odpowiedzi.

Klient wysyła request ponownie.

Czy wykonamy operację drugi raz?

Jak temu zapobiec?

Idempotency-Key
       ↓
Redis / DB
       ↓
operation

I będziemy rozważać, gdzie dokładnie trzymać idempotency key i dlaczego.

Moduł 3 — RabbitMQ + Event Driven Architecture

To prawdopodobnie będzie bardzo ważny fragment przygotowania, ponieważ RabbitMQ znajduje się bezpośrednio w stacku Coveron.

Przejdziemy:

RabbitMQ
exchange
queue
routing key
binding
direct
topic
fanout
competing consumers
acknowledgement
nack
requeue
dead-letter exchange
TTL
retry
prefetch
publisher confirms
Problemy produkcyjne

Co jeśli:

DB transaction
      ↓
publish event

DB commituje.

RabbitMQ nie dostaje eventu.

Mamy niespójność.

Rozwiązaniem będzie:

Transactional Outbox
                ┌─────────────┐
                │   MySQL     │
                │             │
Request ───────►│ business    │
                │ data        │
                │             │
                │ outbox      │
                └──────┬──────┘
                       │
                       │ CDC
                       ▼
                 Debezium
                       │
                       ▼
                  RabbitMQ
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
      service A    service B    service C

I tutaj naturalnie przejdziemy do Debezium.

Moduł 4 — Debezium + CDC

To jest rzecz, której szczególnie nie chciałbym u Ciebie pominąć.

Musisz rozumieć:

Change Data Capture

czyli:

MySQL
  │
  │ binlog
  ▼
Debezium
  │
  ▼
Kafka / RabbitMQ / downstream

Będziemy omawiać:

po co Debezium,
binlog,
CDC,
transactional outbox,
event publication,
ordering,
duplicate events,
schema evolution,
snapshot,
replay,
exactly-once vs at-least-once.

I bardzo ważne pytanie:

„Dlaczego nie zrobić po prostu INSERT do DB, a następnie rabbit.Publish()?”

Musisz umieć odpowiedzieć bez zastanawiania się.

Moduł 5 — DDD

Oferta mówi wprost:

Passionate about Domain Driven Design

więc potraktowałbym to poważnie.

Nie będziemy uczyć się DDD jako zestawu definicji.

Zrobimy DDD na przykładzie Coveron.

Spróbujemy rozbić domenę:

Coveron
│
├── Identity Monitoring
│
├── Dark Web Monitoring
│
├── Credit Monitoring
│
├── Alerts
│
├── Identity Recovery
│
├── Insurance
└── User / Subscription

Potem:

Bounded Context
Entity
Value Object
Aggregate
Aggregate Root
Domain Event
Application Service
Domain Service
Repository
Anti-Corruption Layer

I najważniejsze:

Jak znaleźć granicę mikroserwisu?

Bo:

1 mikroserwis = 1 tabela

jest złym podejściem.

Moduł 6 — System Design

To będzie centralny moduł kursu.

Będziemy robić prawdziwe zadania rekrutacyjne.

Zadanie 1

Zaprojektuj system dark-web monitoring dla 10 mln użytkowników.

Zaczynamy od:

Requirements
       ↓
Traffic estimation
       ↓
API
       ↓
Data model
       ↓
Services
       ↓
Events
       ↓
Queues
       ↓
Storage
       ↓
Caching
       ↓
Scaling
       ↓
Failure handling
       ↓
Security
       ↓
Observability
Zadanie 2

Zaprojektuj system real-time security alerts.

Zadanie 3

Zaprojektuj system wykrywający, że email użytkownika pojawił się w wycieku.

Zadanie 4

Zaprojektuj pipeline:

external data
      ↓
ingestion
      ↓
processing
      ↓
matching
      ↓
user notification
Zadanie 5

Zaprojektuj system identity recovery.

I będę Cię przepytywał jak interviewer:

„Dlaczego RabbitMQ?”

„Dlaczego nie Kafka?”

„Co jeśli consumer padnie?”

„Co jeśli event zostanie przetworzony dwa razy?”

„Jak zapewnisz ordering?”

„Co jeśli Redis padnie?”

„Dlaczego MySQL?”

„Jak to skalujesz 10×?”

„Co jest bottleneckiem?”

Moduł 7 — Kubernetes + Production

Tutaj masz już sporą bazę, więc nie będziemy robić podstawowego kursu Kubernetes.

Skupimy się na interview-level Kubernetes:

Deployment
Service
Ingress
ConfigMap
Secret
probes
requests / limits
HPA
rolling update
graceful shutdown
readiness vs liveness
pod disruption
autoscaling
resource starvation
observability
zero-downtime deployment

I bardzo ważne pytanie:

„Masz 20 replik consumerów RabbitMQ. Jak Kubernetes wpływa na sposób konsumowania wiadomości?”

Moduł 8 — Security

Ten moduł będzie szczególnie istotny właśnie dla Coveron.

Bo firma zajmuje się PII i identity theft protection.

Musisz umieć rozmawiać o:

Authentication
OAuth2
JWT
access token
refresh token
service-to-service authentication
Authorization
RBAC
ABAC
least privilege
Data
encryption at rest
encryption in transit
TLS
secrets
key management
hashing
encryption vs hashing
PII

Najważniejsze pytanie:

„Czy możesz zalogować SSN użytkownika?”

Odpowiedź oczywiście:

nie.

Ale będziemy rozważać:

logs
metrics
traces
database
backups
queues
errors

bo PII może wyciec nie tylko z DB.

🧠 Osobny blok: API Design

Oferta bardzo mocno podkreśla API.

Przerobimy:

REST
resource modeling
HTTP semantics
status codes
pagination
filtering
sorting
versioning
idempotency
error format
rate limiting

Np.:

GET /v1/monitoring/assets

vs

GET /v1/users/{id}/monitoring-assets

i będziemy rozmawiać dlaczego.

gRPC
protobuf
unary RPC
streaming
deadlines
metadata
interceptors
retries
service-to-service communication

I przede wszystkim:

REST vs gRPC — kiedy co wybrać?

🗄️ MySQL

Nie wystarczy znać:

SELECT *
FROM users;

Będziemy robić:

indexes
composite indexes
covering indexes
transactions
isolation levels
deadlocks
locks
MVCC
optimistic/pessimistic locking
query plans
replication
read replicas
partitioning

Przykładowe pytanie:

Masz tabelę z 500 mln rekordów. Query nagle trwa 8 sekund. Co robisz?

⚡ Redis

Skupimy się na rzeczywistych zastosowaniach:

Redis
├── cache
├── rate limiter
├── distributed lock
├── idempotency
├── session
└── ephemeral state

I najważniejsze:

Kiedy Redis jest dobrym rozwiązaniem, a kiedy używanie go jest błędem?

🔎 OpenSearch / Elasticsearch

Musisz rozumieć przede wszystkim dlaczego w takim systemie w ogóle potrzebujemy search engine.

Np.:

MySQL
   │
   │ events
   ▼
OpenSearch
   │
   ▼
search / investigation / analytics

Przerobimy:

inverted index
indexing
analyzers
full-text search
shards
replicas
eventual consistency
relevance
indexing pipeline
🧪 Coding interview

Nie chcę, żebyśmy robili 100 losowych zadań LeetCode.

Dla tej rekrutacji zrobiłbym około 25–30 zadań w Go, podzielonych na:

Easy
slices
maps
strings
intervals
queues
stacks
Medium
concurrency
worker pool
rate limiter
LRU cache
producer/consumer
concurrent map
retry mechanism
TTL cache
Backend exercises

Na przykład:

Napisz worker pool obsługujący 1000 jobów.

Potem interviewer dodaje:

„Teraz każdy job może trwać 30 sekund.”

Potem:

„Dodaj timeout.”

Potem:

„Dodaj retry.”

Potem:

„Dodaj graceful shutdown.”

Potem:

„Teraz mamy 5 instancji serwisu.”

I nagle zwykłe zadanie staje się system design + Go concurrency.

To jest dokładnie sposób, w jaki chciałbym Cię przygotowywać.

🗣️ Senior / behavioral interview

Tego absolutnie nie pominiemy.

Oferta mówi:

mentor colleagues

shape team's vision

collaborate with product/design/engineering

więc mogą pytać:

„Opowiedz o sytuacji, kiedy nie zgadzałeś się z decyzją techniczną.”
„Kiedy ostatnio zmieniłeś zdanie?”
„Jak przekonałeś zespół do swojej architektury?”
„Jak pomogłeś mniej doświadczonemu developerowi?”
„Opowiedz o błędzie architektonicznym, który popełniłeś.”
„Jaki był Twój największy wpływ na projekt?”

To ostatnie jest szczególnie interesujące, bo niedawne relacje kandydatów do Nord Security wskazują na pytania o wpływ na projekt i product/business impact, nie tylko o czystą wiedzę techniczną.

Jest też relacja kandydata z 2025 r., według której proces obejmował HR, technical i culture interview, a część techniczna mogła być live codingiem albo zadaniem domowym.

Nie traktowałbym tego jako gwarantowanego przebiegu dla Coveron w 2026, ale jako sygnał, jak się przygotować.

🏗️ Najważniejszy projekt kursu

Chciałbym zrobić z Tobą jeden duży projekt:

„Coveron-like Identity Monitoring Platform”

Architektura:

                         ┌──────────────┐
                         │   Frontend   │
                         └──────┬───────┘
                                │
                         REST / gRPC
                                │
                         ┌──────▼───────┐
                         │ API Gateway  │
                         └──────┬───────┘
                                │
              ┌─────────────────┼──────────────────┐
              │                 │                  │
        ┌─────▼─────┐    ┌──────▼─────┐    ┌──────▼─────┐
        │ Identity  │    │ Monitoring │    │   Alerts   │
        │ Service   │    │  Service   │    │  Service   │
        └─────┬─────┘    └──────┬─────┘    └──────┬─────┘
              │                 │                  │
              ▼                 ▼                  ▼
            MySQL            RabbitMQ            Redis
                                │
                                ▼
                           ┌──────────┐
                           │ Debezium │
                           └────┬─────┘
                                │
                                ▼
                          ┌───────────┐
                          │ OpenSearch│
                          └───────────┘

I zbudujemy go w Go.

Nie po to, żeby mieć projekt na GitHubie.

Po to, żeby każdą decyzję móc potem wykorzystać podczas rozmowy:

„Dlaczego zaprojektowałeś system właśnie w ten sposób?”

📅 Proponowany harmonogram

Jeżeli będziesz miał około 60–90 minut dziennie, zrobiłbym:

Tydzień	Temat	Priorytet
1	Go advanced + concurrency	⭐⭐⭐⭐⭐
2	REST + gRPC + API design	⭐⭐⭐⭐
3	RabbitMQ + event-driven	⭐⭐⭐⭐⭐
4	DDD + microservices	⭐⭐⭐⭐⭐
5	MySQL + Redis + Debezium	⭐⭐⭐⭐⭐
6	System Design	⭐⭐⭐⭐⭐
7	Kubernetes + production	⭐⭐⭐⭐
8	Security + OpenSearch	⭐⭐⭐⭐
9	Coding interview	⭐⭐⭐⭐
10	Mock interviews	⭐⭐⭐⭐⭐

Ale nie musimy czekać 10 tygodni. Możemy zrobić to intensywnie.

🔥 Jak będziemy pracować

Najważniejsza rzecz: nie chcę Ci po prostu wykładać materiału.

Każdy temat będzie miał:

1. Teorię

Krótko i konkretnie.

2. Diagram

Żebyś widział system.

3. Przykład w Go
4. Zadanie

Ty rozwiązujesz.

5. Review

Ja robię review jak Senior Engineer.

6. Interview questions

Przepytuję Cię.

7. „Senior answer”

Pokazuję Ci, jak odpowiedzieć podczas rozmowy, a nie tylko jak rozwiązać problem.

🎯 Co szczególnie będziemy wzmacniać u Ciebie

Z tego, co już przerabialiśmy, nie zaczynałbym od zera z Go, Dockerem czy Kubernetesem.

Masz już praktyczny kontakt m.in. z:

Go,
PostgreSQL,
GORM/Prisma,
RabbitMQ,
Docker,
Kubernetes,
Helm,
Testcontainers,
Stripe,
CI/CD,
mikroserwisami.

Dlatego największy zwrot z czasu widzę w:

1. System Design

2. Distributed Systems

3. RabbitMQ + event-driven architecture

4. DDD

5. Debezium / transactional outbox / CDC

6. MySQL internals

7. Security / PII

8. Umiejętność obrony decyzji technicznych

To właśnie może być różnica między:

„dobry Mid developer”

a

„kandydat, którego można zatrudnić jako Seniora”.

I proponuję jeszcze jedną rzecz

Nie będziemy robić kursu „na ślepo”. Będę Cię prowadził jak przez prawdziwy proces rekrutacyjny Coveron.

Zaczniemy od Lekcji 1: „Jak wygląda system design interview na Mid/Senior Go — od pustej kartki do architektury”.

Na pierwszym ćwiczeniu dostaniesz ode mnie tylko wymaganie:

„Zaprojektuj system monitorowania wycieków danych dla użytkowników Coveron.”

Ty spróbujesz go zaprojektować bez przygotowania, tak jak na rozmowie. Ja następnie zrobię Ci bardzo szczegółowy review: co powiedziałeś dobrze, czego zabrakło, gdzie odpowiedź jest seniorska, a gdzie brzmi jeszcze jak Mid — i na tej podstawie będziemy budować dalszy kurs.

To będzie znacznie skuteczniejsze niż rozpoczęcie od kolejnego wykładu.