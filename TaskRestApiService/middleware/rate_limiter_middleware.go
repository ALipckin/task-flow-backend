package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net"
	"strings"
	"sync"
	"time"
)

type client struct {
	limiter  *rate.Limiter // rate limiter for this IP
	lastSeen time.Time     // last seen time of the client
}

var (
	clients      = make(map[string]*client) // map of IP to client data
	clientsMutex sync.Mutex                 // mutex to protect clients map

	globalRateLimit  rate.Limit = 5  // default requests per second
	globalBurstLimit int        = 10 // default burst capacity
)

// getLimiter returns the rate limiter for a given IP,
// creating a new one if it doesn't exist
func getLimiter(ip string) *rate.Limiter {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	c, exists := clients[ip]
	if !exists {
		limiter := rate.NewLimiter(globalRateLimit, globalBurstLimit)
		clients[ip] = &client{limiter, time.Now()}
		return limiter
	}
	c.lastSeen = time.Now()
	return c.limiter
}

// cleanupClients removes clients which have not been seen for more than 3 minutes
func cleanupClients() {
	for {
		time.Sleep(time.Minute)
		clientsMutex.Lock()
		for ip, c := range clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		clientsMutex.Unlock()
	}
}

// ResetClients clears the clients map (useful for tests)
func ResetClients() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients = make(map[string]*client)
}

// RateLimiterWithConfig returns a Gin middleware with configurable rate and burst limits
func RateLimiterWithConfig(rateLimit rate.Limit, burstLimit int) gin.HandlerFunc {
	globalRateLimit = rateLimit
	globalBurstLimit = burstLimit

	go cleanupClients()

	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too Many Requests",
			})
			return
		}

		c.Next()
	}
}

// getClientIP extracts the client IP address from X-Forwarded-For header or RemoteAddr
func getClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}
