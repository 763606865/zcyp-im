package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/response"
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
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, items)
}

func (h *AppHandler) CreateApp(c *gin.Context) {
	var input service.CreateAppInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app, err := h.appService.CreateApp(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, app)
}

func (h *AppHandler) GetApp(c *gin.Context) {
	app, err := h.appService.GetApp(c.Param("app_code"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrAppNotFound) {
			status = http.StatusNotFound
		}

		response.Error(c, status, err.Error())
		return
	}

	response.OK(c, app)
}
