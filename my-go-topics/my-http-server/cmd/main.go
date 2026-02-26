package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/krzysztofkolcz/my-http-server/internal/config"
	"github.com/krzysztofkolcz/my-http-server/internal/constants"
	"github.com/krzysztofkolcz/my-http-server/internal/daemon"
	logger "github.com/krzysztofkolcz/my-http-server/internal/log"
	"github.com/krzysztofkolcz/my-http-server/internal/runner"
	"github.com/samber/oops"
)

var (
	BuildInfo               = "{}"
	gracefulShutdownSec     = flag.Int64("graceful-shutdown", 1, "graceful shutdown seconds")
	gracefulShutdownMessage = flag.String("graceful-shutdown-message", "Graceful shutdown in %d seconds",
		"graceful shutdown message")
)

// func main() {

// 	ctx := context.Background()

// 	c := myhttpcontroller.NewAPIController(ctx)

// 	strictHandler := myhttpserver.NewStrictHandlerWithOptions(c)

// 	server := http.Server{
// 		Addr:    ":6767",
// 		Handler: myhttpserver.Handler(strictHandler),
// 	}

// 	log.Println("Server listening on :6767")

// 	log.Fatal(server.ListenAndServe())
// }

// main is the entry point for the application. It is intentionally kept small
// because it is hard to test, which would lower test coverage.
func main() {
	flag.Parse()

	exitCode := runner.RunFuncWithSignalHandling(run, runner.RunFlags{
		GracefulShutdownSec:     *gracefulShutdownSec,
		GracefulShutdownMessage: *gracefulShutdownMessage,
		Env:                     constants.APIName,
	})
	os.Exit(exitCode)
}

func run(ctx context.Context, cfg *config.Config) error {
	logger.InitLogger(*cfg)

	logger.Debug(ctx, "Starting the application", slog.Any("config", cfg))

	// Create and start CMK Server
	s, err := daemon.NewMyHttpServer(ctx, cfg)
	if err != nil {
		return oops.In("main").Wrapf(err, "creating api server")
	}

	err = s.Start(ctx)
	if err != nil {
		return oops.In("main").Wrapf(err, "starting api server")
	}

	logger.Info(ctx, "API Server has started")

	<-ctx.Done()

	err = s.Close(ctx)
	if err != nil {
		return oops.In("main").Wrapf(err, "closing server")
	}

	return nil
}
