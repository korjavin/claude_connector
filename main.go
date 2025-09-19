package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/user/claude-connector/handlers"
	"github.com/user/claude-connector/middleware"
)

func main() {
	port := os.Getenv("MCP_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("API_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: API_SECRET_KEY environment variable not set.")
	}

	csvPath := os.Getenv("CSV_FILE_PATH")
	if csvPath == "" {
		log.Fatal("FATAL: CSV_FILE_PATH environment variable not set.")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	mcpGroup := router.Group("/mcp")
	{
		mcpGroup.Use(middleware.AuthMiddleware(apiKey))
		mcpGroup.POST("", handlers.MCPHandler(csvPath))
	}

	log.Printf("Starting MCP server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
