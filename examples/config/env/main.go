package main

import (
	"fmt"
	"os"

	quix "github.com/fztcjjl/quix"
	"github.com/fztcjjl/quix/core/config"
)

func main() {
	// 模拟设置环境变量（生产环境中由系统设置）
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("APP_NAME", "production-app")
	os.Setenv("APP_DEBUG", "false")

	// 不加载文件，纯环境变量
	cfg, err := config.NewKoanf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	app := quix.New(quix.WithConfig(cfg))

	fmt.Println("=== 环境变量配置示例 ===")
	fmt.Printf("server.port: %d\n", app.Config().Int("server.port"))
	fmt.Printf("app.name:    %s\n", app.Config().String("app.name"))
	fmt.Printf("app.debug:   %v\n", app.Config().Bool("app.debug"))

	// 也可以文件 + 环境变量组合使用（环境变量优先）
	// cfg, err := config.NewKoanf(config.WithFile("config.yaml"))
}
