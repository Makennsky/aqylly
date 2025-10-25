package main

import (
	"log"
	"time"

	"github.com/Makennsky/aqylly"
)

func main() {
	// Create router with default middleware (Logger and Recovery)
	router := aqylly.Default()

	// Enable HTTP/2 (enabled by default)
	router.EnableHTTP2 = true

	// Optional: Enable HTTP/3
	// router.EnableHTTP3 = true

	// Custom middleware examples
	router.Use(aqylly.RequestID())
	router.Use(aqylly.Secure())

	// Basic routes
	router.GET("/", func(c *aqylly.Context) {
		c.JSON(200, map[string]interface{}{
			"message": "Welcome to Aqylly Router!",
			"version": "1.0.0",
		})
	})

	router.GET("/ping", func(c *aqylly.Context) {
		c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	// URL parameters example
	router.GET("/users/:id", func(c *aqylly.Context) {
		userID := c.Param("id")
		c.JSON(200, map[string]string{
			"user_id": userID,
			"message": "User details",
		})
	})

	router.GET("/users/:id/posts/:postId", func(c *aqylly.Context) {
		userID := c.Param("id")
		postID := c.Param("postId")
		c.JSON(200, map[string]interface{}{
			"user_id": userID,
			"post_id": postID,
		})
	})

	// Catch-all parameter
	router.GET("/files/*filepath", func(c *aqylly.Context) {
		filepath := c.Param("filepath")
		c.JSON(200, map[string]string{
			"filepath": filepath,
		})
	})

	// Query parameters example
	router.GET("/search", func(c *aqylly.Context) {
		query := c.Query("q")
		page := c.QueryIntDefault("page", 1)
		limit := c.QueryIntDefault("limit", 10)

		c.JSON(200, map[string]interface{}{
			"query": query,
			"page":  page,
			"limit": limit,
		})
	})

	// POST example with JSON binding
	router.POST("/users", func(c *aqylly.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Age   int    `json:"age"`
		}

		if err := c.BindJSON(&user); err != nil {
			c.JSON(400, map[string]string{
				"error": "Invalid JSON",
			})
			return
		}

		c.JSON(201, map[string]interface{}{
			"message": "User created",
			"user":    user,
		})
	})

	// Route groups example
	api := router.Group("/api")
	{
		api.GET("/status", func(c *aqylly.Context) {
			c.JSON(200, map[string]string{
				"status": "ok",
			})
		})

		// Nested groups
		v1 := api.Group("/v1")
		{
			v1.GET("/products", func(c *aqylly.Context) {
				c.JSON(200, map[string]interface{}{
					"products": []string{"Product 1", "Product 2"},
				})
			})

			v1.GET("/products/:id", func(c *aqylly.Context) {
				productID := c.Param("id")
				c.JSON(200, map[string]string{
					"product_id": productID,
					"name":       "Sample Product",
				})
			})
		}

		v2 := api.Group("/v2")
		{
			v2.GET("/products", func(c *aqylly.Context) {
				c.JSON(200, map[string]interface{}{
					"products": []string{"Product A", "Product B", "Product C"},
					"version":  "2.0",
				})
			})
		}
	}

	// Protected routes with authentication
	admin := router.Group("/admin", aqylly.BasicAuth("admin", "secret"))
	{
		admin.GET("/dashboard", func(c *aqylly.Context) {
			c.JSON(200, map[string]string{
				"message": "Admin Dashboard",
			})
		})

		admin.GET("/users", func(c *aqylly.Context) {
			c.JSON(200, map[string]interface{}{
				"users": []string{"User 1", "User 2", "User 3"},
			})
		})
	}

	// CORS example
	corsGroup := router.Group("/cors",
		aqylly.CORS(
			[]string{"*"}, // Allow all origins
			[]string{"GET", "POST", "PUT", "DELETE"},
			[]string{"Content-Type", "Authorization"},
		),
	)
	{
		corsGroup.GET("/data", func(c *aqylly.Context) {
			c.JSON(200, map[string]string{
				"message": "CORS enabled endpoint",
			})
		})
	}

	// Rate limiting example
	limited := router.Group("/limited", aqylly.RateLimiter(10)) // 10 requests per second
	{
		limited.GET("/resource", func(c *aqylly.Context) {
			c.JSON(200, map[string]string{
				"message": "Rate limited resource",
			})
		})
	}

	// Different content types
	router.GET("/html", func(c *aqylly.Context) {
		html := `
		<!DOCTYPE html>
		<html>
		<head><title>Aqylly</title></head>
		<body>
			<h1>Welcome to Aqylly Router!</h1>
			<p>This is an HTML response example.</p>
		</body>
		</html>
		`
		c.HTML(200, html)
	})

	router.GET("/text", func(c *aqylly.Context) {
		c.String(200, "Plain text response: %s", "Hello, World!")
	})

	router.GET("/data", func(c *aqylly.Context) {
		data := []byte("Binary data response")
		c.Data(200, "application/octet-stream", data)
	})

	// Redirect example
	router.GET("/redirect", func(c *aqylly.Context) {
		c.Redirect(302, "/")
	})

	// Cookie example
	router.GET("/cookie/set", func(c *aqylly.Context) {
		c.SetCookie("session_id", "abc123", 3600, "/", "", false, true)
		c.JSON(200, map[string]string{
			"message": "Cookie set",
		})
	})

	router.GET("/cookie/get", func(c *aqylly.Context) {
		cookie, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(404, map[string]string{
				"error": "Cookie not found",
			})
			return
		}

		c.JSON(200, map[string]string{
			"session_id": cookie,
		})
	})

	// Custom 404 handler
	router.NotFound = func(c *aqylly.Context) {
		c.JSON(404, map[string]interface{}{
			"error":   "Not Found",
			"path":    c.Path(),
			"message": "The requested resource was not found",
		})
	}

	// Custom 405 handler
	router.MethodNotAllowed = func(c *aqylly.Context) {
		c.JSON(405, map[string]interface{}{
			"error":   "Method Not Allowed",
			"method":  c.Method(),
			"path":    c.Path(),
			"message": "The requested method is not allowed for this resource",
		})
	}

	// Timeout example
	router.GET("/slow", aqylly.Timeout(2*time.Second), func(c *aqylly.Context) {
		// Simulate slow operation
		time.Sleep(3 * time.Second)
		c.JSON(200, map[string]string{
			"message": "This will timeout",
		})
	})

	// Panic recovery example
	router.GET("/panic", func(c *aqylly.Context) {
		panic("Something went wrong!")
	})

	// All HTTP methods example
	router.Any("/methods", func(c *aqylly.Context) {
		c.JSON(200, map[string]string{
			"method":  c.Method(),
			"message": "This endpoint supports all HTTP methods",
		})
	})

	// Start server
	log.Println("Starting server on :8080")
	log.Println("Visit http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
