package daemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/krzysztofkolcz/my-http-server-002/internal/api/myhttpserver2"
	"github.com/krzysztofkolcz/my-http-server-002/internal/config"
	"github.com/krzysztofkolcz/my-http-server-002/internal/constants"
	"github.com/krzysztofkolcz/my-http-server-002/internal/controllers/myhttpcontroller2"
	"github.com/krzysztofkolcz/my-http-server-002/internal/errs"
	"github.com/krzysztofkolcz/my-http-server-002/internal/handlers"
	logger "github.com/krzysztofkolcz/my-http-server-002/internal/log"
	"github.com/krzysztofkolcz/my-http-server-002/internal/middleware"
	"github.com/samber/oops"
)

const (
	ReadHeaderTimeout    = 5 * time.Second
	ReadTimeout          = 10 * time.Second
	WriteTimeout         = 10 * time.Second
	IdleTimeout          = 120 * time.Second
	ServerLogDomain      = "server daemon"
	AuthzRefreshInterval = 120 * time.Second
	ShutdownTimeout      = 120 * time.Second
)

type MyHttpServer2 struct {
	controller *myhttpcontroller2.APIController
	server     *http.Server
}

type Server interface {
	Start(ctx context.Context) error
	Close() error
}

func NewMyHttpServer(
	ctx context.Context,
	cfg *config.Config,
) (*MyHttpServer2, error) {
	controller := myhttpcontroller2.NewAPIController(ctx)

	httpServer, err := createHTTPServer(cfg, controller)
	if err != nil {
		return nil, oops.In(ServerLogDomain).Wrapf(err, "creating http server")
	}

	return &MyHttpServer2{
		controller: controller,
		server:     httpServer,
	}, nil
}

func (s *MyHttpServer2) Close(ctx context.Context) error {
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

func (s *MyHttpServer2) Start(ctx context.Context) error {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "server encountered an error", err)

			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}
	}()

	return nil
}

func SetupSwagger() (*openapi3.T, error) {
	swagger, err := myhttpserver2.GetSwagger()
	if err != nil {
		return nil, errs.Wrapf(err, "failed to load swagger file")
	}
	for _, srv := range swagger.Servers {
		fmt.Printf("Before: srv.URL: %s", srv.URL)
		fmt.Println()
		srv.URL = strings.Replace(srv.URL, "{host}", "", 1)
		fmt.Printf("After: srv.URL: %s", srv.URL)
		fmt.Println()
	}

	return swagger, nil
}

/*
https://chatgpt.com/g/g-p-6979069f038081918633e25bb9943f89-nauka-golanga/c/69a165f1-3778-8329-bd68-f71eedd843ed

Jak przepływa request?

Cały flow:

HTTP Request

	↓

InjectRequestID middleware

	↓

OAPI Request Validator middleware

	↓

Param binding

	↓

Strict JSON decode

	↓

StrictMiddlewareFunc (jeśli są)

	↓

Controller

	↓

Response validation

	↓

ResponseErrorHandler

	↓

write.ErrorResponse
*/
func createHTTPServer(
	cfg *config.Config,
	ctr *myhttpcontroller2.APIController,
) (*http.Server, error) {
	swagger, err := SetupSwagger()
	if err != nil {
		return nil, oops.In(ServerLogDomain).Wrapf(err, "setup swagger")
	}

	// Middlewares run in a FILO. Last middleware on the slice is the first one ran
	// First middleware to run should be the InjectRequestID

	httpHandler := myhttpserver2.HandlerWithOptions(
		myhttpserver2.NewStrictHandlerWithOptions(
			ctr,
			[]myhttpserver2.StrictMiddlewareFunc{},
			myhttpserver2.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  handlers.RequestErrorHandlerFunc(),
				ResponseErrorHandlerFunc: handlers.ResponseErrorHandlerFunc(),
			},
		),
		myhttpserver2.StdHTTPServerOptions{
			BaseURL:          constants.BasePath,
			BaseRouter:       NewServeMux(constants.BasePath),
			ErrorHandlerFunc: handlers.ParamsErrorHandler(),
			Middlewares: []myhttpserver2.MiddlewareFunc{ // Middlewares are applied from last to first
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
