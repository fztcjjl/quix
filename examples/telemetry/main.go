// Example: telemetry demonstrates OpenTelemetry integration in quix.
//
// Run:
//
//	go run ./examples/telemetry
//
// The stdout exporter will output OTel traces and metrics to stderr.
package main

import (
	"github.com/fztcjjl/quix"
	"github.com/fztcjjl/quix/core/telemetry"
	"github.com/gin-gonic/gin"
)

func main() {
	app := quix.New(
		quix.WithTelemetry(
			telemetry.WithServiceName("telemetry-demo"),
			telemetry.WithStdoutExporter(true),
		),
	)

	app.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	app.GET("/hello/:name", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello " + c.Param("name")})
	})

	app.Run()
}
