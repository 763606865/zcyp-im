package im

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"

	"zcyp-im/internal/auth"
	"zcyp-im/internal/service"
	wsgateway "zcyp-im/internal/ws"
)

type ConnectionHandler struct {
	appService   *service.AppService
	userService  *service.UserService
	tokenService *auth.TokenService
	hub          *wsgateway.Hub
}

type issueAccessTokenRequest struct {
	AppCode string `json:"app_code" binding:"required"`
	AppKey  string `json:"app_key" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

func NewConnectionHandler(
	appService *service.AppService,
	userService *service.UserService,
	tokenService *auth.TokenService,
	hub *wsgateway.Hub,
) *ConnectionHandler {
	return &ConnectionHandler{appService: appService, userService: userService, tokenService: tokenService, hub: hub}
}

func (h *ConnectionHandler) IssueAccessToken(c *gin.Context) {
	var req issueAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err := h.appService.ValidateApp(req.AppCode, req.AppKey)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrAppNotFound) {
			status = http.StatusUnauthorized
		}

		c.JSON(status, gin.H{
			"error": "invalid app credentials",
		})
		return
	}

	if _, err := h.userService.GetActiveUser(req.AppCode, req.UserID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		if errors.Is(err, service.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, err := h.tokenService.Issue(req.AppCode, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app_code":   req.AppCode,
		"user_id":    req.UserID,
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (h *ConnectionHandler) Connect(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token is required",
		})
		return
	}

	claims, err := h.tokenService.Parse(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	if _, err := h.userService.GetActiveUser(claims.AppCode, claims.UserID); err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		if errors.Is(err, service.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	client := h.hub.Register(conn, claims.AppCode, claims.UserID)
	client.Serve(context.Background())
}
