package main

import (
	"context"
	"os"

	"github.com/fztcjjl/quix/core/log"
	"github.com/rs/zerolog"
)

func main() {
	ctx := context.Background()

	// 使用 zerolog 替换全局默认
	l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().Timestamp().Logger()
	log.SetDefault(log.NewZerolog(l))

	log.Info(ctx, "使用 Zerolog logger")
	log.Info(ctx, "带字段的日志", "method", "GET", "path", "/users")

	// With 追加公共字段
	reqLogger := log.With("service", "quix")
	reqLogger.Info(ctx, "请求处理完成", "status", 200, "duration_ms", 42)
	reqLogger.Error(ctx, "请求失败", "status", 500, "err", "internal error")
}
