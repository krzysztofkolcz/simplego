package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/C5383717/my-todo/internal/config"
	logger "github.com/C5383717/my-todo/internal/log"
)

type RunFlags struct {
	GracefulShutdownSec     int64
	GracefulShutdownMessage string
	Env                     string
}

// RunFuncWithSignalHandling runs the given function with signal handling. When
// a CTRL-C is received, the context will be cancelled on which the function can
// act upon. It returns the exit code.
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

	err = f(ctx, cfg)
	if err != nil {
		logger.Error(ctx, "Failed to start the application", err)
		_, _ = fmt.Fprintln(os.Stderr, err)

		return 1
	}

	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(runFlags.GracefulShutdownMessage, runFlags.GracefulShutdownSec))
	time.Sleep(time.Duration(runFlags.GracefulShutdownSec) * time.Second)

	return 0
}
