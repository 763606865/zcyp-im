package im

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/auth"
	"zcyp-im/internal/response"
)

const authClaimsKey = "auth_claims"

func AuthMiddleware(tokenService *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.AbortError(c, http.StatusUnauthorized, "missing bearer token")
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := tokenService.Parse(tokenString)
		if err != nil {
			response.AbortError(c, http.StatusUnauthorized, "invalid token")
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
