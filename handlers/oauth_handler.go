package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func init() {
	gob.Register(&oauth2.Token{})
}

// OAuth2Config holds the configuration for the OAuth2 client
type OAuth2Config struct {
	*oauth2.Config
}

// NewOAuth2Config creates a new OAuth2Config
func NewOAuth2Config(clientID, clientSecret, redirectURL string) *OAuth2Config {
	return &OAuth2Config{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://claude.ai/oauth/authorize",
				TokenURL: "https://claude.ai/oauth/token",
			},
			Scopes: []string{"profile"},
		},
	}
}

// HandleLogin redirects the user to the OAuth2 provider's login page
func (conf *OAuth2Config) HandleLogin(c *gin.Context, store sessions.Store) {
	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
		return
	}

	// Generate a random state string to prevent CSRF attacks
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}
	state := base64.StdEncoding.EncodeToString(b)
	session.Values["state"] = state
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	url := conf.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleCallback handles the callback from the OAuth2 provider
func (conf *OAuth2Config) HandleCallback(c *gin.Context, store sessions.Store) {
	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
		return
	}

	// Check that the state matches
	state := session.Values["state"]
	if state == nil || state.(string) != c.Query("state") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid state"})
		return
	}

	// Exchange the authorization code for a token
	code := c.Query("code")
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	session.Values["token"] = token
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated"})
}
