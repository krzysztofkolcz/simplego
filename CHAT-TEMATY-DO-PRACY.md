W takim razie w dobie AI i przygotowań do rozmowy rekrutacyjnej - na ile znać szczegóły golanga, na ile bardziej architekturę systemów, lub inne narzędzia, a na ile co innego? Co jest ważne w tych rozmowach rekrutacyjnych, do czego się przygotowywać?

Tak. I w Twojej sytuacji nie próbowałbym już zostać „chodzącą dokumentacją Go”. Przy rekrutacji na Mid/Senior Backend Engineer ważniejsze jest pokazanie, że potrafisz zaprojektować, zbudować, zdiagnozować i utrzymać system, niż pamiętanie wszystkich szczegółów składni.

AI jeszcze bardziej przesuwa tę granicę.

Jak podzieliłbym przygotowanie

Dla stanowiska typu Senior/Mid-Senior Go Backend Engineer zrobiłbym mniej więcej:

Obszar	Waga	Co powinieneś umieć
System design / architektura	25%	projektowanie systemów, trade-offy, skalowanie
Go — praktyka i internals	20%	concurrency, memory, interfaces, errors, performance
Backend / distributed systems	20%	DB, cache, queues, consistency, idempotency
Production / DevOps	15%	Kubernetes, observability, CI/CD, debugging
Coding / algorytmy	10%	rozwiązywanie problemów bez pomocy AI
Behavioral / komunikacja	10%	decyzje techniczne, ownership, incydenty, współpraca

To oczywiście zależy od firmy. W firmie infrastrukturalnej Go może być 30–40%, a w typowym produkcie SaaS system design i backend mogą dominować.

1. Go — ale nie „wszystko z Go”

Nie uczyłbym się:

„Muszę znać każdy szczegół standard library.”

Zamiast tego musisz bardzo dobrze rozumieć mechanizmy, które mają wpływ na zachowanie programu.

Absolutne podstawy

Powinieneś swobodnie wyjaśnić:

slice vs array,
map,
pointer/value semantics,
interfaces,
method sets,
embedding,
error handling,
defer,
panic/recover,
goroutines,
channels,
select,
mutex/RWMutex,
atomic operations,
context,
race conditions,
cancellation,
worker pools,
graceful shutdown.

Na przykład interviewer:

„Co się stanie, jeśli przekażę slice do funkcji i zmodyfikuję jego element?”

Powinieneś odpowiedzieć bez zastanawiania się.

Ale jednocześnie:

„Dlaczego append może spowodować zmianę underlying array?”

— tutaj warto już rozumieć mechanizm, a nie tylko znać regułę.

2. Go internals — tylko te, które mają znaczenie

Na poziomie Seniora bardzo dobrze znałbym:

Memory
stack vs heap,
escape analysis,
allocations,
garbage collector,
pointers,
object lifetime.
Runtime
scheduler,
G/M/P,
goroutines,
work stealing,
GC.
Concurrency

Nie tylko:

„channel służy do komunikacji.”

Ale:

„Dlaczego ten kod może mieć race condition?”

„Kiedy użyłbyś mutexa zamiast kanału?”

„Jak ograniczyć liczbę równoległych requestów?”

„Co stanie się z goroutine, która czeka na channel po anulowaniu requestu?”

To są pytania znacznie bardziej senioralne.

3. System design — tutaj inwestowałbym bardzo dużo

To prawdopodobnie największy obszar, który warto teraz rozwijać.

Musisz umieć dostać pytanie:

„Zaprojektuj system do wysyłania powiadomień.”

i w 30–45 minut przejść przez:

requirements
      ↓
API
      ↓
data model
      ↓
architecture
      ↓
scaling
      ↓
failure modes
      ↓
observability
      ↓
trade-offs

I przede wszystkim zadawać pytania.

Np.:

Ile requestów/s?

Czy wiadomość może zostać dostarczona dwa razy?

Czy kolejność ma znaczenie?

Jaki SLA?

Czy potrzebujemy strong consistency?

Ile danych?

Jak długo przechowujemy dane?

To jest bardzo ważne.

Senior nie zaczyna od:

„Postawmy Kafkę.”

Senior zaczyna od:

„Jakie mamy wymagania?”

4. Distributed systems

To będzie bardzo wartościowe szczególnie dla Go.

Powinieneś rozumieć:

Messaging
RabbitMQ,
Kafka,
NATS,
at-least-once delivery,
at-most-once,
exactly-once jako problem praktyczny,
consumer groups,
acknowledgements,
retries,
dead-letter queues.
Distributed transactions
ACID,
isolation levels,
optimistic/pessimistic locking,
transactions,
outbox pattern,
saga.
Reliability
timeout,
retry,
exponential backoff,
circuit breaker,
idempotency,
deduplication.
Consistency
strong consistency,
eventual consistency,
CAP,
replication,
leader/follower.

Nie musisz znać matematycznych dowodów CAP.

Musisz umieć powiedzieć:

„Tutaj eventual consistency jest OK, ponieważ...”

i:

„Tutaj potrzebujemy idempotency key, ponieważ retry może spowodować podwójne obciążenie klienta.”

To robi dużo większe wrażenie niż znajomość 50 metod API.

5. PostgreSQL

Bardzo mocno.

Backend developer bez dobrej znajomości bazy jest mocno ograniczony.

Przygotowałbym:

indexes,
B-tree,
composite indexes,
EXPLAIN,
query planning,
transactions,
isolation levels,
locks,
deadlocks,
MVCC,
pagination,
N+1,
connection pool,
replication,
partitioning.

I klasyczne pytanie:

„Ta tabela ma 100 milionów rekordów. Zapytanie zaczęło trwać 5 sekund. Co robisz?”

Nie:

„Dodaję indeks.”

Tylko:

„Najpierw EXPLAIN ANALYZE.”

6. Kubernetes

Tutaj nie musisz być Kubernetes architectem.

Ale powinieneś rozumieć produkcyjny model działania aplikacji.

Czyli:

Internet
   ↓
Load Balancer
   ↓
Ingress
   ↓
Service
   ↓
Pods
   ↓
Container

oraz:

Deployment,
ReplicaSet,
Service,
ConfigMap,
Secret,
probes,
requests/limits,
HPA,
rolling deployment,
graceful shutdown,
logs,
metrics.

I przede wszystkim:

„Pod jest CrashLoopBackOff. Jak to diagnozujesz?”

Powinieneś mieć procedurę:

kubectl get pod
kubectl describe pod
kubectl logs
kubectl logs --previous
kubectl get events

a następnie rozumować, dlaczego się zepsuło.

7. Observability

To jest moim zdaniem niedoceniany temat na rozmowach senioralnych.

Powinieneś rozumieć:

logs + metrics + traces

oraz:

latency,
throughput,
error rate,
saturation,
RED,
USE,
correlation/request IDs,
distributed tracing,
Prometheus,
OpenTelemetry.

Pytanie:

„Klienci mówią, że API jest wolne. Co robisz?”

Dobra odpowiedź nie brzmi:

„Patrzę w logi.”

Tylko:

„Najpierw sprawdzam, czy problem jest globalny czy dotyczy konkretnego endpointu/regionu/klienta. Patrzę na latency percentiles, error rate, saturation, następnie trace konkretnego requestu i schodzę do DB/external dependencies.”

8. Security

W Twoim przypadku zdecydowanie warto.

Szczególnie:

OAuth2/OIDC,
JWT,
refresh tokens,
sessions,
RBAC,
password hashing,
secrets,
TLS,
encryption,
OWASP Top 10,
SSRF,
SQL injection,
CSRF,
authentication vs authorization.

I tutaj Twoje wcześniejsze zainteresowanie Zitadel/Casbin/SuperTokens jest bardzo sensowne.

Nie musisz znać wszystkich produktów.

Musisz rozumieć problem.

9. Coding interview

Tutaj AI robi dużą różnicę.

Firmy nadal mogą sprawdzać:

„Napisz funkcję X.”

Ale coraz bardziej wartościowe jest pokazanie, że potrafisz myśleć samodzielnie.

Dlatego ćwiczyłbym:

arrays,
maps,
strings,
linked lists,
trees,
graphs,
BFS/DFS,
heaps,
binary search,
sliding window,
two pointers,
basic dynamic programming.

Ale nie spędzałbym 200 godzin na LeetCode.

Dla Mid/Senior:

30–50 dobrze dobranych zadań, rozumianych naprawdę dobrze, będzie lepsze niż 300 rozwiązań zapamiętanych schematycznie.

10. AI — i tutaj jest ciekawa rzecz

Na rozmowie nie próbowałbym udowadniać:

„AI nie jest mi potrzebne.”

Wręcz przeciwnie.

Warto umieć powiedzieć:

„Używam AI do X, Y i Z, ale w tych obszarach nie ufam mu bezpośrednio.”

Na przykład:

AI świetnie:

boilerplate,
test cases,
dokumentacja,
refactoring,
eksplorowanie API,
code review,
generowanie SQL jako punkt startowy.

Ale Ty odpowiadasz za:

architecture,
security,
correctness,
concurrency,
data integrity,
performance,
production behavior.

To jest bardzo dobra odpowiedź na dzisiejszy rynek.

Najważniejsza rzecz: doświadczenie

I tutaj masz potencjalnie największy skarb.

Nie przygotowywałbym się wyłącznie poprzez książki.

Weź 5–10 rzeczy, które rzeczywiście zrobiłeś w pracy, i przygotuj się tak, żeby móc o każdej mówić przez 10 minut.

Np.:

„Mieliśmy problem z X.”

„Rozważyłem A/B/C.”

„Wybrałem B, ponieważ...”

„Wdrożenie spowodowało...”

„Popełniliśmy błąd...”

„Po wdrożeniu znaleźliśmy problem...”

„Następnym razem zrobiłbym...”

To jest złoto na Senior interview.

Przygotowałbym się również do pytań typu
„Opowiedz o trudnej decyzji technicznej.”
„Opowiedz o błędzie, który popełniłeś.”
„Opowiedz o produkcyjnym incydencie.”
„Jak przekonałeś zespół do swojej decyzji?”
„Co zrobiłbyś inaczej?”
„Kiedy wybrałeś prostsze rozwiązanie zamiast bardziej skalowalnego?”
„Jak radzisz sobie z niezgodą techniczną?”
„Jak oceniasz kod wygenerowany przez AI?”
„Co robisz, gdy nie znasz technologii?”

To są pytania, które pokazują seniority, a nie znajomość składni.

Gdybym miał Ci rozpisać kolejność nauki

Biorąc pod uwagę Twój cel — zmiana na znacznie lepiej płatną pracę Go/backend — zrobiłbym:

Etap 1 — Go

2–3 tygodnie

Concurrency, memory, interfaces, errors, context, runtime.

Etap 2 — PostgreSQL + backend

2 tygodnie

Transactions, isolation, indexes, locks, performance.

Etap 3 — Distributed Systems

3 tygodnie

Queues, retries, idempotency, consistency, caching, outbox.

Etap 4 — System Design

3–4 tygodnie

10–15 systemów projektowanych od zera.

Etap 5 — Kubernetes + production

1–2 tygodnie

Deployment, networking, observability, failures.

Etap 6 — interview

ciągle

Coding + behavioral + mock interviews.

I zacząłbym aplikować już podczas Etapu 2, a nie po ukończeniu całej ścieżki.

I jedna bardzo ważna rzecz w kontekście Nord

Jeżeli oferta, o której wcześniej rozmawialiśmy, była na Mid–Senior Go Backend Engineer, to nie przygotowywałbym się tak, jak do egzaminu z języka Go.

Przygotowywałbym się tak, żeby interviewer po 60–90 minutach pomyślał:

„Temu człowiekowi mogę dać backendowy problem produkcyjny i on samodzielnie go rozłoży, zaprojektuje rozwiązanie, napisze kod i będzie wiedział, jak to później monitorować.”

To jest właśnie poziom, który może uzasadniać 2–3× większe wynagrodzenie.

I dlatego, patrząc na Twój obecny etap, nie inwestowałbym teraz kolejnych dziesiątek godzin w budowanie SaaS boilerplate. Twoja największa dźwignia może być obecnie w przygotowaniu do zmiany pracy, a SaaS rozwijałbym równolegle bardzo małym kosztem czasu.