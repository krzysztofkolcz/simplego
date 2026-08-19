# Sprint 1-2 tygodnie — Coveron (Nord Security) Backend Engineer Go

> Utworzono: 2026-08-08. Scalenie priorytetów z NORDSECURITY-CHAT.md (trafność w JD) i PLAN_NAUKI.md (istniejący szkielet repo, nacisk na testing).
> Zasada: **nie budujemy nowego projektu od zera** — rozszerzamy to, co już jest w `my-go-topics/` (zwłaszcza outbox worker), bo to szybciej daje materiał do obrony na rozmowie niż nowy szkielet.

## Dlaczego tak, a nie 10 tygodni z NORDSECURITY-CHAT.md

Przy 1-2 tygodniach nie ma miejsca na pełne 8 modułów. Zostawiamy tylko to, co JD wymienia wprost lub co realnie pada na rozmowach Mid/Senior Go, i wplatamy testing (którego oryginalny kurs prawie nie dotyka) w każdy dzień zamiast na końcu.

## Dzień 0 (dziś) — Diagnostyka

Zrób bez przygotowania, na czas ~30 min:

> "Zaprojektuj system monitorowania wycieków danych (dark web monitoring) dla 10 mln użytkowników Coveron."

Napisz odpowiedź w `DZIENNIK.md` albo osobnym pliku — potem robimy review jak interviewer (co brzmi senior, co jeszcze mid, czego zabrakło). To ustawia resztę sprintu pod Twoje realne luki, nie zgadywane.

## Dni 1-3 — System Design + Distributed Systems (najwyższy priorytet)

- Reliability: retry + exponential backoff + jitter, timeout, circuit breaker
- Idempotency: idempotency-key w Redis/DB, dlaczego "nie wykonamy operacji drugi raz"
- Consistency: strong vs eventual, read-after-write, optimistic locking
- 2-3 zadania system design z NORDSECURITY-CHAT (dark-web monitoring, real-time alerts, identity recovery) — na głos, z przepytywaniem "dlaczego RabbitMQ a nie Kafka", "co jeśli consumer padnie"

## Dni 4-6 — RabbitMQ + Transactional Outbox + Debezium (must-have, bo dosłownie w stacku)

**Ćwiczenie na istniejącym kodzie:** znajdź `OutboxWorker` w `my-go-topics` (wspomniany w PLAN_NAUKI.md jako mockujący publikację) i:
1. Podłącz prawdziwego RabbitMQ (docker-compose, jeden kontener) zamiast mocka
2. Dodaj publisher confirms + obsługę nack/requeue + dead-letter exchange
3. Napisz test integracyjny (masz już testcontainers) sprawdzający: DB commit + brak publikacji = spójność zachowana dzięki outbox
4. Przeczytaj/zrozum Debezium jako alternatywę do ręcznego outbox workera (CDC z binloga zamiast publikowania z aplikacji) — nie trzeba go stawiać, ale musisz umieć narysować diagram i uzasadnić "dlaczego CDC zamiast rabbit.Publish() wprost po INSERT"

To jeden dzień daje Ci realny, obronny artefakt zamiast tylko teorii.

## Dni 7-8 — DDD dla domeny Coveron + Security/PII

- Rozbij domenę Coveron na bounded contexts (Identity Monitoring, Dark Web, Credit, Alerts, Recovery, Insurance, User/Subscription) — papier/diagram, nie kod
- Przejrzyj istniejący `go-with-domain/` pod kątem: gdzie granica agregatu, co powinno być Value Object
- Security: co NIE loguje się nigdy (SSN, dane finansowe), encryption at rest/in transit, gdzie PII może wyciec poza DB (logi, metryki, kolejki, błędy, backupy)
- **Delta MySQL vs Postgres** (używasz pgx/SQLC na co dzień): gap locks, różnice w izolacji, replication — 1-2h wystarczą, nie potrzebujesz pełnego kursu MySQL

## Dzień 9 — Coding interview (Go)

5-6 zadań, nie 30: worker pool z timeoutem + retry + graceful shutdown (stopniowane jak w oryginalnym kursie — "teraz dodaj X"), rate limiter, LRU/TTL cache. Do każdego: table-driven test + `go test -race`.

## Dzień 10-12 — Mock interview + behavioral

- Pełny mock system design (któreś zadanie, którego jeszcze nie robiłeś) z pytaniami dociskowymi
- Behavioral: przygotuj 3-4 konkretne historie (konflikt architektoniczny, mentoring, błąd i jego naprawa, wpływ na projekt/biznes) — nie improwizuj na miejscu

## Bufor

Zostaw 1-2 dni luzu — coś się zawsze przesunie.

---

Powiązane: [[NORDSECURITY.md]] (oferta), [[NORDSECURITY-CHAT.md]] (pełny kurs 10-tyg., źródło priorytetów), [[PLAN_NAUKI.md]] (długoterminowy plan Senior, źródło struktury repo i nacisku na testing).
