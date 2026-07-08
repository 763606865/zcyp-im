package im

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/auth"
)

const authClaimsKey = "auth_claims"

func AuthMiddleware(tokenService *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := tokenService.Parse(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(authClaimsKey, claims)
		c.Next()
	}
}

func mustClaims(c *gin.Context) auth.Claims {
	value, ok := c.Get(authClaimsKey)
	if !ok {
		return auth.Claims{}
	}

	claims, _ := value.(auth.Claims)
	return claims
}
