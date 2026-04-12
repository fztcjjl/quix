package main

import (
	"context"

	"github.com/fztcjjl/quix/core/log"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// 使用 zap 替换全局默认
	rawLog, _ := zap.NewProduction()
	defer rawLog.Sync()
	log.SetDefault(log.NewZap(rawLog.Sugar()))

	log.Info(ctx, "使用 Zap logger")
	log.Info(ctx, "带字段的日志", "method", "GET", "path", "/users")

	// With 追加公共字段
	reqLogger := log.With("service", "quix")
	reqLogger.Info(ctx, "请求处理完成", "status", 200)
	reqLogger.Error(ctx, "请求失败", "err", "connection refused")
}
