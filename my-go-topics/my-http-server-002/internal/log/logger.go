package log

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/krzysztofkolcz/my-http-server-002/internal/config"
	"github.com/krzysztofkolcz/my-http-server-002/internal/constants"
	"github.com/krzysztofkolcz/my-http-server-002/utils"
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
