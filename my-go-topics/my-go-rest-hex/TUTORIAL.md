# C2
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

## Application layer
orkiestracja
przypadki użycia systemu
spójność transakcyjna
granica systemu
Use case = jedna intencja użytkownika

Transakcje w application layer

# C3 CQRS
```
internal/
├── application/
│   └── users/
│       ├── commands/
│       │   └── create_user.go
│       └── queries/
│           ├── get_user.go
│           └── list_users.go
```
Jak rozumiem, przenoszę /users/create_user.go do users/commands/create_user.go
Zmieniam nazwę na CreateUserCommand i zwracam jedynie error?

Jak rozumiem przenoszę users/get_user.go do user/queries/get_user.go itd.?

## Commands
NIE zwraca danych
tylko sukces / błąd
zapisuje stan

