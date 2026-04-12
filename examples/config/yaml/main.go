package main

import (
	"fmt"
	"os"

	quix "github.com/fztcjjl/quix"
	"github.com/fztcjjl/quix/core/config"
)

func main() {
	cfg, err := config.NewKoanf(config.WithFile("config.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	app := quix.New(quix.WithConfig(cfg))

	// 读取各种类型的值
	fmt.Println("=== YAML 配置加载示例 ===")
	fmt.Printf("app.name:  %s\n", app.Config().String("app.name"))
	fmt.Printf("app.debug: %v\n", app.Config().Bool("app.debug"))
	fmt.Printf("server.host: %s\n", app.Config().String("server.host"))
	fmt.Printf("server.port: %d\n", app.Config().Int("server.port"))

	// 嵌套键名访问
	fmt.Printf("database.host: %s\n", app.Config().String("database.host"))
	fmt.Printf("database.port: %d\n", app.Config().Int("database.port"))

	// Get 返回原始值
	fmt.Printf("server (raw): %v\n", app.Config().Get("server"))

	// Bind 到结构体
	type ServerConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}
	var sc ServerConfig
	if err := app.Config().Bind("server", &sc); err != nil {
		fmt.Fprintf(os.Stderr, "bind failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("server struct: %+v\n", sc)

	// 不存在的 key 返回零值
	fmt.Printf("nonexistent: %q\n", app.Config().String("nonexistent"))
}
