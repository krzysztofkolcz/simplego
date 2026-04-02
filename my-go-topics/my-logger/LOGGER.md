# Streszczenie
slog.InfoContext - pobiera domyślny logger ustawiony globalnie przez slog.setDefault
slog.LogAttr - standarodwa funkcja. Wypisuje log. Korzysta z slog.Attr.

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)) - tworzy logger z handlerem json. Ten handler nie ma dostępu do kontekstu definiowanego przez funkcje slogctx.
logger := slog.New(handler).With("service", "billing-api") - dodanie stałych pól do loggera
logger.InfoContext - logger to konkretna instancja logera. W handlerze loggera mogę mieć dostęp do zmiennych zawartych w kontekście
logger.Info - bez dostępu do kontekstu

handler := slogctx.NewHandler( slog.NewJSONHandler(os.Stdout, nil), nil,) logger := slog.New(handler) - tworzy handler, który ma dostęp do kontekstu definiowanego przez funkcje slogctx
slogctx.Info - pobiera logger z kontekstu
ctx := slogctx.NewCtx(ctx, logger) - zapisuje logger do kontekstu
ctx := slogctx.Append( ctx, "request_id", "abc123") - dodaje atrybut do kontekstu na koniec listy
ctx := slogctx.Prepend - jw. dodaje atrybut do kontekstu ale na początku
ctx := slogctx.With(ctx, slog.String("taskType", task.Type())) - zwraca kontekst z atrybutem "taskType"
slogctx.LogAttr - jak slog.LogAttr - wypisuje log, ale bardziej wydajnie, korzysta z slog.Attr.

# Poziomy logów (slog.Level)

slog ma poziomy:

LevelDebug = -4
LevelInfo  = 0
LevelWarn  = 4
LevelError = 8

Czyli dla LevelInfo widoczne będą LevelInfo, LevelWarn i LevelError

# slog.InfoContext vs logger.InfoContext vs slogctx.Info
## slog.InfoContext 
```
slog.InfoContext(ctx, "slog default logger - processing request") 
```

pobiera domyślny logger ustawiany przez 
```
slog.setDefault(logger)
```
### Flow
```
ctx
 ↓
slog.InfoContext
 ↓
slog.Default()
 ↓
handler (slogctx handler)
 ↓
extract attrs from ctx
 ↓
write log
```

## logger.InfoContext
```
logger.InfoContext(ctx, "logger - processing request")
```
wywołanie konkretnej instancji loggera
```
logger := ...
```
### Flow
```
ctx
 ↓
logger.InfoContext
 ↓
handler (slogctx handler)
 ↓
extract attrs from ctx
 ↓
write log
```

## slogctx.Info
```
slogctx.Info(ctx, "slogctx - processing request") // pobiera logger z kontekstu? - tylko tu będzie user_id 42?
```

Pobranie loggera z kontekstu
Np. 
```
slogctx.NewCtx(ctx, logger)
```
wstawia logger do kontekstu

Jeeli nie ma loggera w kontekście, pobiera domyślny logger ustawiony przez 
```
slog.setDefault
```

Pseudokod:
```
func Info(ctx context.Context, msg string, args ...any) {
    logger := slogctx.FromCtx(ctx)
    logger.InfoContext(ctx, msg, args...)
}
```

# logger.Info vs logger.InfoContext
logger.InfoContext - w handlerze loggera mogę mieć dostęp do zmiennych zawartych w kontekście
logger.Info - bez dostępu do kontekstu

# slogctx.NewCtx - zapisuje logger do kontekstu
```
ctx := slogctx.NewCtx(ctx, logger)

slogctx.Info(ctx, "user logged in",
    "user_id", userID,
)
```
Co się dzieje
Logger jest zapisany w context
slogctx.Info():
wyciąga logger z ctx
wywołuje logger.InfoContext(...)
Pseudo-kod:
```
func Info(ctx context.Context, msg string, args ...any) {
    logger := slogctx.FromCtx(ctx)
    logger.InfoContext(ctx, msg, args...)
}
```

# slogctx.Append
```
ctx := slogctx.Append(
    context.Background(),
    "request_id", "abc123",
)

slog.InfoContext(ctx, "processing request")
```
Append zapisuje dane w context.

# slogctx.Append vs slogctx.Prepend
```
ctx := slogctx.Prepend(context.Background(),
    "service", "auth-api",
)

ctx = slogctx.Append(ctx,
    "request_id", "abc123",
)

slog.InfoContext(ctx, "login attempt")
```

```
prepend attrs
logger attrs
log attrs
append attrs
```


# log.InfoContext a slogctx.Info
slogctx.Info - pobiera logger z kontekstu
chyba automatycznie wypisuje wartości, zawarte w kontekście

log.InfoContext - standardowy slog (co to oznacza?)
jezeli ma handler, który pobiera z kontekstu, to wypisze wartości w kontekście

# slog.New(...) vs slogctx.NewHandler(...)
## Pytanie
Czy loger utworzony przez :
```
	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)

logger := slog.New(handler)
```
odczytuje atrybuty dodane przez:
```
slogctx.Append
```
,
a logger utworzony przez :
```
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
``` 
tego nie robi?

## Odp.
Tak — dokładnie tak to działa.
Różnica polega na tym, czy handler potrafi odczytać atrybuty zapisane w context przez slogctx.

### 1. Logger z slogctx.NewHandler
```
handler := slogctx.NewHandler(
	slog.NewJSONHandler(os.Stdout, nil),
	nil,
)

logger := slog.New(handler)
```

Ten handler rozumie slog-context.

Czyli gdy wcześniej dodasz dane do context:
```
ctx := context.Background()
ctx = slogctx.Append(ctx, "user_id", 42)

logger.InfoContext(ctx, "user login")
```
to slogctx.NewHandler:
odczyta atrybuty z context
doda je do slog.Record
przekaże do JSONHandler

Output:
```
{
  "time": "...",
  "level": "INFO",
  "msg": "user login",
  "user_id": 42
}
```

### 2. Logger z samym JSONHandler
```
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

```
Ten handler nie zna slog-context.

Czyli:

```
ctx := context.Background()
ctx = slogctx.Append(ctx, "user_id", 42)

logger.InfoContext(ctx, "user login")
```

wynik:

```
{
  "time": "...",
  "level": "INFO",
  "msg": "user login"
}

```
user_id nie pojawi się, bo JSONHandler:
nie czyta danych z context
ignoruje je.

### 3. Dlaczego tak jest

Standardowy slog działa tak:
```
logger -> handler -> output
```

Handler dostaje:
```
Handle(ctx context.Context, record slog.Record)
```

Ale standardowe handlery (JSONHandler, TextHandler) nie używają ctx.

One patrzą tylko na:
```
record.Attrs
```

### 4. Co robi slogctx.NewHandler
slog-context tworzy wrapper handlera.

Schemat:
```
slogctx handler
        ↓
   JSON handler
        ↓
      stdout
```

Handler slogctx:

1. czyta atrybuty z context
2. dodaje je do record
3. przekazuje dalej.

Pseudo-kod:
```
func Handle(ctx context.Context, r slog.Record) {
	attrs := slogctx.Extract(ctx)

	for _, a := range attrs {
		r.AddAttrs(a)
	}

	next.Handle(ctx, r)
}
```

### 5. Dlaczego trzeba używać InfoContext

Bo tylko wtedy context trafia do handlera.
działa
```
logger.InfoContext(ctx, "msg")
```
nie działa
```
logger.Info("msg")
```

bo ctx nie jest przekazany.

### 6. Podsumowanie
logger	                                    czy czyta slogctx.Append
slog.New(JSONHandler)	                    ❌ nie
slog.New(slogctx.NewHandler(JSONHandler))	✅ tak
7. Typowy setup w aplikacji

Najczęściej robi się:

```
handler := slogctx.NewHandler(
	slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}),
	nil,
)

logger := slog.New(handler)

slog.SetDefault(logger)
```

Potem w kodzie:

```
ctx = slogctx.Append(ctx, "request_id", id)
slog.InfoContext(ctx, "request started")
```

### 7. Przepły slogctx.NewHandler(...)
```
log
 ↓
slogctx handler
 ↓
json handler
 ↓
stdout
```

# Własny handler pobierający z kontekstu
```
// === mytesthandler
type ctxMyKey string

const myIDKey ctxMyKey = "my_id"

type MyHandler struct {
	handler slog.Handler
}

func (h *MyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// wywoywany przez logger.InfoContext
func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
	// pobieramy my_id z context
	fmt.Println("MyHandler.Handle!")
	if myId := ctx.Value(myIDKey); myId != nil {
		r.AddAttrs(slog.Any("my_id", myId))
	}

	return h.handler.Handle(ctx, r)
}

func (h *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MyHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *MyHandler) WithGroup(name string) slog.Handler {
	return &MyHandler{handler: h.handler.WithGroup(name)}
}

func RunMyHandler() {
	base := slog.NewJSONHandler(os.Stdout, nil)
	handler := &MyHandler{handler: base}
	logger := slog.New(handler)
	ctx := context.WithValue(context.Background(), myIDKey, 44)
	logger.InfoContext(ctx, "user logged in")
}
```

# slog.New(handler).With(...)
Dodanie stałych pól do loggera
```
logger := slog.New(handler).With(
    "service", "billing-api",
)
```

# Grupowanie pól
```
logger.Info("order created",
    slog.Group("user",
        "id", 123,
        "email", "user@test.com",
    ),
)
```

# TODO - poziomy logów
```
opts := &slog.HandlerOptions{
	Level: slog.LevelDebug,
}

handler := slog.NewJSONHandler(os.Stdout, opts)

logger := slog.New(handler)

logger.Debug("cache miss")
logger.Info("user login")
```

# trace_id, span_id
```
handler := slogctx.NewHandler(
    slog.NewJSONHandler(os.Stdout, nil),
    &slogctx.HandlerOptions{
        Appenders: []slogctx.AttrExtractor{
            slogctx.ExtractAppended,
            slogotel.ExtractTraceSpanID,
        },
    },
)
```

```
handler := slogctx.NewHandler(
    slog.NewJSONHandler(os.Stdout, nil),
    &slogctx.HandlerOptions{
        Prependers: []slogctx.AttrExtractor{
            slogctx.ExtractPrepended,
        },
        Appenders: []slogctx.AttrExtractor{
            slogctx.ExtractAppended,
            slogotel.ExtractTraceSpanID,
        },
    },
)

logger := slog.New(handler)

slog.SetDefault(logger)
```

# slog.LogAttrs - jak logger.Info, tylko wydajniejsze
```
slog.LogAttrs(ctx, slog.LevelInfo, "hello",
    slog.String("user", "krzysztof"),
    slog.Int("id", 42),
)
```
jest równoznaczne
```
logger.Info("hello",
    "user", "krzysztof",
    "id", 42,
)
```
, ale bardziej wydajne

Typy:
```
slog.String	string
slog.Int	int
slog.Int64	int64
slog.Bool	bool
slog.Float64	float
slog.Time	time
slog.Duration	duration
slog.Any	dowolny typ
```

```
logger.Info("user login",
    "user", "krzysztof",
    "id", 42,
)
```

Najczęściej używane API (Info)
```
logger.Info("user login",
    "user", "krzysztof",
    "id", 42,
)
```

Tu przekazujesz:

...any

czyli wewnętrznie powstaje slice:
```
[]any{
    "user", "krzysztof",
    "id", 42,
}
```

Potem slog musi:

1️⃣ sparsować key/value
2️⃣ sprawdzić czy liczba argumentów jest parzysta
3️⃣ zamienić wszystko na Attr

czyli wewnętrznie powstaje:
```
[]Attr{
    {Key:"user", Value:"krzysztof"},
    {Key:"id", Value:42},
}
```

To oznacza:

konwersję
sprawdzanie
alokacje
2. LogAttrs – bardziej bezpośrednie API
```
logger.LogAttrs(ctx, slog.LevelInfo, "user login",
    slog.String("user", "krzysztof"),
    slog.Int("id", 42),
)
```

Tutaj przekazujesz:

```
[]slog.Attr
```

czyli struktury są już gotowe.

Nie ma:

parsowania
konwersji
sprawdzania

3. Co trafia do loggera
Info()
```
[]any
   ↓
parsowanie
   ↓
[]Attr
   ↓
handler
```

LogAttrs()
```
[]Attr
   ↓
handler
```

4. Benchmark z dokumentacji Go

Uproszczony benchmark:
```
func BenchmarkInfo(b *testing.B) {
    for i := 0; i < b.N; i++ {
        logger.Info("msg",
            "user", "krzysztof",
            "id", 42,
        )
    }
}

func BenchmarkLogAttrs(b *testing.B) {
    for i := 0; i < b.N; i++ {
        logger.LogAttrs(ctx, slog.LevelInfo, "msg",
            slog.String("user", "krzysztof"),
            slog.Int("id", 42),
        )
    }
}
```
Typowe wyniki:

Info        240 ns/op
LogAttrs    160 ns/op

czyli ~30-40% szybciej.

## Dlaczego slog.String istnieje
Bo tworzy typed attribute:
slog.String("user", "krzysztof")
czyli:
```
Attr{
    Key: "user",
    Value: Value{Kind: String}
}
```
To pozwala loggerowi:
szybciej serializować
uniknąć reflection.

## Ciekawostka: jeszcze szybszy pattern
W bardzo gorących ścieżkach można re-używać Attr.
```
userAttr := slog.String("user", "krzysztof")
logger.LogAttrs(ctx, slog.LevelInfo, "login", userAttr)
logger.LogAttrs(ctx, slog.LevelInfo, "logout", userAttr)
```

Zero nowych alokacji.

Jedna bardzo ważna rzecz (mało znana)

## Największy koszt logowania to nie Info() vs LogAttrs(), tylko:
string formatting
JSON serialization
I/O

czyli zapis do:

stdout
file
network

Dlatego w większości kodu używa się po prostu:

logger.Info(...)

# Wytłumacz funkcje
wytłumacz, co robią te funkcje:
```
func InjectRequest(ctx context.Context, r *http.Request) context.Context {
	requestID, _ := utils.GetRequestID(ctx)
	tenant, _ := utils.ExtractTenantID(ctx)

	return slogctx.With(ctx,
		slog.String("requestId", requestID),
		slog.String("tenantId", tenant),
		slog.Group("requestData",
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
		),
	)
}

func InjectTask(ctx context.Context, task *asynq.Task) context.Context {
	return slogctx.With(ctx, slog.String("taskType", task.Type()))
}

func InjectSystemEvent(
	ctx context.Context,
	event string,
) context.Context {
	return slogctx.With(ctx, slog.String("eventName", event))
}

func ErrorAttr(err error) slog.Attr {
	return slog.Attr{
		Key:   slogctx.ErrKey,
		Value: slog.StringValue(err.Error()),
	}
}

func Debug(ctx context.Context, msg string, args ...slog.Attr) {
	slogctx.LogAttrs(ctx, slog.LevelDebug, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...slog.Attr) {
	slogctx.LogAttrs(ctx, slog.LevelWarn, msg, args...)
}

func Info(ctx context.Context, msg string, args ...slog.Attr) {
	slogctx.LogAttrs(ctx, slog.LevelInfo, msg, args...)
}

func Error(ctx context.Context, msg string, err error, args ...slog.Attr) {
	args = append(args, slogctx.Err(err))

	slogctx.LogAttrs(ctx, slog.LevelError, msg, args...)
}
```

Ten kod to mała warstwa (wrapper) wokół slog + slog-context, która:

1️⃣ wstrzykuje dane do context (request, tenant, task, event)
2️⃣ upraszcza logowanie (Info, Debug, Error)
3️⃣ standaryzuje strukturę logów

To jest bardzo typowy wzorzec w większych backendach Go.

Rozbijmy wszystko krok po kroku.

## 1. InjectRequest
```
func InjectRequest(ctx context.Context, r *http.Request) context.Context {
	requestID, _ := utils.GetRequestID(ctx)
	tenant, _ := utils.ExtractTenantID(ctx)

	return slogctx.With(ctx,
		slog.String("requestId", requestID),
		slog.String("tenantId", tenant),
		slog.Group("requestData",
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
		),
	)
}
```

Co robi

Dodaje do context informacje o request HTTP, które potem automatycznie pojawią się w logach.

Dodawane pola:

requestId
tenantId
requestData.method
requestData.host
requestData.path
slogctx.With

To funkcja z biblioteki:

github.com/veqryn/slog-context

która:

dodaje slog.Attr do context

Te atrybuty są potem automatycznie dodawane do logów.

```
slog.Group
slog.Group("requestData",
	slog.String("method", r.Method),
	slog.String("host", r.Host),
	slog.String("path", r.URL.Path),
)
```

tworzy zagnieżdżoną strukturę loga.

Log JSON będzie wyglądał tak:

```
{
  "msg": "processing request",
  "requestId": "abc123",
  "tenantId": "tenant1",
  "requestData": {
    "method": "GET",
    "host": "api.example.com",
    "path": "/orders"
  }
}
```

## 2. InjectTask
```
func InjectTask(ctx context.Context, task *asynq.Task) context.Context {
	return slogctx.With(ctx, slog.String("taskType", task.Type()))
}
```

Dodaje do context:
taskType

czyli np.:

```
{
 "taskType": "send_email"
}
```

To jest przydatne w workerach backgroundowych (asynq).

## 3. InjectSystemEvent
```
func InjectSystemEvent(
	ctx context.Context,
	event string,
) context.Context {
	return slogctx.With(ctx, slog.String("eventName", event))
}
```

Dodaje do logów nazwę eventu systemowego:

eventName

np.

```
{
 "eventName": "user_registered"
}
```

## 4. ErrorAttr
```
func ErrorAttr(err error) slog.Attr {
	return slog.Attr{
		Key:   slogctx.ErrKey,
		Value: slog.StringValue(err.Error()),
	}
}
```

Tworzy atrybut błędu dla loggera.

slogctx.ErrKey to standardowy klucz (np. "error").

Czyli:

ErrorAttr(err)

zwraca:

```
{
 "error": "database connection failed"
}
```

## 5. Debug
```
func Debug(ctx context.Context, msg string, args ...slog.Attr) {
	slogctx.LogAttrs(ctx, slog.LevelDebug, msg, args...)
}
```

Wrapper na:

```
slogctx.LogAttrs
```

czyli logowanie na poziomie DEBUG.

Użycie:

```
Debug(ctx, "user loaded",
	slog.String("userId", id),
)
```

## 6. Warn
```
func Warn(ctx context.Context, msg string, args ...slog.Attr) {
	slogctx.LogAttrs(ctx, slog.LevelWarn, msg, args...)
}
```
Log poziomu:
WARN

## 7. Info
```
func Info(ctx context.Context, msg string, args ...slog.Attr) {
	slogctx.LogAttrs(ctx, slog.LevelInfo, msg, args...)
}
```
Wrapper na:
INFO log
czyli zamiast pisać:

```
slogctx.LogAttrs(ctx, slog.LevelInfo, ...)
```

piszesz:

```
Info(ctx, ...)
```
## 8. Error
```
func Error(ctx context.Context, msg string, err error, args ...slog.Attr) {
	args = append(args, slogctx.Err(err))

	slogctx.LogAttrs(ctx, slog.LevelError, msg, args...)
}
```

Ta funkcja robi coś ważnego.

### 1️⃣ dodaje error do loga
```
slogctx.Err(err)
```

czyli:

```
{
 "error": "database timeout"
}
```
### 2️⃣ loguje na poziomie
ERROR
użycie
```
Error(ctx, "failed to load user", err,
	slog.String("userId", id),
)
```

log:

```
{
 "level": "ERROR",
 "msg": "failed to load user",
 "userId": "123",
 "error": "user not found"
}
```

## 9. Jak to działa razem

Przykład w aplikacji:
```
ctx := InjectRequest(ctx, r)
Info(ctx, "processing request")
```

Log:

```
{
 "msg": "processing request",
 "requestId": "abc123",
 "tenantId": "tenant1",
 "requestData": {
   "method": "GET",
   "path": "/orders"
 }
}
```

A w workerze:

```
ctx := InjectTask(ctx, task)
Info(ctx, "task started")
```

log:

```
{
 "msg": "task started",
 "taskType": "send_email"
}
```

## 10. Dlaczego to jest bardzo dobry pattern

Ten kod robi 3 ważne rzeczy:

### 1️⃣ centralizuje logowanie

zamiast:

slogctx.LogAttrs(...)

wszędzie

masz:

Info(...)
Error(...)
### 2️⃣ automatycznie dodaje kontekst

np:

requestId
tenantId
taskType
eventName
### 3️⃣ wymusza strukturę logów

wszystkie logi mają te same pola.

## 11. Jak wygląda log końcowy
```
{
 "time": "2026-03-06T11:20:00Z",
 "level": "INFO",
 "msg": "task started",
 "requestId": "abc123",
 "tenantId": "tenant1",
 "taskType": "send_email",
 "requestData": {
   "method": "POST",
   "path": "/orders"
 }
}
```

## ✅ Najkrócej
Ten plik to:
logging helper / logging facade
który:
dodaje metadata do context
upraszcza logowanie
wymusza strukturę logów.