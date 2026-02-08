package middleware

import (
	"golang.org/x/time/rate"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouterWithLimiter(r rate.Limit, b int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ResetClients()

	router := gin.New()
	router.Use(RateLimiterWithConfig(r, b))
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	return router
}

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	router := setupRouterWithLimiter(1, 5)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
}

func TestRateLimiter_BlocksWhenLimitExceeded(t *testing.T) {
	router := setupRouterWithLimiter(1, 2)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.6.7.8:9999"

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if i < 2 {
			assert.Equal(t, 200, w.Code)
		} else {
			assert.Equal(t, 429, w.Code)
		}
	}
}

func TestRateLimiter_ResetsAfterTime(t *testing.T) {
	router := setupRouterWithLimiter(1, 1)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:1111"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)
	assert.Equal(t, 429, w2.Code)

	time.Sleep(time.Millisecond * 1100)

	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req)
	assert.Equal(t, 200, w3.Code)
}
