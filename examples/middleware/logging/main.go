package main

import (
	"net/http"

	quix "github.com/fztcjjl/quix"
	"github.com/gin-gonic/gin"
)

func main() {
	// Use WithLoggingSkipPaths to skip logging for specific paths
	app := quix.New(quix.WithLoggingSkipPaths("/healthz"))

	app.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	app.GET("/notfound", func(c *gin.Context) {
		c.String(http.StatusNotFound, "not found")
	})
	app.GET("/server-error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "oops")
	})
	// /healthz is skipped by Logging middleware (no log output)
	app.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	// Start server
	// curl http://localhost:8080/ok            → 200 (Info level log)
	// curl http://localhost:8080/notfound      → 404 (Warn level log)
	// curl http://localhost:8080/server-error  → 500 (Error level log)
	// curl http://localhost:8080/healthz       → 200 (no log, skipped)
	app.Run()
}
