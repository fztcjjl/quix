package main

import (
	"context"
	"log/slog"
	"os"

	quix "github.com/fztcjjl/quix"
	"github.com/fztcjjl/quix/core/logger"
)

func main() {
	// 方式 1: 使用默认 slog logger（零配置）
	app := quix.New()
	log := app.Logger()

	ctx := context.Background()

	log.Info(ctx, "使用默认 slog logger")
	log.Info(ctx, "带字段的日志", "method", "GET", "path", "/users", "status", 200)

	// 方式 2: 自定义 slog logger
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	customSL := slog.New(handler)
	app2 := quix.New(quix.WithLogger(logger.NewSlog(customSL)))
	log2 := app2.Logger()

	// With 追加公共字段
	reqLogger := log2.With("service", "quix", "version", "1.0.0")
	reqLogger.Info(ctx, "带公共字段的请求日志", "method", "POST", "path", "/users")
	reqLogger.Error(ctx, "发生错误", "err", "connection refused")
}
