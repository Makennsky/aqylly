package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Makennsky/aqylly"
)

func main() {
	// Create router for internal microservice
	router := aqylly.Default()

	// Configure HTTP/2 settings
	router.HTTP2Config.MaxConcurrentStreams = 100

	router.GET("/", func(c *aqylly.Context) {
		c.JSON(200, map[string]interface{}{
			"message":  "HTTP/2 Cleartext (h2c) - No TLS!",
			"protocol": c.Request.Proto,
		})
	})

	router.GET("/api/data", func(c *aqylly.Context) {
		c.JSON(200, map[string]interface{}{
			"data": []string{"item1", "item2", "item3"},
		})
	})

	router.POST("/api/process", func(c *aqylly.Context) {
		var payload map[string]interface{}
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(400, map[string]string{"error": "Invalid JSON"})
			return
		}

		c.JSON(200, map[string]interface{}{
			"status":  "processed",
			"payload": payload,
		})
	})

	log.Println("Starting HTTP/2 Cleartext (h2c) server on :8080")
	log.Println("This is useful for internal microservices without TLS overhead")
	log.Println("")
	log.Println("Test with curl:")
	log.Println("  curl --http2-prior-knowledge http://localhost:8080")
	log.Println("")
	log.Println("Or with Go client:")
	printGoClientExample()

	if err := router.RunH2C(":8080"); err != nil {
		log.Fatal(err)
	}
}

func printGoClientExample() {
	example := `
	package main

	import (
		"fmt"
		"io"
		"net/http"
		"github.com/Makennsky/aqylly"
	)

	func main() {
		client := &http.Client{
			Transport: aqylly.NewH2CTransport(),
		}

		resp, err := client.Get("http://localhost:8080")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Protocol: %s\n", resp.Proto)
		fmt.Printf("Body: %s\n", body)
	}
`
	fmt.Println(example)
}
