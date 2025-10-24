package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/maksat/aqylly"
)

func main() {
	// Create HTTP/2 Cleartext client
	client := &http.Client{
		Transport: aqylly.NewH2CTransport(),
	}

	// Test basic GET request
	resp, err := client.Get("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Protocol: %s\n", resp.Proto)
	fmt.Printf("Response: %s\n", body)
	fmt.Println()

	// Test API endpoint
	resp2, err := client.Get("http://localhost:8080/api/data")
	if err != nil {
		log.Fatal(err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("API Response: %s\n", body2)
}
