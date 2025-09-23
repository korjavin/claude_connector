package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Use 'Bearer <token>'"})
			return
		}

		tokenString := parts[1]

		keySet, err := jwk.Fetch(context.Background(), "http://hydra:4444/.well-known/jwks.json")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch JWKS"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid header not found")
			}
			keys, ok := keySet.LookupKeyID(kid)
			if !ok {
				return nil, fmt.Errorf("key with kid %s not found", kid)
			}
			var pubkey interface{}
			if err := keys.Raw(&pubkey); err != nil {
				return nil, fmt.Errorf("failed to get raw public key")
			}
			return pubkey, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}
