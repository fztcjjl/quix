package main

import (
	"context"

	quix "github.com/fztcjjl/quix"
	"github.com/fztcjjl/quix/core/logger"
	"go.uber.org/zap"
)

func main() {
	// 使用 zap 作为日志实现
	rawLog, _ := zap.NewProduction()
	defer rawLog.Sync()
	sl := rawLog.Sugar()

	app := quix.New(quix.WithLogger(logger.NewZap(sl)))
	log := app.Logger()

	ctx := context.Background()

	log.Info(ctx, "使用 Zap logger")
	log.Info(ctx, "带字段的日志", "method", "GET", "path", "/users")

	// With 追加公共字段
	reqLogger := log.With("service", "quix")
	reqLogger.Info(ctx, "请求处理完成", "status", 200)
	reqLogger.Error(ctx, "请求失败", "err", "connection refused")
}
