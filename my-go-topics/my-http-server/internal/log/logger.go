package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/krzysztofkolcz/my-http-server/internal/config"
	"github.com/krzysztofkolcz/my-http-server/internal/constants"
	slogctx "github.com/veqryn/slog-context"
)

var (
	logLevel = new(slog.LevelVar)
)

func setLogLevels(cfg config.Config) {
	setLogLevel(logLevel, cfg.LogLevel)
}

func setLogLevel(levelVar *slog.LevelVar, level string) {
	switch strings.ToLower(level) {
	case constants.LogLevelDebug.String():
		levelVar.Set(slog.LevelDebug)
	case constants.LogLevelInfo.String():
		levelVar.Set(slog.LevelInfo)
	case constants.LogLevelWarn.String():
		levelVar.Set(slog.LevelWarn)
	case constants.LogLevelError.String():
		levelVar.Set(slog.LevelError)
	default:
		levelVar.Set(slog.LevelInfo)
	}
}

func InitLogger(cfg config.Config) *slog.Logger {
	setLogLevels(cfg)

	baseHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true, // przydatne w dev
		},
	)

	handler := slogctx.NewHandler(
		baseHandler,
		&slogctx.HandlerOptions{
			Prependers: []slogctx.AttrExtractor{
				slogctx.ExtractPrepended,
			},
		},
	)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
