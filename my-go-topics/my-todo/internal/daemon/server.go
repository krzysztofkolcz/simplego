package daemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"time"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/config"
	"github.com/C5383717/my-todo/internal/constants"
	"github.com/C5383717/my-todo/internal/errs"
	"github.com/C5383717/my-todo/internal/handlers"
	httphandlers "github.com/C5383717/my-todo/internal/http/handlers"
	logger "github.com/C5383717/my-todo/internal/log"
	"github.com/C5383717/my-todo/internal/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/oops"
)

const (
	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 10 * time.Second
	WriteTimeout      = 10 * time.Second
	IdleTimeout       = 120 * time.Second
	ServerLogDomain   = "server daemon"
	ShutdownTimeout   = 120 * time.Second
)

// TodoServer is the HTTP server for the TODO API.
type TodoServer struct {
	server *http.Server
}

// NewTodoServer wires the HTTP server.
// It receives the command and query buses — it does not know about repositories
// or domain types. The bus is the boundary between HTTP and application layers.
func NewTodoServer(
	ctx context.Context,
	cfg *config.Config,
	cmdBus bus.CommandBus,
	qryBus bus.QueryBus,
) (*TodoServer, error) {
	httpServer, err := createHTTPServer(cfg, cmdBus, qryBus)
	if err != nil {
		return nil, oops.In(ServerLogDomain).Wrapf(err, "creating http server")
	}

	return &TodoServer{server: httpServer}, nil
}

func (s *TodoServer) Start(ctx context.Context) error {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "server encountered an error", err)

			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}
	}()

	return nil
}

func (s *TodoServer) Close(ctx context.Context) error {
	shutdownCtx, shutdownRelease := context.WithTimeout(ctx, ShutdownTimeout)
	defer shutdownRelease()

	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		return oops.In("HTTP Server").
			WithContext(ctx).
			Wrapf(err, "Failed shutting down HTTP server")
	}

	logger.Info(ctx, "Completed graceful shutdown of HTTP server")

	return nil
}

func SetupSwagger() (*openapi3.T, error) {
	swagger, err := todoapi.GetSwagger()
	if err != nil {
		return nil, errs.Wrapf(err, "failed to load swagger file")
	}
	// Strip the host variable from server URLs so the OAPI validator does not
	// require a Host header that matches the spec template exactly.
	for _, srv := range swagger.Servers {
		fmt.Printf("Before: srv.URL: %s", srv.URL)
		fmt.Println()
		srv.URL = strings.Replace(srv.URL, "{host}", "", 1)
		fmt.Printf("After: srv.URL: %s", srv.URL)
		fmt.Println()
	}

	return swagger, nil
}

// createHTTPServer builds the http.Server with the full middleware stack.
//
// Middleware execution order (FILO — last item in slice runs first):
//  1. InjectRequestID   — runs first: all subsequent middleware see a requestID
//  2. PanicRecovery     — wraps everything so panics never escape
//  3. LoggingMiddleware — logs each request with method, path, status, duration
//  4. OAPIMiddleware    — validates request body/params against the OpenAPI spec
func createHTTPServer(
	cfg *config.Config,
	cmdBus bus.CommandBus,
	qryBus bus.QueryBus,
) (*http.Server, error) {
	swagger, err := SetupSwagger()
	if err != nil {
		return nil, oops.In(ServerLogDomain).Wrapf(err, "setup swagger")
	}

	todoHandler := httphandlers.NewTodoHandler(cmdBus, qryBus)

	httpHandler := todoapi.HandlerWithOptions(
		todoapi.NewStrictHandlerWithOptions(
			todoHandler,
			[]todoapi.StrictMiddlewareFunc{},
			todoapi.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  handlers.RequestErrorHandlerFunc(),
				ResponseErrorHandlerFunc: handlers.ResponseErrorHandlerFunc(),
			},
		),
		todoapi.StdHTTPServerOptions{
			BaseURL:          constants.BasePath,
			BaseRouter:       NewServeMux(constants.BasePath),
			ErrorHandlerFunc: handlers.ParamsErrorHandler(),
			Middlewares: []todoapi.MiddlewareFunc{ // applied FILO — last runs first
				middleware.OAPIMiddleware(swagger),
				middleware.LoggingMiddleware(),
				middleware.PanicRecoveryMiddleware(),
				middleware.InjectRequestID(),
			},
		},
	)

	return &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           httpHandler,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
	}, nil
}
