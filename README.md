# Aqylly

Быстрый и production-ready HTTP роутер на чистом Go без сторонних зависимостей (кроме `golang.org/x/net` для HTTP/2).

## Особенности

- ✅ **Минимум зависимостей**: только стандартная библиотека Go + `golang.org/x/net/http2`
- ✅ **Быстрый**: использует Radix Tree для эффективного роутинга
- ✅ **URL параметры**: поддержка динамических параметров `:id` и wildcard `*path`
- ✅ **Middleware**: гибкая система middleware на глобальном и групповом уровне
- ✅ **Группировка маршрутов**: вложенные группы с общими префиксами и middleware
- ✅ **Query параметры**: удобная работа с query параметрами с type-safe API
- ✅ **HTTP/2**: полная поддержка HTTP/2 с Server Push и h2c (cleartext)
- ✅ **Context API**: стандартный `context.Context` для таймаутов, отмены и передачи данных
- ✅ **Graceful Shutdown**: корректное завершение с использованием context
- ✅ **Production Ready**: HPACK compression, multiplexing, flow control

## Установка

```bash
go get github.com/maksat/aqylly
```

## Быстрый старт

```go
package main

import (
    "log"
    "github.com/maksat/aqylly"
)

func main() {
    // Создаем роутер с дефолтными middleware (Logger и Recovery)
    router := aqylly.Default()

    // Простой GET маршрут
    router.GET("/", func(c *aqylly.Context) {
        c.JSON(200, map[string]string{
            "message": "Hello, World!",
        })
    })

    // Запускаем сервер
    log.Fatal(router.Run(":8080"))
}
```

## Примеры использования

### Базовые маршруты

```go
router := aqylly.New()

// HTTP методы
router.GET("/users", getUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
router.PATCH("/users/:id", patchUser)

// Поддержка всех методов
router.Any("/ping", func(c *aqylly.Context) {
    c.String(200, "pong")
})
```

### URL параметры

```go
// Параметр :id
router.GET("/users/:id", func(c *aqylly.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{"user_id": userID})
})

// Несколько параметров
router.GET("/users/:id/posts/:postId", func(c *aqylly.Context) {
    userID := c.Param("id")
    postID := c.Param("postId")
    c.JSON(200, map[string]interface{}{
        "user_id": userID,
        "post_id": postID,
    })
})

// Catch-all параметр
router.GET("/files/*filepath", func(c *aqylly.Context) {
    filepath := c.Param("filepath")
    c.String(200, "Filepath: %s", filepath)
})
```

### Query параметры

```go
router.GET("/search", func(c *aqylly.Context) {
    // Получить query параметр
    query := c.Query("q")

    // С дефолтным значением
    page := c.QueryIntDefault("page", 1)
    limit := c.QueryIntDefault("limit", 10)

    // Массив значений
    tags := c.QueryArray("tags")

    c.JSON(200, map[string]interface{}{
        "query": query,
        "page":  page,
        "limit": limit,
        "tags":  tags,
    })
})
```

### JSON и другие форматы

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

### Группировка маршрутов

```go
router := aqylly.Default()

// Создать группу
api := router.Group("/api")
{
    api.GET("/status", statusHandler)

    // Вложенная группа
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
// Глобальные middleware
router := aqylly.New()
router.Use(aqylly.Logger())
router.Use(aqylly.Recovery())
router.Use(aqylly.RequestID())
router.Use(aqylly.Secure())

// Middleware для группы
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

### Встроенные Middleware

#### Logger
Логирует HTTP запросы:
```go
router.Use(aqylly.Logger())
```

#### Recovery
Перехватывает panic и возвращает 500:
```go
router.Use(aqylly.Recovery())
```

#### CORS
Настраивает CORS headers:
```go
router.Use(aqylly.CORS(
    []string{"https://example.com"},
    []string{"GET", "POST"},
    []string{"Content-Type"},
))
```

#### BasicAuth
Базовая HTTP аутентификация:
```go
router.Use(aqylly.BasicAuth("username", "password"))
```

#### RateLimiter
Ограничение количества запросов:
```go
router.Use(aqylly.RateLimiter(100)) // 100 запросов в секунду
```

#### RequestID
Добавляет уникальный ID к каждому запросу:
```go
router.Use(aqylly.RequestID())
```

#### Secure
Добавляет security headers:
```go
router.Use(aqylly.Secure())
```

#### Timeout
Устанавливает timeout для запросов:
```go
router.Use(aqylly.Timeout(5 * time.Second))
```

### Кастомные Middleware

```go
// Простой middleware
func MyMiddleware() aqylly.HandlerFunc {
    return func(c *aqylly.Context) {
        // До обработки запроса
        log.Println("Before request")

        // Обработать запрос
        c.Next()

        // После обработки запроса
        log.Println("After request")
    }
}

router.Use(MyMiddleware())

// Middleware с abort
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
    // Request информация
    method := c.Method()           // HTTP метод
    path := c.Path()              // Путь запроса
    fullPath := c.FullPath()      // Полный URL
    clientIP := c.ClientIP()      // IP клиента

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

### HTTP/2 поддержка

HTTP/2 включен по умолчанию при использовании TLS. Роутер автоматически настраивает ALPN negotiation.

#### HTTP/2 с TLS

```go
router := aqylly.Default()

// Настройка HTTP/2 параметров
router.HTTP2Config = &aqylly.HTTP2Config{
    MaxConcurrentStreams: 250,
    MaxReadFrameSize:     16384,
    IdleTimeout:          120,
}

// Запуск с TLS и HTTP/2
router.RunTLS(":443", "cert.pem", "key.pem")
```

#### HTTP/2 Server Push

```go
router.GET("/", func(c *aqylly.Context) {
    // Push CSS и JS до отправки HTML
    c.Push("/static/style.css", nil)
    c.Push("/static/app.js", nil)

    c.HTML(200, "<html>...</html>")
})
```

#### HTTP/2 Cleartext (h2c) для микросервисов

Для внутренних микросервисов можно использовать HTTP/2 без TLS:

```go
router := aqylly.Default()

// Запуск HTTP/2 без TLS
router.RunH2C(":8080")
```

Клиент для h2c:

```go
client := &http.Client{
    Transport: aqylly.NewH2CTransport(),
}

resp, _ := client.Get("http://localhost:8080")
fmt.Printf("Protocol: %s\n", resp.Proto) // HTTP/2.0
```

### Context API с таймаутами и отменой

Полная поддержка `context.Context` для таймаутов, отмены и передачи данных:

```go
// Таймаут для запроса
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

// Передача данных через context
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

// Доступ к стандартному context.Context
ctx := c.Context()
deadline, ok := c.Deadline()
```

### Graceful Shutdown

```go
router := aqylly.Default()

// Настройка маршрутов...

// Обработка сигналов для graceful shutdown
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

### Кастомные обработчики ошибок

```go
router := aqylly.New()

// Кастомный 404
router.NotFound = func(c *aqylly.Context) {
    c.JSON(404, map[string]string{
        "error": "Page not found",
        "path":  c.Path(),
    })
}

// Кастомный 405 (Method Not Allowed)
router.MethodNotAllowed = func(c *aqylly.Context) {
    c.JSON(405, map[string]string{
        "error":  "Method not allowed",
        "method": c.Method(),
    })
}
```

## Производительность

Aqylly использует оптимизированное Radix Tree для роутинга, что обеспечивает:
- O(1) для статических маршрутов
- O(log n) для динамических параметров
- Минимальное использование памяти
- Высокую производительность

## Примеры

### Базовый пример

```bash
cd examples
go run main.go
```

Затем откройте http://localhost:8080 в браузере.

### HTTP/2 с TLS и Server Push

```bash
# Сначала сгенерируйте сертификаты
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Запустите сервер
cd examples
go run http2_tls.go
```

Откройте https://localhost:8443

### HTTP/2 Cleartext (h2c) для микросервисов

Сервер:
```bash
cd examples
go run http2_h2c.go
```

Клиент:
```bash
cd examples
go run http2_h2c_client.go
```

Или с curl:
```bash
curl --http2-prior-knowledge http://localhost:8080
```

## Структура проекта

```
aqylly/
├── router.go        # Основной роутер с HTTP/2 поддержкой
├── context.go       # Context API с context.Context
├── tree.go          # Radix tree для URL routing
├── middleware.go    # Встроенные middleware
├── group.go         # Группировка маршрутов
├── http2.go         # HTTP/2 конфигурация и h2c
├── go.mod
├── README.md
└── examples/
    ├── main.go           # Базовый пример
    ├── http2_tls.go      # HTTP/2 с TLS и Server Push
    ├── http2_h2c.go      # HTTP/2 Cleartext сервер
    ├── http2_h2c_client.go # HTTP/2 Cleartext клиент
    └── go.mod
```

## Лицензия

MIT

## Вклад

Contributions are welcome! Feel free to open issues or submit pull requests.

## Авторы

Created with ❤️ by Maksat
