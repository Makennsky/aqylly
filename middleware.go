package aqylly

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

// Logger returns a middleware that logs HTTP requests
func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		statusCode := c.statusCode
		clientIP := c.ClientIP()

		log.Printf("[%s] %s %s %d %v from %s",
			method,
			path,
			getStatusColor(statusCode),
			statusCode,
			duration,
			clientIP,
		)
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return 500 Internal Server Error
				c.AbortWithJSON(500, map[string]interface{}{
					"error":   "Internal Server Error",
					"message": fmt.Sprintf("%v", err),
				})
			}
		}()

		c.Next()
	}
}

// CORS returns a middleware that handles CORS
func CORS(allowOrigins, allowMethods, allowHeaders []string) HandlerFunc {
	return func(c *Context) {
		origin := c.Header("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowOrigin := range allowOrigins {
			if allowOrigin == "*" || allowOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
				c.SetHeader("Access-Control-Allow-Origin", "*")
			} else {
				c.SetHeader("Access-Control-Allow-Origin", origin)
			}

			c.SetHeader("Access-Control-Allow-Methods", joinSlice(allowMethods))
			c.SetHeader("Access-Control-Allow-Headers", joinSlice(allowHeaders))
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight OPTIONS request
		if c.Method() == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// BasicAuth returns a basic authentication middleware
func BasicAuth(username, password string) HandlerFunc {
	return func(c *Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != username || pass != password {
			c.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithJSON(401, map[string]string{
				"error": "Unauthorized",
			})
			return
		}
		c.Next()
	}
}

// RateLimiter returns a simple rate limiting middleware
// Note: This is a basic in-memory implementation
func RateLimiter(requestsPerSecond int) HandlerFunc {
	type client struct {
		lastRequest time.Time
		count       int
	}

	clients := make(map[string]*client)

	return func(c *Context) {
		ip := c.ClientIP()
		now := time.Now()

		if cl, exists := clients[ip]; exists {
			if now.Sub(cl.lastRequest) < time.Second {
				cl.count++
				if cl.count > requestsPerSecond {
					c.AbortWithJSON(429, map[string]string{
						"error": "Too Many Requests",
					})
					return
				}
			} else {
				cl.count = 1
				cl.lastRequest = now
			}
		} else {
			clients[ip] = &client{
				lastRequest: now,
				count:       1,
			}
		}

		c.Next()
	}
}

// RequestID returns a middleware that adds a unique request ID
func RequestID() HandlerFunc {
	return func(c *Context) {
		requestID := c.Header("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.SetHeader("X-Request-ID", requestID)
		c.Next()
	}
}

// Timeout returns a middleware that sets a timeout for requests
func Timeout(duration time.Duration) HandlerFunc {
	return func(c *Context) {
		// Create a channel to signal completion
		done := make(chan struct{})

		// Run handler in goroutine
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for either completion or timeout
		select {
		case <-done:
			return
		case <-time.After(duration):
			c.AbortWithJSON(408, map[string]string{
				"error": "Request Timeout",
			})
		}
	}
}

// Secure returns a middleware that adds security headers
func Secure() HandlerFunc {
	return func(c *Context) {
		c.SetHeader("X-Content-Type-Options", "nosniff")
		c.SetHeader("X-Frame-Options", "DENY")
		c.SetHeader("X-XSS-Protection", "1; mode=block")
		c.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// Compress returns a middleware that compresses responses (placeholder)
// Note: Full implementation would require gzip compression
func Compress() HandlerFunc {
	return func(c *Context) {
		// Placeholder for compression logic
		// In a real implementation, you would check Accept-Encoding header
		// and wrap the ResponseWriter with a compression writer
		c.Next()
	}
}

// Helper functions

func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "✓"
	case code >= 300 && code < 400:
		return "→"
	case code >= 400 && code < 500:
		return "⚠"
	default:
		return "✗"
	}
}

func joinSlice(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
