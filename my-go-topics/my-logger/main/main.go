package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	slogctx "github.com/veqryn/slog-context"
)

func main() {

	// runLogExample()
	// RunUserHandler()
	// RunMyHandler()
	// RunSlogctx2()
	// RunSlogctx3()
	RunSlogctx4()
	// RunSlogctxAppend()
}

func runLogExample() {
	fmt.Println("=== slog.Info")
	slog.Info("Application started")

	slog.Info("user logged in",
		"user_id", 123,
		"email", "user@example.com",
	)
	fmt.Println()

	fmt.Println("=== logger1 text handler:")
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	logger.Info("app started")
	fmt.Println()

	fmt.Println("=== logger2 json handler:")
	handler2 := slog.NewJSONHandler(os.Stdout, nil)
	logger2 := slog.New(handler2)
	logger2.Info("user created",
		"user_id", 42,
	)
	fmt.Println()

	opts3 := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler3 := slog.NewJSONHandler(os.Stdout, opts3)

	logger3 := slog.New(handler3)

	fmt.Println("=== logger3 json handler with LevelDebug:")
	logger3.Debug("cache miss")
	logger3.Info("user login")
	fmt.Println()

	fmt.Println("=== logger4 json handler with With('service') option field:")
	handler4 := slog.NewJSONHandler(os.Stdout, nil)
	logger4 := slog.New(handler4).With( // dodaje stałe pola loggera
		"service", "auth-api",
	)
	logger4.Info("login attempt")
	fmt.Println()

	fmt.Println("=== logger5 With('user_id):")
	userLogger5 := logger.With(
		"user_id", 42,
	)
	userLogger5.Info("user updated profile")
	fmt.Println()

	fmt.Println("=== logger4 logger4.Eror:")
	err := errors.New("database connection failed")
	logger4.Error("failed to save user",
		"error", err,
		"user_id", 42,
	)
	fmt.Println()

	fmt.Println("=== logger4 log.InfoContext:")
	ctx := context.Background()
	httpHandler(ctx, logger4)
	fmt.Println()

	fmt.Println("=== logger4 - tu nie wypisze request_id:")
	// tu nie wypisze request_id
	ctx2 := context.WithValue(ctx, "request_id", uuid.New().String())
	httpHandler(ctx2, logger4)
	fmt.Println()

	// dopiero tu wypisze
	// w sumie to po prostu pobranie z kontekstu i log.With(...)
	fmt.Println("=== logger4 - tu wypisze request_id")
	fmt.Println("=== logger4 - ale to po prostu pobranie wartosci z kontekstu i dodanie do loga:")
	httpHandlerMock2(ctx2, logger4)
	fmt.Println()

	fmt.Println("=== logger4 - group")
	logger4.Info("logger4: order created",
		slog.Group("user",
			"id", 42,
			"email", "user@test.com",
		),
	)
	fmt.Println()
}

func httpHandler(ctx context.Context, log *slog.Logger) {
	log.InfoContext(ctx, "processing request")
}

func httpHandlerMock2(ctx context.Context, logger *slog.Logger) {
	log := LoggerFromContext(ctx, logger)
	log.Info("request received")
}

func LoggerFromContext(ctx context.Context, log *slog.Logger) *slog.Logger {
	if id, ok := ctx.Value("request_id").(string); ok {
		return log.With("request_id", id)
	}
	return log
}

// === user
type ctxKey string

const userIDKey ctxKey = "user_id"

type UserHandler struct {
	handler slog.Handler
}

func (h *UserHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *UserHandler) Handle(ctx context.Context, r slog.Record) error {
	// pobieramy user_id z context
	if userID := ctx.Value(userIDKey); userID != nil {
		r.AddAttrs(slog.Any("user_id", userID))
	}

	return h.handler.Handle(ctx, r)
}

func (h *UserHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &UserHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *UserHandler) WithGroup(name string) slog.Handler {
	return &UserHandler{handler: h.handler.WithGroup(name)}
}

func RunUserHandler() {

	// zwykły handler JSON
	base := slog.NewJSONHandler(os.Stdout, nil)

	// nasz handler z obsługą context
	handler := &UserHandler{handler: base}

	logger := slog.New(handler)

	// tworzymy context z user_id
	ctx := context.WithValue(context.Background(), userIDKey, 42)

	// log z context
	// wywołana metoda Handle UerHandlera
	logger.InfoContext(ctx, "user logged in")

	// log bez context
	logger.Info("something happened")
}

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

// === slogctx
// https://chatgpt.com/c/69a8038c-1498-8329-9b71-c6bb3fc87bba
func RunSlogctx() {
	ctx := context.Background()
	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)
	logger := slog.New(handler)

	logger.InfoContext(ctx, "some info")
	// slog.SetDefault(logger)
}

func RunSlogctx2() {
	ctx := context.Background()
	// slogctx handler
	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)
	// logger zawiera slogctx handler
	logger := slog.New(handler)

	// dodaje request_id do ctx
	ctx = slogctx.Append(
		ctx,
		"request_id", "abc123",
	)

	// ustawiam default logger
	slog.SetDefault(logger)

	// ponizsza funkcja uzywa globalnego loggera ustawionego przez slog.SetDefault
	slog.InfoContext(ctx, "slog default logger - processing request")

	// konkretna instancja loggera, logger przekazany jawnie
	logger.InfoContext(ctx, "logger - processing request")

	// probuje pobrac logger z kontekstu.
	// func Info(ctx context.Context, msg string, args ...any) {
	// 		logger := slogctx.FromCtx(ctx)
	// 		logger.InfoContext(ctx, msg, args...)
	// }
	// ale ja nie dodalem loggera.
	// nie zrobilem:
	// ctx = slogctx.NewCtx(ctx, logger)
	// wiec ctx nie zawiera loggerra
	// wiec fallback do slog.Default
	slogctx.Info(ctx, "slogctx - processing request") // pobiera logger z kontekstu? Ale nie dodawałem loggera do kontekstu?

	// rozumiem, ze tworze tu kontekst z loggerem?
	ctx = slogctx.With(
		ctx,
		"user_id", 42,
	)

	// pytanie, dlaczego tutaj nie wypisze user_id, skoro domyślny logger ma handler slogctx?
	slog.InfoContext(ctx, "slog default logger - processing request")

	// pytanie, dlaczego tutaj nie wypisze user_id, skoro domyślny logger ma handler slogctx?
	logger.InfoContext(ctx, "logger - processing request")

	// pobiera logger z kontekstu? - tylko tu będzie user_id 42?
	slogctx.Info(ctx, "slogctx - processing request")

}

// Porownanie z RunSlogctx4
// Tutaj nie ma extraktorow
func RunSlogctx3() {
	ctx := context.Background()
	// slogctx handler
	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil, // brak extraktorow, handler nie czyta context
	)
	// logger zawiera slogctx handler
	logger := slog.New(handler)

	// ustawiam default logger
	slog.SetDefault(logger)

	// rozumiem, ze tworze tu nowy kontekst z loggerem?
	// zapisuje dane w context ALE pod specjalnym kluczem slogctx To jeszcze NIC nie loguje.
	// tworzy NOWY context z tymi danymi
	// nadpisuje poprzednie wartości
	// Ale w praktyce częściej używa się: slogctx.Append(ctx, ...) Bo: można dokładać kolejne wartości
	// działa jak middleware chain
	ctx = slogctx.With(
		ctx,
		"user_id", 42,
	)

	// pytanie, dlaczego tutaj nie wypisze user_id, skoro domyślny logger ma handler slogctx?
	// Bo robi: robi: Logger -> Handler -> JSON
	// A Twój handler: NIE ma extractorów, NIE patrzy do context,  NIE wie o user_id
	slog.InfoContext(ctx, "slog.InfoContext")

	// pytanie, dlaczego tutaj nie wypisze user_id, skoro domyślny logger ma handler slogctx?
	logger.InfoContext(ctx, "logger.InfoContext")

	// pobiera logger z kontekstu? - tylko tu będzie user_id 42?
	// Dlaczego slogctx.Info(...) działa?
	// To jest zupełnie inna funkcja.
	// slogctx.Info(ctx, "...")
	// robi coś takiego:
	// wyciąga attrs z context ręcznie
	// robi:
	// slog.InfoContext(ctx, msg, extractedAttrs...)
	// Czyli:
	//  omija handler
	// sam dokleja dane
	slogctx.Info(ctx, "slogctx.Info")
}

func RunSlogctx4() {
	ctx := context.Background()

	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		&slogctx.HandlerOptions{ // extraktor
			Prependers: []slogctx.AttrExtractor{
				slogctx.ExtractPrepended,
			},
		},
	)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx = slogctx.Append(ctx, "user_id", 42)

	slog.InfoContext(ctx, "slog.InfoContext")
	logger.InfoContext(ctx, "logger.InfoContext")
	slogctx.Info(ctx, "slogctx.Info")
}

// slogctx.Append
func RunSlogctxAppend() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx = slogctx.Append(
		ctx,
		"request_id", "abc123",
	)

	logger.Info("logger.Info")                    // Nie wypisuje request_id
	logger.InfoContext(ctx, "logger.InfoContext") // Nie wypisuje request_id - brak handlera, który pobiera z kontekstu

	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)
	loggerWithSlogctx := slog.New(handler)
	loggerWithSlogctx.Info("loggerWithSlogctx.Info")                                    // Nie wypisuje request_id
	loggerWithSlogctx.InfoContext(ctx, "loggerWithSlogctx.InfoContext", "key", "value") // Wypisuje request_id

}

/*
ctx = slogctx.Append(
	ctx,
	"request_id", "abc123",
)

ctx := slogctx.Append(ctx,
    slog.String("request_id", "abc123"),
)

*/
