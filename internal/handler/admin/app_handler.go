package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/service"
)

type AppHandler struct {
	appService *service.AppService
}

func NewAppHandler(appService *service.AppService) *AppHandler {
	return &AppHandler{appService: appService}
}

func (h *AppHandler) ListApps(c *gin.Context) {
	items, err := h.appService.ListApps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

func (h *AppHandler) CreateApp(c *gin.Context) {
	var input service.CreateAppInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	app, err := h.appService.CreateApp(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, app)
}

func (h *AppHandler) GetApp(c *gin.Context) {
	app, err := h.appService.GetApp(c.Param("app_code"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrAppNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, app)
}
