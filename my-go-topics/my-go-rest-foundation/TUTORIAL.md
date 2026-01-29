https://chatgpt.com/g/g-p-6979069f038081918633e25bb9943f89-nauka-golanga/c/69790b88-3d84-832b-9f90-8c916cedac11

Zrób mi lekcję z pierwszego dnia:
📅 TYDZIEŃ 1 — Solidny REST w Go (fundament)

🎯 Cel: pewnie pisać handlery HTTP bez frameworkowego „magicznego myślenia”

Dzień 1 – Projekt + HTTP podstawy

utwórz repo

main.go

net/http

/health

logging requestów (z wykorzystaniem https://github.com/veqryn/slog-context)

Efekt: serwer startuje, obsługuje requesty

Super wybór kierunku 💪
Jedziemy Dzień 1 – fundament pod „czyste” HTTP w Go, bez frameworkowej magii.

📅 TYDZIEŃ 1 — Dzień 1
Solidny REST w Go (fundament)

🎯 Cel dnia

rozumiesz jak działa net/http

potrafisz uruchomić serwer

masz /health

masz logowanie requestów (middleware)

używasz context + slog-context (bez globali)

Efekt końcowy:
➡️ serwer startuje
➡️ przyjmuje request
➡️ loguje je w spójny sposób

1️⃣ Utwórz repozytorium
mkdir go-rest-foundation
cd go-rest-foundation
go mod init github.com/twoj-login/go-rest-foundation

```
go mod init github.com/krzysztofkolcz/my-go-rest-foundation-002
```


Dodaj zależność:

go get github.com/veqryn/slog-context


Struktura na start (minimalna, ale zdrowa):

.
├── go.mod
├── go.sum
└── main.go

2️⃣ main.go – minimalny serwer HTTP

Zaczynamy od najprostszego możliwego serwera.

package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}


Uruchom:

go run main.go


Test:

curl -i http://localhost:8080/health


✔️ Checkpoint

rozumiesz ServeMux

wiesz czym jest HandlerFunc

serwer działa synchronicznie (1 request = 1 goroutine)

3️⃣ Dlaczego net/http, a nie framework?

Bo:

każdy framework = wrapper na net/http

debugowanie jest prostsze

nie uczysz się „magicznych hooków”

łatwiej pisać testy

Framework = opinia
net/http = fundament

4️⃣ Dodaj logger oparty o slog + context
Dlaczego slog-context?

Bo:

logger żyje w context.Context

możesz go przekazywać przez warstwy

każdy request ma własne metadata

Logger globalny (tylko konfiguracja!)
import (
	"log/slog"
	"os"
)

func initLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
}

5️⃣ Middleware: logowanie requestów

To jest kluczowy moment tej lekcji.

Middleware = funkcja, która:

bierze http.Handler

zwraca http.Handler

import (
	"net/http"
	"time"

	slogctx "github.com/veqryn/slog-context"
	"log/slog"
)

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqLogger := logger.With(
				"method", r.Method,
				"path", r.URL.Path,
			)

			ctx := slogctx.NewContext(r.Context(), reqLogger)

			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			duration := time.Since(start)

			reqLogger.Info("request handled",
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}


💡 Ważne

logger NIE jest globalny

logger żyje w context.Context

downstream (handlery) mogą go używać

6️⃣ Użycie loggera w handlerze /health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	logger := slogctx.FromContext(r.Context())

	logger.Info("health check called")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

7️⃣ Składamy wszystko razem (main.go)
func main() {
	logger := initLogger()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	loggedMux := loggingMiddleware(logger)(mux)

	logger.Info("starting server", "port", 8080)

	err := http.ListenAndServe(":8080", loggedMux)
	if err != nil {
		logger.Error("server failed", "err", err)
	}
}

8️⃣ Test końcowy
curl http://localhost:8080/health


Logi (JSON):

{
  "time":"2026-01-27T20:15:00Z",
  "level":"INFO",
  "msg":"health check called",
  "method":"GET",
  "path":"/health"
}

✅ Co MUSISZ z tego dnia wynieść

Po Dniu 1 powinieneś umieć odpowiedzieć na te pytania:

czym jest http.Handler

czym różni się ServeMux od routera frameworka

jak działa middleware

dlaczego logger w context jest lepszy niż globalny

jak request „przepływa” przez aplikację

📌 Zadanie domowe (ważne!)

1️⃣ Dodaj endpoint /ping
2️⃣ Dodaj log remote_addr
3️⃣ Zmień level logów na DEBUG i zobacz różnicę

Jeśli chcesz, Dzień 2 zrobimy:

własny App struct

dependency injection bez frameworków

test handlera (httptest)

błędy HTTP jako typy

Chcesz jechać dalej 👉 Dzień 2?

Możesz wytłumaczyć mi konstrukcję tej funkcji:

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
?

Świetne pytanie — to kluczowy moment, w którym Go „klika” 🧠
Rozłóżmy tę konstrukcję krok po kroku, bardzo spokojnie, bez skrótów myślowych.

1️⃣ Najpierw: co my tu w ogóle budujemy?

Ta funkcja NIE jest handlerem.

Ona jest:

fabryką middleware

Czyli:

bierze konfigurację (logger)

zwraca funkcję middleware

middleware opakowuje handler

2️⃣ Sygnatura funkcji — czytamy ją jak zdanie
func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler


Czytaj na głos:

Funkcja loggingMiddleware bierze *slog.Logger
i zwraca funkcję,
która bierze http.Handler
i zwraca http.Handler

To dokładnie definicja middleware w net/http:

type Middleware func(http.Handler) http.Handler

3️⃣ Rozbijmy to na etapy (desugaring)

Zapiszmy to bez skrótów, jakby Go NIE miał funkcji anonimowych:

Krok 1: funkcja z loggerem
func loggingMiddleware(logger *slog.Logger) Middleware {
	return middleware
}

Krok 2: middleware
func middleware(next http.Handler) http.Handler {
	return wrappedHandler
}

Krok 3: opakowany handler
func wrappedHandler(w http.ResponseWriter, r *http.Request) {
	// before
	next.ServeHTTP(w, r)
	// after
}

4️⃣ Teraz składamy to w jedną całość

Go pozwala pisać funkcje zwracające funkcje, więc robimy to inline:

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// before
			next.ServeHTTP(w, r)
			// after
		})
	}
}


To jest dokładnie to samo, tylko krócej.

5️⃣ Co tu się dzieje w runtime (BARDZO ważne)
1️⃣ Wywołanie przy starcie aplikacji
mw := loggingMiddleware(logger)


👉 TU:

logger jest „zamknięty” w closure

middleware zostaje skonfigurowany

NIC jeszcze nie obsługuje requestów

2️⃣ Owijanie handlera
wrapped := mw(mux)


👉 TU:

mux staje się next

powstaje nowy handler

nadal brak requestów

3️⃣ Request przychodzi
HTTP request
   ↓
wrapped.ServeHTTP
   ↓
loggingMiddleware BEFORE
   ↓
mux.ServeHTTP
   ↓
loggingMiddleware AFTER

6️⃣ Dlaczego są TU TRZY funkcje?

Bo każda ma inną odpowiedzialność:

Warstwa	Odpowiedzialność
loggingMiddleware(logger)	konfiguracja
func(next http.Handler)	kompozycja
func(w, r)	obsługa requestu
7️⃣ Dlaczego nie prościej?

Bo wtedy:
❌ logger byłby globalny
❌ middleware nie byłby konfigurowalny
❌ testy byłyby trudniejsze

To podejście daje:
✅ dependency injection
✅ czyste testy
✅ brak magii
✅ pełną kontrolę

8️⃣ Minimalny przykład mentalny

Wyobraź sobie:

makeTea := func(herb string) func(water string) string {
	return func(water string) string {
		return water + " + " + herb
	}
}

greenTea := makeTea("green tea")
fmt.Println(greenTea("hot water"))


To dokładnie ten sam wzorzec.

9️⃣ TL;DR (zapamiętaj to jedno)

Middleware w Go to funkcja, która zwraca funkcję, która zwraca handler.
Logger, DB, config → idą do zewnętrznej funkcji.
Request → trafia dopiero do najgłębszej.

Jeśli chcesz, w następnym kroku mogę:

narysować ASCII flow requestu

pokazać jak złożyć 3 middleware

porównać to z Gin / Echo / Chi

Co wybierasz? 👇