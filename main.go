package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/korjavin/claude_connector/handlers"
	"github.com/korjavin/claude_connector/middleware"
)

// CommitSHA will be set at build time via ldflags
var CommitSHA = "unknown"

var store *sessions.CookieStore

func main() {
	port := os.Getenv("MCP_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	clientID := os.Getenv("CLAUDE_OAUTH_CLIENT_ID")
	if clientID == "" {
		log.Fatal("FATAL: CLAUDE_OAUTH_CLIENT_ID environment variable not set.")
	}

	clientSecret := os.Getenv("CLAUDE_OAUTH_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("FATAL: CLAUDE_OAUTH_CLIENT_SECRET environment variable not set.")
	}

	redirectURL := os.Getenv("CLAUDE_OAUTH_REDIRECT_URL")
    if redirectURL == "" {
        log.Fatal("FATAL: CLAUDE_OAUTH_REDIRECT_URL environment variable not set.")
    }

	csvPath := os.Getenv("CSV_FILE_PATH")
	if csvPath == "" {
		log.Fatal("FATAL: CSV_FILE_PATH environment variable not set.")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("FATAL: SESSION_SECRET environment variable not set.")
	}
	store = sessions.NewCookieStore([]byte(sessionSecret))


	oauthConfig := handlers.NewOAuth2Config(clientID, clientSecret, redirectURL)

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

	authGroup := router.Group("/auth")
	{
		authGroup.GET("/login", func(c *gin.Context) {
			oauthConfig.HandleLogin(c, store)
		})
		authGroup.GET("/callback", func(c *gin.Context) {
			oauthConfig.HandleCallback(c, store)
		})
	}

	mcpGroup := router.Group("/mcp")
	{
		mcpGroup.Use(middleware.AuthMiddleware(store))
		mcpGroup.POST("", handlers.MCPHandler(csvPath))
	}

	log.Printf("Starting MCP server on port %s (commit: %s)", port, CommitSHA)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
