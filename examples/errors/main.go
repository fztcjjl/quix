package main

import (
	"fmt"
	"net/http"

	"github.com/fztcjjl/quix"
	"github.com/fztcjjl/quix/core/errors"
	qhttp "github.com/fztcjjl/quix/core/transport/http/server"
	"github.com/gin-gonic/gin"
)

func main() {
	app := quix.New()

	app.GET("/success", qhttp.Handler(func(c *gin.Context) error {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
		return nil
	}))

	app.GET("/notfound", qhttp.Handler(func(c *gin.Context) error {
		return errors.NotFound("user_not_found", "用户不存在")
	}))

	app.GET("/badrequest", qhttp.Handler(func(c *gin.Context) error {
		return errors.BadRequest("param_invalid", "参数验证失败")
	}))

	app.GET("/internal", qhttp.Handler(func(c *gin.Context) error {
		return fmt.Errorf("database connection failed")
	}))

	app.GET("/custom", qhttp.Handler(func(c *gin.Context) error {
		return &errors.Error{
			Code:       "rate_limited",
			Message:    "请求过于频繁，请稍后重试",
			StatusCode: http.StatusTooManyRequests,
			Details: map[string]any{
				"retry_after": 60,
				"limit":       100,
			},
		}
	}))

	fmt.Println("Server listening on :8080")
	app.Run()
}
