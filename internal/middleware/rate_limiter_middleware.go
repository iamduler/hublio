package middleware

import (
	"net/http"
	"shopping-cart/internal/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter     *rate.Limiter
	lastRequest time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client)
)

func getClientIP(c *gin.Context) string {
	ip := c.ClientIP()

	if ip == "" {
		ip = c.Request.RemoteAddr // Get the client's IP address from the request when the client is using a proxy
	}

	return ip
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock() // Lock the mutex to prevent race conditions
	defer mu.Unlock()

	client, exists := clients[ip]

	if !exists {
		requestsPerSecond := utils.GetIntEnv("RATE_LIMIT_REQUESTS_SEC", 5)
		burst := utils.GetIntEnv("RATE_LIMIT_BURST", 10)

		limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst) // 5 requests per second, burst of 10 requests
		newClient := &Client{limiter, time.Now()}                        // Create a new client with the limiter and the last request time
		clients[ip] = newClient                                          // Store the new client in the map
		return newClient.limiter
	}

	client.lastRequest = time.Now()
	return client.limiter
}

func CleanUpOldClients() {
	for {
		time.Sleep(1 * time.Minute) // Clean up old clients every minute
		mu.Lock()

		for ip, client := range clients {
			if time.Since(client.lastRequest) > 3*time.Minute {
				delete(clients, ip)
			}
		}

		mu.Unlock()
	}
}

/*
Test: ab -n 100 -c 10 -H "X-API-KEY: duler-api-key" http://localhost:8080/api/v1/users
-n 100: 100 requests
-c 10: 10 concurrent requests
-H "X-API-KEY: duler-api-key": API key
*/
func RateLimitMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := getRateLimiter(ip)

		if !limiter.Allow() {
			if shouldLogRateLimit(ip) {
				logger.Warn().
					Str("ip", ip).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("query", c.Request.URL.RawQuery).
					Str("user-agent", c.Request.UserAgent()).
					Str("referer", c.Request.Referer()).
					Str("protocol", c.Request.Proto).
					Str("host", c.Request.Host).
					Interface("headers", c.Request.Header).
					Msg("Too many requests")
			}

			// Return a 429 status code and a JSON response with the error message
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}

/*
Cache the rate limit logs to prevent duplicate logs
sync.Map is a thread-safe map, allow multiple goroutines to access the map concurrently
*/
var rateLimitLogCache = sync.Map{}

func shouldLogRateLimit(ip string) bool {
	now := time.Now()

	if val, ok := rateLimitLogCache.Load(ip); ok {
		lastLogTime, ok := val.(time.Time)

		// If the last log time is within 1 minute, return false
		if ok && now.Sub(lastLogTime) < 1*time.Minute {
			return false
		}
	}

	// If the last log time is not within 1 minute, return true
	rateLimitLogCache.Store(ip, now)
	return true
}
