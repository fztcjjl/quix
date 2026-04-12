package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/fztcjjl/quix/core/log"
)

func main() {
	ctx := context.Background()

	// 使用自定义 slog logger 替换全局默认
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	log.SetDefault(log.NewSlog(slog.New(handler)))

	log.Info(ctx, "使用 slog logger")
	log.Info(ctx, "带字段的日志", "method", "GET", "path", "/users", "status", 200)

	// With 追加公共字段
	reqLogger := log.With("service", "quix", "version", "1.0.0")
	reqLogger.Info(ctx, "带公共字段的请求日志", "method", "POST", "path", "/users")
	reqLogger.Error(ctx, "发生错误", "err", "connection refused")
}
