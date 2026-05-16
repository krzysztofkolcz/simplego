Agregat Subscription
  
  Jeden abonament na tenanta, żyje w schemacie tenanta. Dwa pola domenowe:

  plan:   free | pro
  status: active | cancelled

  Metody agregatu:
  - Upgrade() — free → pro, error jeśli już pro lub cancelled
  - Downgrade(todoCount int) — pro → free, error jeśli todoCount > 10 lub cancelled
  - Cancel() — error jeśli już cancelled
  - CanCreateTodo(todoCount int) error — free + todoCount >= 10 → ErrPlanLimitReached

  Domain events: SubscriptionCreated, SubscriptionUpgraded, SubscriptionDowngraded, SubscriptionCancelled

  ---
  Integracja z CreateTodo

  CreateTodoUseCase.Handle przed delegowaniem do CreateTodoHandler:
  1. Odczytuje aktualny abonament (read-only)
  2. Liczy aktualne todo tenanta
  3. Wywołuje subscription.CanCreateTodo(count) — jeśli error, zwraca go
  4. Dopiero wtedy tworzy todo
  
  HTTP handler mapuje ErrPlanLimitReached → 422 (Unprocessable Entity).

  ---
  Endpointy:

  ┌────────┬─────────────────────────────┬────────────────────────────────────┐
  │ Method │            Path             │                Opis                │
  ├────────┼─────────────────────────────┼────────────────────────────────────┤
  │ POST   │ /v1/subscriptions           │ Utwórz (przy onboardingu tenanta)  │
  ├────────┼─────────────────────────────┼────────────────────────────────────┤
  │ GET    │ /v1/subscriptions/current   │ Odczytaj plan i status             │
  ├────────┼─────────────────────────────┼────────────────────────────────────┤
  │ PUT    │ /v1/subscriptions/upgrade   │ Upgrade do pro                     │
  ├────────┼─────────────────────────────┼────────────────────────────────────┤
  │ PUT    │ /v1/subscriptions/downgrade │ Downgrade do free (sprawdza limit) │
  ├────────┼─────────────────────────────┼────────────────────────────────────┤
  │ DELETE │ /v1/subscriptions           │ Anuluj                             │
  └────────┴─────────────────────────────┴────────────────────────────────────┘

  ---
  Kolejność implementacji (od dołu do góry):

  1. Migracja SQL (003_subscription)
  2. Zapytania sqlc + generowanie
  3. Agregat domenowy + inwarianty + domain events
  4. Repository interface + implementacja
  5. Unit of Work
  6. Command handlery + query handler
  7. OpenAPI spec → oapi-codegen
  8. HTTP handler
  9. Use case adaptery + wiring w main.go
  10. Integracja z CreateTodo (limit check)
  11. Testy integracyjne