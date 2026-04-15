package main

import (
	"context"
	"os"

	"github.com/fztcjjl/quix/core/log"
	"github.com/rs/zerolog"
)

func main() {
	ctx := context.Background()

	// 1. 创建 Logger：zerolog ConsoleWriter + 时间戳
	l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().Timestamp().Logger()
	zl := log.NewZerolog(l)
	defer zl.Close()

	// 2. 设置为全局默认
	log.SetDefault(zl)

	// 3. 各级别日志
	log.Info(ctx, "应用启动")
	log.Warn(ctx, "磁盘空间不足", "free", "9.1GB", "threshold", "100GB")
	log.Error(ctx, "数据库连接失败", "err", "connection refused")

	// 4. key-value 字段
	log.Info(ctx, "处理请求", "method", "GET", "path", "/users", "status", 200)

	// 5. 非字符串 key 自动转为 key_0、key_1
	log.Info(ctx, "带非字符串 key", 123, "value", true, "flag")

	// 6. 奇数尾部 key 被静默丢弃
	log.Info(ctx, "奇数参数", "key1", "val1", "orphan_key")

	// 7. With 创建子 logger，追加公共字段
	reqLogger := log.With("service", "quix", "version", "1.0.0")
	reqLogger.Info(ctx, "带公共字段的请求日志", "method", "POST", "path", "/users")

	// 8. SetLevel 级别过滤：设为 Error 后 Info/Warn 被抑制
	log.SetLevel(log.LevelError)
	log.Info(ctx, "这条 info 不会输出")
	log.Warn(ctx, "这条 warn 不会输出")
	log.Error(ctx, "这条 error 正常输出")
}
