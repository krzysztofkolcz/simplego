# Plan nauki — Go Senior + Architektura

> Ostatnia aktualizacja: 2026-07-14

---

## Gdzie jesteś teraz

Na podstawie tego repozytorium: jesteś solidnym mid-level Go developerem.

**Masz już:**
- REST API (Chi, OAPI codegen, OpenAPI validation)
- DDD — domain, application (command/query), infrastructure
- Multi-tenancy, SQLC, pgx, transakcje
- Outbox pattern, graceful shutdown, slog/slogctx
- Kubernetes, Helm, testcontainers (integracyjne)
- Podstawy concurrency (goroutines, channels)

**Czego brakuje do Senior:**
- Testowanie na poważnie (unit testy, testy architektury, benchmarki, fuzz)
- Observability (metryki, distributed tracing — nie tylko logi)
- gRPC + protobuf
- Message broker (Kafka lub NATS) — tylko wspomniany, nieimplementowany
- Advanced concurrency (sync primitives, context propagation, race detector)
- Profiling & performance (pprof, trace)
- System design judgment — kiedy CZEGO używać i dlaczego

---

## Struktura repozytorium (propozycja)

```
simplego/
├── PLAN_NAUKI.md               ← ten plik
├── DZIENNIK.md                 ← dziennik postępu (już masz)
├── book-100gomistakes/         ← ćwiczenia z "100 Go Mistakes" (już masz)
├── go-with-domain/             ← eksperymenty DDD (już masz)
│
├── my-go-topics/               ← tutoriale z konkretnych narzędzi (już masz)
│   ├── my-migrations/          ← najbardziej zaawansowany projekt
│   ├── my-http-server-002/     ← OAPI codegen
│   └── ...
│
├── patterns/                   ← NOWE: czyste wzorce bez frameworków
│   ├── concurrency/            ← worker pool, pipeline, fan-out/fan-in, semaphore
│   ├── resilience/             ← circuit breaker, retry, rate limiter
│   ├── cqrs-eventsourcing/     ← event store, projections
│   └── saga/                   ← choreography vs orchestration
│
├── projects/                   ← NOWE: kompletne mini-projekty z myślą o produkcji
│   ├── grpc-service/           ← gRPC + protobuf + reflection
│   ├── kafka-consumer/         ← konsument Kafka z at-least-once delivery
│   ├── observability/          ← Prometheus metrics + OpenTelemetry tracing
│   └── auth-service/           ← JWT + RBAC, production-ready
│
└── challenges/                 ← NOWE: system design ćwiczenia (README + implementacja)
    ├── url-shortener/
    ├── rate-limiter/
    └── distributed-cache/
```

---

## Faza 1 — Solidne fundamenty (6–10 tygodni)

**Cel:** Zamknąć luki w podstawach Go, żeby pisać kod który przechodzi code review u seniora.

### 1.1 Testowanie — priorytet #1

Twoje obecne testy to głównie integracyjne z testcontainers. To za mało na seniora.

**Co zrobić:**
- Zapoznaj się z table-driven tests (standardowy wzorzec w Go)
- Naucz się testować bez mocków — dependency injection przez interfejsy + fake (nie mock)
- `go test -race` jako nawyk przy każdym commicie
- Benchmarki: `func BenchmarkXxx(b *testing.B)` + `go test -bench=.`
- Fuzz testing: `func FuzzXxx(f *testing.F)` — minimum jeden przykład

**Ćwiczenie w tym repo:**
- Dodaj unit testy do `my-migrations/internal/domain/` — każda metoda domenowa pokryta
- Napisz fake repository (w pamięci) dla co najmniej jednego use case'a
- Zmierz benchmark dla jednego query SQLC

**Zasoby:**
- Dave Cheney: "Prefer table driven tests" (blog)
- "Learn Go with Tests" — learngotests.com (darmowe, praktyczne)

---

### 1.2 Concurrency — od podstaw do produkcji

Channels/goroutines masz. Brakuje wzorców do produkcyjnego użycia.

**Co zrobić:**
- `sync.WaitGroup`, `sync.Mutex`, `sync.RWMutex`, `sync.Once`, `sync.Map`
- Context cancellation chain — jak poprawnie propagować, jak wykryć wyciek goroutine
- `errgroup` (`golang.org/x/sync/errgroup`) — górna granica goroutine'ów
- Race detector jako obowiązek: `go test -race ./...`

**Wzorce do zakodowania (w `patterns/concurrency/`):**
```
worker_pool/    — N workerów, M zadań, graceful shutdown
pipeline/       — fan-out/fan-in
semaphore/      — ograniczenie równoległości przez kanał buforowany
ticker_worker/  — periodic job bez wycieku (twój OutboxWorker, ale czysty)
```

**Zasoby:**
- "Concurrency in Go" — Katherine Cox-Buday (książka, ~200 str — warto)
- Rob Pike: "Concurrency is not Parallelism" (YouTube, 30 min)

---

### 1.3 Dokończ "100 Go Mistakes"

Już zacząłeś. Masz rozdziały 9 (generics), 10 (embedding), 11 (functional options).

**Plan:**
- 2–3 rozdziały tygodniowo, każdy z ćwiczeniem w `book-100gomistakes/`
- Najważniejsze rozdziały dla mid→senior: **#12 (testing), #8 (goroutines/channels), #4 (control flow)**
- Przy każdym błędzie: napisz test który go demonstruje, potem poprawkę

---

## Faza 2 — Production Services (8–12 tygodni)

**Cel:** Skończyć z "działa lokalnie". Umieć wyciągnąć produkcyjny serwis go z metrykamt, tracingiem i gRPC.

### 2.1 Observability — metryki + tracing

Masz logi (slog). Brakuje dwóch pozostałych filarów.

**Metryki — `projects/observability/`:**
- Prometheus client Go: counter, gauge, histogram
- Middleware zliczający HTTP requests/latency/errors
- `/metrics` endpoint
- Uruchom lokalnie Prometheus + Grafana przez docker-compose (jeden plik, w repo)

**Distributed Tracing:**
- OpenTelemetry SDK dla Go (`go.opentelemetry.io/otel`)
- Propagacja trace przez HTTP headers (W3C TraceContext)
- Jaeger lub Tempo jako backend (docker-compose)
- Zinstrumentuj jeden z obecnych serwisów (my-migrations → handler → repo → db)

**Dlaczego to senior-skill:** Bez observability nie jesteś w stanie debugować produkcji. Logi mówią CO, metryki mówią ILE, tracing mówi GDZIE.

---

### 2.2 gRPC + Protobuf

Masz tylko REST. Na rynku (szczególnie microservices) gRPC jest wszechobecne.

**Co zrobić w `projects/grpc-service/`:**
- Zdefiniuj `.proto` — 2–3 serwisy, streaming (unary + server stream)
- Wygeneruj Go kod (`protoc` + `protoc-gen-go`)
- Napisz server + client
- Dodaj interceptors (odpowiednik middleware): logging, error handling, auth
- Reflection serwera (umożliwia `grpcurl` bez `.proto`)

**Dodaj gRPC do istniejącego my-migrations:**
- Ten sam use case co HTTP, ale przez gRPC — to ćwiczy separation of concerns

---

### 2.3 Message Broker — dokończ outbox

W `my-migrations` masz `OutboxWorker` który mockuje publikację. Czas zintegrować z prawdziwym brokerem.

**Co zrobić w `projects/kafka-consumer/`:**
- Kafka przez docker-compose (bitnami/kafka, jeden kontener)
- Producer: Go → Kafka (użyj `segmentio/kafka-go` lub `confluentinc/confluent-kafka-go`)
- Consumer: Kafka → Go, at-least-once, idempotent processing
- Dead Letter Queue — co gdy przetwarzanie failuje

**Potem:** Podłącz OutboxWorker z my-migrations do tego Kafka producera — zamiast `println`.

---

## Faza 3 — Architektura zaawansowana (8–12 tygodni)

**Cel:** Rozumieć tradeoffs, umieć zaproponować architekturę i uzasadnić ją na design review.

### 3.1 Event Sourcing + CQRS (pełne)

Masz CQRS (command/query split). Brakuje event sourcing — gdzie stanem systemu jest log eventów, a nie ostatni rekord w bazie.

**W `patterns/cqrs-eventsourcing/`:**
- Event Store (PostgreSQL jako event log — wystarczy na start)
- Aggregaty z metodami Apply(event)
- Projekcje — read models budowane z eventów
- Snapshot co N eventów (optymalizacja)

**Kiedy używać:** Audit log, undo/redo, temporal queries. Kiedy NIE używać: proste CRUD, wysoki throughput bez audytu.

---

### 3.2 Saga Pattern

Dla distributed transactions bez 2PC (które skaluje się słabo).

**W `patterns/saga/`:**
- Choreography saga — każdy serwis reaguje na eventy poprzedniego
- Orchestration saga — centralny orchestrator (Saga Orchestrator)
- Compensating transactions — co gdy krok N failuje, jak cofnąć kroki 1..N-1

**Ćwiczenie:** "Order service" — złóż zamówienie (sprawdź stock → zarezerwuj payment → wyślij)

---

### 3.3 Resilience Patterns

**W `patterns/resilience/`:**
- Circuit Breaker (użyj `sony/gobreaker` albo napisz własny)
- Retry z exponential backoff + jitter
- Rate Limiter — token bucket (stdlib `golang.org/x/time/rate`) + sliding window
- Bulkhead — izolacja pul zasobów

**Każdy wzorzec:** Implementacja + test który pokazuje co się dzieje gdy downstream failuje.

---

## Faza 4 — Senior Mindset (ciągłe)

**Cel:** Przejście od "piszę kod" do "projektuję systemy i podejmuję decyzje architektoniczne".

### 4.1 System Design — ćwicz na głos / na papierze

Senior różni się od mida tym, że zanim napisze linię kodu, myśli o systemie.

**Ćwiczenia w `challenges/`:**
- Zaprojektuj URL shortener (write-heavy vs read-heavy, caching)
- Zaprojektuj rate limiter (distributed, Redis, Lua scripts)
- Zaprojektuj distributed cache (eviction policies, consistency)
- Dla każdego: napisz `DESIGN.md` z decyzjami i tradeoffs, POTEM implementuj

**Format ADR (Architecture Decision Record):**
```markdown
# ADR-001: Wybór brokera wiadomości

## Status: Accepted
## Kontekst: ...
## Decyzja: Kafka zamiast RabbitMQ
## Konsekwencje: ...
```

---

### 4.2 Code Review — ucz się czytać kod

- Przeglądaj open source Go projektów: `go-chi/chi`, `jackc/pgx`, `hashicorp/raft`
- Czytaj PR-y, nie tylko kod — patrz jak seniorzy komentują
- Każdy swój projekt: zanim commitujesz, zrób self-code-review i zapisz co byś zmienił

---

### 4.3 Performance Engineering

**Profiling (nie optymalizuj bez danych):**
- `go tool pprof` — CPU profile, memory profile, goroutine profile
- `go tool trace` — timeline goroutines
- Zinstrumentuj jeden endpoint z my-migrations i zmierz gdzie idzie czas

**Zasada seniora:** "Measure, don't guess. Optimize the bottleneck, not the code you wrote today."

---

## Jak pracować z AI (jak ze mną — Claude)

AI zmienił rynek, ale nie zmienił czego potrzebujesz jako senior.

### Czego AI NIE zastąpi:
- **System design judgment** — AI zaproponuje rozwiązanie, ale nie wie co jest ważne dla TWOJEGO produktu i TWOICH constraints
- **Debugging produkcji** — musisz rozumieć co się dzieje, żeby wiedzieć co pytać
- **Code review** — rozumienie DLACZEGO coś jest złe, nie tylko że jest złe
- **Komunikacja i mentoring** — tego AI za ciebie nie zrobi

### Jak używać AI żeby SZYBCIEJ się uczyć (nie żeby nie uczyć się wcale):
1. **Pytaj o tradeoffs, nie o odpowiedzi.** Zamiast "napisz mi circuit breaker" → "jakie są tradeoffs między circuit breakerem a retry z exponential backoff i kiedy wybrać jedno zamiast drugiego?"
2. **Używaj AI jako pair programmera, nie ghostwritera.** Napisz kod, potem daj AI do code review. Nie odwrotnie.
3. **Generuj ćwiczenia.** "Daj mi 5 edge cases do przetestowania w moim rate limiterze" — AI jest świetny do tego.
4. **Wyjaśniaj kod AI własnymi słowami.** Jeśli nie możesz wytłumaczyć czemu coś działa, nie rozumiesz tego.
5. **Nie akceptuj kodu AI bez zrozumienia każdej linii.** To najszybsza droga do bycia seniorem który jest zależny od AI.

### Workflow nauki z AI:
```
Temat → Przeczytaj dokumentację/książkę → Napisz PoC SAM
     → Code review z AI → Popraw → Zrozum komentarze
     → Napisz testy SAMEMU → Użyj AI do edge cases
```

---

## KPIs — skąd będziesz wiedzieć że rośniesz

| Co | Jak sprawdzić |
|---|---|
| Concurrency | Napisz worker pool bez data race (`go test -race`) |
| Testing | Coverage >80% w domain/, każdy use case ma fake repo |
| Observability | Serwis z metrykami w Prometheusie, tracing w Jaegerze |
| gRPC | Działający gRPC server + client + interceptors |
| Kafka | Producer/consumer z at-least-once + DLQ |
| System design | Napisz DESIGN.md dla 3 różnych systemów zanim zaczniesz kodować |
| Code review | Daj swój kod do review — czy rozumiesz każdy komentarz? |
| Performance | Wiesz gdzie jest bottleneck w jednym z serwisów (profiling) |

---

## Kolejność priorytetów (jeśli masz ograniczony czas)

1. **Testowanie** — bez tego nie ma code review, nie ma refactoringu, nie ma seniora
2. **Observability** — Prometheus + OpenTelemetry (szybki win, duża wartość)
3. **gRPC** — rynek tego wymaga
4. **Concurrency patterns** — zamknij luki, zanim zabijesz produkcję wyścigiem
5. **Kafka** — po gRPC, bo podobna filozofia (async, message-driven)
6. **Event Sourcing + Saga** — dopiero gdy masz solidne fundamenty

---

## Zasoby

| Zasób | Kiedy |
|---|---|
| "100 Go Mistakes" (już masz) | Teraz, równolegle |
| "Concurrency in Go" — Katherine Cox-Buday | Faza 1 |
| learngotests.com — "Learn Go with Tests" | Faza 1 |
| "Designing Data-Intensive Applications" — Kleppmann | Faza 3 (must-read dla każdego seniora) |
| ardanlabs.com/training | Ultimate Go — dobre dla concurrency |
| bytesizego.com | Newsletter, krótkie case studies produkcyjne |
| github.com/ThreeDotsLabs/wild-workouts-go-ddd-example | Referencja DDD w Go (produkcyjna jakość) |

---

## Dziennik postępu

Uzupełniaj `DZIENNIK.md` po każdej sesji — co zrobiłeś, co zrozumiałeś, co dalej.
Format: data, temat, co wyszło, co nie wyszło, następny krok.
