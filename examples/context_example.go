package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Makennsky/aqylly"
)

func main() {
	router := aqylly.Default()

	// Пример 1: Внешний контекст приходит из http.Request
	router.GET("/external", func(c *aqylly.Context) {
		// Получаем контекст из запроса (он может быть установлен снаружи)
		ctx := c.Context()

		// Проверяем, есть ли значения в контексте
		if traceID := ctx.Value("trace-id"); traceID != nil {
			fmt.Printf("External trace ID: %v\n", traceID)
		}

		// Проверяем deadline если он был установлен
		if deadline, ok := ctx.Deadline(); ok {
			fmt.Printf("Request has deadline: %v\n", deadline)
		}

		c.JSON(200, map[string]string{
			"message": "External context received",
		})
	})

	// Пример 2: Middleware добавляет значения в контекст
	router.Use(func(c *aqylly.Context) {
		// Добавляем значение в контекст
		c.Set("request-start", time.Now())
		c.Set("user-id", "user123")
		c.Next()
	})

	router.GET("/with-values", func(c *aqylly.Context) {
		// Получаем значения из контекста
		startTime, _ := c.Get("request-start")
		userID, _ := c.Get("user-id")

		c.JSON(200, map[string]interface{}{
			"user_id":  userID,
			"duration": time.Since(startTime.(time.Time)),
		})
	})

	// Пример 3: Таймаут внутри хендлера
	router.GET("/timeout", func(c *aqylly.Context) {
		// Устанавливаем таймаут
		cancel, _ := c.WithTimeout(2 * time.Second)
		defer cancel()

		// Симулируем долгую операцию
		select {
		case <-time.After(3 * time.Second):
			c.JSON(200, map[string]string{"status": "completed"})
		case <-c.Done():
			// Контекст отменен (таймаут или отмена клиента)
			c.JSON(408, map[string]string{"error": c.Err().Error()})
		}
	})

	// Пример 4: Проброс контекста в другие сервисы
	router.GET("/call-service", func(c *aqylly.Context) {
		// Получаем контекст из запроса
		ctx := c.Context()

		// Делаем запрос к другому сервису с тем же контекстом
		req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com/api", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		c.JSON(200, map[string]string{"status": "service called with context"})
	})

	// Запускаем сервер
	log.Println("Server started on :8080")
	log.Println("")
	log.Println("Test with external context:")
	printClientExample()

	router.Run(":8080")
}

func printClientExample() {
	example := `
	// Клиент передает свой контекст:
	package main

	import (
		"context"
		"net/http"
		"time"
	)

	func main() {
		// Создаем контекст с значениями
		ctx := context.Background()
		ctx = context.WithValue(ctx, "trace-id", "abc-123-xyz")

		// Устанавливаем таймаут
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Создаем запрос с контекстом
		req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/external", nil)

		// Отправляем - сервер получит наш контекст!
		client := &http.Client{}
		resp, _ := client.Do(req)
		defer resp.Body.Close()
	}
`
	fmt.Println(example)
}
