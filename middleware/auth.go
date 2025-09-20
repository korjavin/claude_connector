package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// AuthMiddleware validates the OAuth2 token from the session
func AuthMiddleware(store sessions.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := store.Get(c.Request, "session-name")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
			return
		}

		token, ok := session.Values["token"].(*oauth2.Token)
		if !ok || !token.Valid() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated or token expired. Please visit /auth/login"})
			return
		}

		// You could add more validation here, e.g., checking token claims.
		// You can also pass the token to the context for use in handlers.
		c.Set("token", token)

		c.Next()
	}
}
