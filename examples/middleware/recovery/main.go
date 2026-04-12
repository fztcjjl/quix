package main

import (
	"net/http"

	quix "github.com/fztcjjl/quix"
	"github.com/gin-gonic/gin"
)

func main() {
	app := quix.New()

	app.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	app.GET("/panic", func(c *gin.Context) {
		panic("something went wrong!")
	})

	// Start server
	// curl http://localhost:8080/ok        → 200 (normal)
	// curl http://localhost:8080/panic     → 500 (recovered, not crashed)
	app.Run()
}
