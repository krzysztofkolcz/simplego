Struktura od chata

```
internal/
├── domain/
│   ├── user.go             // domain - regóły biznesowe, brak zależności
│   ├── errors.go
│   └── user_repository.go   // PORT
│
├── application/                    // Transakcje w tej warstwie
│   └── users/
│       ├── create_user.go   // use case, CreateUserInput, CreateUserOutput
│       ├── get_user.go
│       └── list_users.go
│
├── adapters/
│   ├── http/
│   │   └── users_handler.go        // adapter http, port wyjściowy -> application/users/create_user.go h.CreateUser.Execute
│   └── postgres/
│       └── user_repository.go      // adapter postgres, port wyjściowy, implementuje domain.UserRepository
│
├── app/
│   └── app.go                      // App = wiring (jedno miejsce)
```

Struktura z książki:
```
├── xxx
│   ├── xxx
│   └── xxx
│       ├── xxx
│       └── xxx
│           ├── xxx
│           └── xxx
│               ├── xxx
│               └── xxx
```

```
internal/
├── trainer/
│   ├── domain/
│   |   └── hour/
│   |       ├── availability.go 
│   |       ├── hour.go 
│   |       └── repository.go  // PORT?
│   ├── adapters/                       // podobnie jak chat, tylko adapters pod trainer - czyli konkretny kontekst (?)
│   |   ├── hour_memory_repository.go  // adapter in memory. implementuje repository (podobnie jak u chata)
│   |   └── hour_postgres_repository.go  // adapter postgresql. implementuje repository
│   ├── app/
│   |   ├── command/ 
│   |   |   ├── cancel_training.go 
│   |   |   ├── ....go 
│   |   |   └── ....go 
│   |   ├── query/
│   |   |   ├── available_hours.go 
│   |   |   ├── ....go 
│   |   |   └── ....go 
│   |   └── app.go  // app zawiera wszystkie command i queries
│   ├── ports/
│   |   ├── http.go // HttpServer zawiera app, z wszystkimi commands i queries
```

```
		if err := c.hourRepo.UpdateHour(ctx, hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeAvailable(); err != nil {
				return nil, err
			}
			return h, nil
		}); err != nil {
			return errors.NewSlugError(err.Error(), "unable-to-update-availability")
		}
```

TODO:
event-driven architecture, 
event-sourcing, 
polyglot persistence,