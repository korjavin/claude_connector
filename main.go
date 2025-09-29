package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/korjavin/claude_connector/handlers"
	"github.com/korjavin/claude_connector/middleware"
)

// CommitSHA will be set at build time via ldflags
var CommitSHA = "unknown"

func main() {
	port := os.Getenv("MCP_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	csvPath := os.Getenv("CSV_FILE_PATH")
	if csvPath == "" {
		log.Fatal("FATAL: CSV_FILE_PATH environment variable not set.")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint (no authentication required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"commit":    CommitSHA,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	mcpGroup := router.Group("/mcp")
	{
		mcpGroup.Use(middleware.AuthMiddleware())
		mcpGroup.POST("", handlers.MCPHandler(csvPath))
	}

	log.Printf("Starting MCP server on port %s (commit: %s)", port, CommitSHA)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
