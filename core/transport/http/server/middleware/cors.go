package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a CORS middleware with default configuration.
// Default allows all origins, common methods (GET, POST, PUT, PATCH, DELETE, OPTIONS)
// and common headers.
func CORS() gin.HandlerFunc {
	return cors.Default()
}

// WithCORSConfig returns a CORS middleware with the given configuration.
func WithCORSConfig(cfg cors.Config) gin.HandlerFunc {
	return cors.New(cfg)
}
