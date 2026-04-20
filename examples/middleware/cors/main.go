package main

import (
	"net/http"
	"time"

	quix "github.com/fztcjjl/quix"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Example 1: Default CORS (built into default middleware chain, allows all origins)
	// curl -H "Origin: http://example.com" http://localhost:8080/default
	app := quix.New()

	app.GET("/default", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "default CORS - allows all origins",
		})
	})

	// Example 2: Disable CORS at App level
	// curl -H "Origin: http://example.com" http://localhost:8080/no-cors
	noCORSApp := quix.New(quix.WithCORS(false))

	noCORSApp.GET("/no-cors", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS is disabled",
		})
	})

	// Example 3: Custom CORS configuration at App level
	// curl -H "Origin: http://example.com" http://localhost:8080/custom
	customApp := quix.New(
		quix.WithCORSConfig(cors.Config{
			AllowOrigins:     []string{"http://example.com", "http://localhost:3000"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)

	customApp.GET("/custom", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "custom CORS - only allows specific origins",
		})
	})

	// Start the default app (others are for demonstration)
	// curl -H "Origin: http://example.com" http://localhost:8080/default
	// curl -X OPTIONS -H "Origin: http://example.com" -H "Access-Control-Request-Method: POST" http://localhost:8080/default
	app.Run()
}
