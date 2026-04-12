package main

import (
	"net/http"

	quix "github.com/fztcjjl/quix"
	"github.com/gin-gonic/gin"
)

func main() {
	app := quix.New()

	// Middleware
	app.Use(func(c *gin.Context) {
		c.Next()
	})

	// Routes
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	app.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "hello, " + name})
	})

	// Route group
	api := app.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Start server (Ctrl+C to gracefully shutdown)
	app.Run()
}
