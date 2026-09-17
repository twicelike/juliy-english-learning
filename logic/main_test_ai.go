package main

import (
	"context"
	"fmt"
	"os"

	"github.com/english/backend/internal/ai/client"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	c, _ := client.NewOpenRouterClient(apiKey, "openrouter/free")
	var result map[string]interface{}
	err := c.GenerateJSON(context.Background(), "say hello in JSON format {\"hello\": \"world\"}", nil, &result)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result)
	}
}
