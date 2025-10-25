package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Makennsky/aqylly"
)

func main() {
	router := aqylly.Default()

	// Configure HTTP/2 settings
	router.HTTP2Config = &aqylly.HTTP2Config{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     16384,
		IdleTimeout:          120,
	}

	// Example with Server Push
	router.GET("/", func(c *aqylly.Context) {
		// Push CSS and JS before sending HTML
		if err := c.Push("/static/style.css", nil); err != nil {
			log.Printf("Failed to push /static/style.css: %v", err)
		}

		if err := c.Push("/static/app.js", nil); err != nil {
			log.Printf("Failed to push /static/app.js: %v", err)
		}

		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>HTTP/2 with Server Push</title>
			<link rel="stylesheet" href="/static/style.css">
		</head>
		<body>
			<h1>HTTP/2 Server Push Demo</h1>
			<p>The CSS and JS files were pushed by the server before you requested them!</p>
			<script src="/static/app.js"></script>
		</body>
		</html>
		`
		c.HTML(200, html)
	})

	router.GET("/static/style.css", func(c *aqylly.Context) {
		css := `
		body { font-family: Arial, sans-serif; margin: 40px; }
		h1 { color: #333; }
		`
		c.Data(200, "text/css", []byte(css))
	})

	router.GET("/static/app.js", func(c *aqylly.Context) {
		js := `console.log('JavaScript loaded via HTTP/2 Server Push!');`
		c.Data(200, "application/javascript", []byte(js))
	})

	// Example with context timeout
	router.GET("/slow", func(c *aqylly.Context) {
		// Set timeout for this request
		cancel, _ := c.WithTimeout(2 * time.Second)
		defer cancel()

		// Simulate slow operation
		select {
		case <-time.After(3 * time.Second):
			c.JSON(200, map[string]string{"message": "completed"})
		case <-c.Done():
			c.JSON(408, map[string]string{"error": "Request timeout"})
		}
	})

	// Example with context values
	router.Use(func(c *aqylly.Context) {
		// Store request ID in context
		requestID := time.Now().UnixNano()
		c.Set("request_id", requestID)
		c.Next()
	})

	router.GET("/user/:id", func(c *aqylly.Context) {
		// Get value from context
		requestID, _ := c.Get("request_id")

		userID := c.Param("id")
		c.JSON(200, map[string]interface{}{
			"user_id":    userID,
			"request_id": requestID,
			"protocol":   c.Request.Proto,
		})
	})

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := router.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}

		log.Println("Server exited")
		os.Exit(0)
	}()

	// Generate self-signed certificates for testing:
	// openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
	log.Println("Starting HTTPS server with HTTP/2 on :8443")
	log.Println("Visit https://localhost:8443")
	log.Println("Note: You'll need to generate cert.pem and key.pem")

	if err := router.RunTLS(":8443", "cert.pem", "key.pem"); err != nil {
		log.Fatal(err)
	}
}
