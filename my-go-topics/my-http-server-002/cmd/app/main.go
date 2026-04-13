package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/krzysztofkolcz/my-http-server-002/internal/config"
	"github.com/krzysztofkolcz/my-http-server-002/internal/constants"
	deamon "github.com/krzysztofkolcz/my-http-server-002/internal/deamon"
	logger "github.com/krzysztofkolcz/my-http-server-002/internal/log"
	"github.com/krzysztofkolcz/my-http-server-002/internal/runner"
	"github.com/samber/oops"
)

var (
	BuildInfo               = "{}"
	gracefulShutdownSec     = flag.Int64("graceful-shutdown", 1, "graceful shutdown seconds")
	gracefulShutdownMessage = flag.String("graceful-shutdown-message", "Graceful shutdown in %d seconds",
		"graceful shutdown message")
)

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

	s, err := deamon.NewMyHttpServer(ctx, cfg)
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
