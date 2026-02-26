package daemon

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
	"github.com/krzysztofkolcz/my-http-server/internal/config"
	"github.com/krzysztofkolcz/my-http-server/internal/constants"
	"github.com/krzysztofkolcz/my-http-server/internal/controllers/myhttpcontroller"
	"github.com/krzysztofkolcz/my-http-server/internal/handlers"
	logger "github.com/krzysztofkolcz/my-http-server/internal/log"
	"github.com/krzysztofkolcz/my-http-server/internal/middleware"
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

type MyHttpServer struct {
	controller *myhttpcontroller.APIController
	server     *http.Server
}

type Server interface {
	Start(ctx context.Context) error
	Close() error
}

func NewMyHttpServer(
	ctx context.Context,
	cfg *config.Config,
) (*MyHttpServer, error) {
	controller := myhttpcontroller.NewAPIController(ctx)

	httpServer, err := createHTTPServer(cfg, controller)
	if err != nil {
		return nil, oops.In(ServerLogDomain).Wrapf(err, "creating http server")
	}

	return &MyHttpServer{
		controller: controller,
		server:     httpServer,
	}, nil
}

func (s *MyHttpServer) Close(ctx context.Context) error {
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

func (s *MyHttpServer) Start(ctx context.Context) error {
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
	swagger, err := myhttpserver.GetSwagger()
	if err != nil {
		// 	return nil, errs.Wrapf(err, "failed to load swagger file")
	}
	// Instead of setting Servers list to nil, we only remove the host from the URL.
	// This is because gorilla/mux used by the OAPI validator only allows hosts
	// without periods '.' in the URL. However, we still want to keep
	// the rest of the Server URL to allow matching path prefix with parameterised tenants.
	for _, srv := range swagger.Servers {
		srv.URL = strings.Replace(srv.URL, "{host}", "", 1)
	}

	return swagger, nil
}

func createHTTPServer(
	cfg *config.Config,
	ctr *myhttpcontroller.APIController,
) (*http.Server, error) {
	swagger, err := SetupSwagger()
	if err != nil {
		return nil, oops.In(ServerLogDomain).Wrapf(err, "setup swagger")
	}

	// Middlewares run in a FILO. Last middleware on the slice is the first one ran
	// First middleware to run should be the InjectRequestID

	httpHandler := myhttpserver.HandlerWithOptions(
		myhttpserver.NewStrictHandlerWithOptions(
			ctr,
			[]myhttpserver.StrictMiddlewareFunc{},
			myhttpserver.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  handlers.RequestErrorHandlerFunc(),
				ResponseErrorHandlerFunc: handlers.ResponseErrorHandlerFunc(),
			},
		),
		myhttpserver.StdHTTPServerOptions{
			BaseURL:    constants.BasePath,
			BaseRouter: NewServeMux(constants.BasePath),
			// ErrorHandlerFunc: handlers.ParamsErrorHandler(),
			Middlewares: []myhttpserver.MiddlewareFunc{ // Middlewares are applied from last to first
				middleware.OAPIMiddleware(swagger),
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
