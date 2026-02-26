package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krzysztofkolcz/my-http-server/internal/config"
	logger "github.com/krzysztofkolcz/my-http-server/internal/log"
)

type RunFlags struct {
	GracefulShutdownSec     int64
	GracefulShutdownMessage string
	Env                     string
}

// RunFuncWithSignalHandling runs the given function with signal handling. When
// a CTRL-C is received, the context will be cancelled on which the function can
// act upon.
// It returns the exitCode
func RunFuncWithSignalHandling(f func(context.Context, *config.Config) error, runFlags RunFlags) int {
	ctx, cancelOnSignal := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer cancelOnSignal()

	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error(ctx, "Failed to load the configuration", err)
		_, _ = fmt.Fprintln(os.Stderr, err)

		return 1
	}

	// log.Debug(ctx, "Starting the application", slog.Any("config", *cfg))

	err = f(ctx, cfg)
	if err != nil {
		logger.Error(ctx, "Failed to start the application", err)
		_, _ = fmt.Fprintln(os.Stderr, err)

		return 1
	}

	// graceful shutdown so running goroutines may finish
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(runFlags.GracefulShutdownMessage, runFlags.GracefulShutdownSec))
	time.Sleep(time.Duration(runFlags.GracefulShutdownSec) * time.Second)

	return 0
}
