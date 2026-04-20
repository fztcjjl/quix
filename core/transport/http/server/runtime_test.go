package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	qerrors "github.com/fztcjjl/quix/core/errors"
	"github.com/fztcjjl/quix/core/transport/http/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetError_WithAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		err := &qerrors.Error{Code: "not_found", Message: "not found", StatusCode: 404}
		ctx.SetError(err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetError_GetError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var retrieved error
	r.GET("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		err := &qerrors.Error{Code: "bad_request", Message: "invalid param", StatusCode: 400}
		ctx.SetError(err)
		retrieved = ctx.GetError()
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.NotNil(t, retrieved)
	var target *qerrors.Error
	assert.True(t, errors.As(retrieved, &target))
	assert.Equal(t, "bad_request", target.Code)
	assert.Equal(t, "invalid param", target.Message)
}

func TestSetError_WithStandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ResponseMiddleware())
	r.GET("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		ctx.SetError(assert.AnError)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetError_NoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		result := ctx.GetError()
		assert.Nil(t, result)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

func TestShouldBindQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var req testRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?name=alice&email=alice@example.com", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result testRequest
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "alice", result.Name)
	assert.Equal(t, "alice@example.com", result.Email)
}

func TestShouldBindUri(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	r := gin.New()
	r.GET("/test/:id/:name", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var req testRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/123/alice", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result testRequest
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "alice", result.Name)
}

func TestShouldBindUriConflictCheck_NoConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	r := gin.New()
	r.POST("/test/:id", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var req testRequest
		if err := ctx.ShouldBindUriConflictCheck(&req, []string{"id"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test/123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result testRequest
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "123", result.ID)
}

func TestShouldBindUriConflictCheck_ValueMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	r := gin.New()
	r.POST("/test/:id", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var req testRequest
		req.ID = "123"
		if err := ctx.ShouldBindUriConflictCheck(&req, []string{"id"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test/123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result testRequest
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "123", result.ID)
}

func TestShouldBindUriConflictCheck_ValueMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	r := gin.New()
	r.POST("/test/:id", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var req testRequest
		req.ID = "999"
		if err := ctx.ShouldBindUriConflictCheck(&req, []string{"id"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test/123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "conflicts with body field")
}

func TestShouldBindUriConflictCheck_NonStructPointer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/test/:id", func(c *gin.Context) {
		ctx := &Context{Context: c}
		var s string
		err := ctx.ShouldBindUriConflictCheck(&s, []string{"id"})
		assert.NoError(t, err)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test/123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
