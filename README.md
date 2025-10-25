# Aqylly

Fast and production-ready HTTP router in pure Go with minimal dependencies (only `golang.org/x/net` for HTTP/2).

## Features

- ✅ **Minimal Dependencies**: Only Go standard library + `golang.org/x/net/http2`
- ✅ **Fast**: Uses Radix Tree for efficient routing
- ✅ **URL Parameters**: Support for dynamic parameters `:id` and wildcard `*path`
- ✅ **Middleware**: Flexible middleware system at global and group levels
- ✅ **Route Grouping**: Nested groups with shared prefixes and middleware
- ✅ **Query Parameters**: Convenient query parameter handling with type-safe API
- ✅ **HTTP/2**: Full HTTP/2 support with Server Push and h2c (cleartext)
- ✅ **Context API**: Standard `context.Context` for timeouts, cancellation, and value passing
- ✅ **Graceful Shutdown**: Proper shutdown using context
- ✅ **Production Ready**: HPACK compression, multiplexing, flow control

## Installation

```bash
go get github.com/maksat/aqylly
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/maksat/aqylly"
)

func main() {
    // Create router with default middleware (Logger and Recovery)
    router := aqylly.Default()

    // Simple GET route
    router.GET("/", func(c *aqylly.Context) {
        c.JSON(200, map[string]string{
            "message": "Hello, World!",
        })
    })

    // Start server
    log.Fatal(router.Run(":8080"))
}
```

## Usage Examples

### Basic Routes

```go
router := aqylly.New()

// HTTP methods
router.GET("/users", getUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
router.PATCH("/users/:id", patchUser)

// Support for all methods
router.Any("/ping", func(c *aqylly.Context) {
    c.String(200, "pong")
})
```

### URL Parameters

```go
// Parameter :id
router.GET("/users/:id", func(c *aqylly.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{"user_id": userID})
})

// Multiple parameters
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
    c.String(200, "Filepath: %s", filepath)
})
```

### Query Parameters

```go
router.GET("/search", func(c *aqylly.Context) {
    // Get query parameter
    query := c.Query("q")

    // With default value
    page := c.QueryIntDefault("page", 1)
    limit := c.QueryIntDefault("limit", 10)

    // Array values
    tags := c.QueryArray("tags")

    c.JSON(200, map[string]interface{}{
        "query": query,
        "page":  page,
        "limit": limit,
        "tags":  tags,
    })
})
```

### JSON and Other Formats

```go
// JSON response
router.GET("/json", func(c *aqylly.Context) {
    c.JSON(200, map[string]string{"message": "hello"})
})

// Bind JSON request
router.POST("/users", func(c *aqylly.Context) {
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    if err := c.BindJSON(&user); err != nil {
        c.JSON(400, map[string]string{"error": "Invalid JSON"})
        return
    }

    c.JSON(201, user)
})

// HTML response
router.GET("/html", func(c *aqylly.Context) {
    html := "<h1>Hello, World!</h1>"
    c.HTML(200, html)
})

// Plain text
router.GET("/text", func(c *aqylly.Context) {
    c.String(200, "Hello, %s!", "World")
})

// Binary data
router.GET("/data", func(c *aqylly.Context) {
    data := []byte("binary data")
    c.Data(200, "application/octet-stream", data)
})
```

### Route Grouping

```go
router := aqylly.Default()

// Create a group
api := router.Group("/api")
{
    api.GET("/status", statusHandler)

    // Nested group
    v1 := api.Group("/v1")
    {
        v1.GET("/users", getUsersV1)
        v1.POST("/users", createUserV1)
    }

    v2 := api.Group("/v2")
    {
        v2.GET("/users", getUsersV2)
        v2.POST("/users", createUserV2)
    }
}
```

### Middleware

```go
// Global middleware
router := aqylly.New()
router.Use(aqylly.Logger())
router.Use(aqylly.Recovery())
router.Use(aqylly.RequestID())
router.Use(aqylly.Secure())

// Group-level middleware
admin := router.Group("/admin", aqylly.BasicAuth("admin", "secret"))
{
    admin.GET("/dashboard", dashboardHandler)
}

// CORS middleware
corsGroup := router.Group("/api",
    aqylly.CORS(
        []string{"*"},
        []string{"GET", "POST", "PUT", "DELETE"},
        []string{"Content-Type", "Authorization"},
    ),
)

// Rate limiting
limited := router.Group("/api", aqylly.RateLimiter(100)) // 100 req/sec
```

### Built-in Middleware

#### Logger
Logs HTTP requests:
```go
router.Use(aqylly.Logger())
```

#### Recovery
Catches panics and returns 500:
```go
router.Use(aqylly.Recovery())
```

#### CORS
Configures CORS headers:
```go
router.Use(aqylly.CORS(
    []string{"https://example.com"},
    []string{"GET", "POST"},
    []string{"Content-Type"},
))
```

#### BasicAuth
Basic HTTP authentication:
```go
router.Use(aqylly.BasicAuth("username", "password"))
```

#### RateLimiter
Request rate limiting:
```go
router.Use(aqylly.RateLimiter(100)) // 100 requests per second
```

#### RequestID
Adds unique ID to each request:
```go
router.Use(aqylly.RequestID())
```

#### Secure
Adds security headers:
```go
router.Use(aqylly.Secure())
```

#### Timeout
Sets timeout for requests:
```go
router.Use(aqylly.Timeout(5 * time.Second))
```

### Custom Middleware

```go
// Simple middleware
func MyMiddleware() aqylly.HandlerFunc {
    return func(c *aqylly.Context) {
        // Before request processing
        log.Println("Before request")

        // Process request
        c.Next()

        // After request processing
        log.Println("After request")
    }
}

router.Use(MyMiddleware())

// Middleware with abort
func AuthMiddleware() aqylly.HandlerFunc {
    return func(c *aqylly.Context) {
        token := c.Header("Authorization")

        if token == "" {
            c.AbortWithJSON(401, map[string]string{
                "error": "Unauthorized",
            })
            return
        }

        c.Next()
    }
}
```

### Context API

```go
router.GET("/demo", func(c *aqylly.Context) {
    // Request information
    method := c.Method()           // HTTP method
    path := c.Path()              // Request path
    fullPath := c.FullPath()      // Full URL
    clientIP := c.ClientIP()      // Client IP

    // Headers
    contentType := c.ContentType()
    userAgent := c.Header("User-Agent")
    c.SetHeader("X-Custom", "value")

    // Cookies
    sessionID, _ := c.Cookie("session_id")
    c.SetCookie("new_cookie", "value", 3600, "/", "", false, true)

    // Form data
    name := c.FormValue("name")
    email := c.PostForm("email")

    // Type checking
    isJSON := c.IsJSON()
    isXML := c.IsXML()
    isForm := c.IsForm()

    // Response helpers
    c.Status(200)
    c.JSON(200, data)
    c.String(200, "text")
    c.HTML(200, "<html>")
    c.Redirect(302, "/other")

    // Error handling
    c.Error(500, errors.New("something went wrong"))
    c.AbortWithStatus(404)
    c.AbortWithJSON(400, map[string]string{"error": "bad request"})
})
```

### HTTP/2 Support

HTTP/2 is enabled by default when using TLS. The router automatically configures ALPN negotiation.

#### HTTP/2 with TLS

```go
router := aqylly.Default()

// Configure HTTP/2 parameters
router.HTTP2Config = &aqylly.HTTP2Config{
    MaxConcurrentStreams: 250,
    MaxReadFrameSize:     16384,
    IdleTimeout:          120,
}

// Run with TLS and HTTP/2
router.RunTLS(":443", "cert.pem", "key.pem")
```

#### HTTP/2 Server Push

```go
router.GET("/", func(c *aqylly.Context) {
    // Push CSS and JS before sending HTML
    c.Push("/static/style.css", nil)
    c.Push("/static/app.js", nil)

    c.HTML(200, "<html>...</html>")
})
```

#### HTTP/2 Cleartext (h2c) for Microservices

For internal microservices, you can use HTTP/2 without TLS:

```go
router := aqylly.Default()

// Run HTTP/2 without TLS
router.RunH2C(":8080")
```

Client for h2c:

```go
client := &http.Client{
    Transport: aqylly.NewH2CTransport(),
}

resp, _ := client.Get("http://localhost:8080")
fmt.Printf("Protocol: %s\n", resp.Proto) // HTTP/2.0
```

### Context API with Timeouts and Cancellation

Full support for `context.Context` for timeouts, cancellation, and value passing:

```go
// Request timeout
router.GET("/slow", func(c *aqylly.Context) {
    cancel, _ := c.WithTimeout(2 * time.Second)
    defer cancel()

    select {
    case <-time.After(3 * time.Second):
        c.JSON(200, map[string]string{"status": "done"})
    case <-c.Done():
        c.JSON(408, map[string]string{"error": "timeout"})
    }
})

// Pass data through context
router.Use(func(c *aqylly.Context) {
    c.Set("request_id", generateID())
    c.Next()
})

router.GET("/api/data", func(c *aqylly.Context) {
    requestID, _ := c.Get("request_id")
    c.JSON(200, map[string]interface{}{
        "request_id": requestID,
        "data": "...",
    })
})

// Access to standard context.Context
ctx := c.Context()
deadline, ok := c.Deadline()
```

### Graceful Shutdown

```go
router := aqylly.Default()

// Configure routes...

// Handle signals for graceful shutdown
go func() {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := router.Shutdown(ctx); err != nil {
        log.Fatal("Shutdown error:", err)
    }
}()

router.Run(":8080")
```

### Custom Error Handlers

```go
router := aqylly.New()

// Custom 404
router.NotFound = func(c *aqylly.Context) {
    c.JSON(404, map[string]string{
        "error": "Page not found",
        "path":  c.Path(),
    })
}

// Custom 405 (Method Not Allowed)
router.MethodNotAllowed = func(c *aqylly.Context) {
    c.JSON(405, map[string]string{
        "error":  "Method not allowed",
        "method": c.Method(),
    })
}
```

## Performance

Aqylly uses an optimized Radix Tree for routing, which provides:
- O(1) for static routes
- O(log n) for dynamic parameters
- Minimal memory usage
- High performance

## Examples

### Basic Example

```bash
cd examples
go run main.go
```

Then open http://localhost:8080 in your browser.

### HTTP/2 with TLS and Server Push

```bash
# First, generate certificates
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Run the server
cd examples
go run http2_tls.go
```

Open https://localhost:8443

### HTTP/2 Cleartext (h2c) for Microservices

Server:
```bash
cd examples
go run http2_h2c.go
```

Client:
```bash
cd examples
go run http2_h2c_client.go
```

Or with curl:
```bash
curl --http2-prior-knowledge http://localhost:8080
```

## Project Structure

```
aqylly/
├── router.go        # Main router with HTTP/2 support
├── context.go       # Context API with context.Context
├── tree.go          # Radix tree for URL routing
├── middleware.go    # Built-in middleware
├── group.go         # Route grouping
├── http2.go         # HTTP/2 configuration and h2c
├── go.mod
├── README.md
└── examples/
    ├── main.go           # Basic example
    ├── http2_tls.go      # HTTP/2 with TLS and Server Push
    ├── http2_h2c.go      # HTTP/2 Cleartext server
    ├── http2_h2c_client.go # HTTP/2 Cleartext client
    └── go.mod
```

## License

MIT

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## Author

Created with ❤️ by Maksat
