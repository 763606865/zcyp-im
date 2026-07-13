package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/model"
	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

const appContextKey = "api_app"

func AppAuthMiddleware(appService *service.AppService) gin.HandlerFunc {
	return func(c *gin.Context) {
		appCode := strings.TrimSpace(c.GetHeader("X-App-Code"))
		appKey := strings.TrimSpace(c.GetHeader("X-App-Key"))
		if appCode == "" || appKey == "" {
			response.AbortError(c, http.StatusUnauthorized, "missing app credentials")
			return
		}

		app, err := appService.ValidateApp(appCode, appKey)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, service.ErrAppNotFound) {
				status = http.StatusUnauthorized
			}
			response.AbortError(c, status, "invalid app credentials")
			return
		}

		c.Set(appContextKey, app)
		c.Next()
	}
}

func mustApp(c *gin.Context) model.App {
	value, ok := c.Get(appContextKey)
	if !ok {
		return model.App{}
	}

	app, _ := value.(model.App)
	return app
}
