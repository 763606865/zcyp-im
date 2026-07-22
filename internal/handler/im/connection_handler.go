package im

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"

	"zcyp-im/internal/auth"
	"zcyp-im/internal/response"
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
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.appService.ValidateApp(req.AppCode, req.AppKey)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrAppNotFound) {
			status = http.StatusUnauthorized
		}

		response.Error(c, status, "invalid app credentials")
		return
	}

	if _, err := h.userService.GetTokenEligibleUser(req.AppCode, req.UserID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		if errors.Is(err, service.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		if errors.Is(err, service.ErrSystemUserTokenNotAllowed) {
			status = http.StatusForbidden
		}
		response.Error(c, status, err.Error())
		return
	}

	token, expiresAt, err := h.tokenService.Issue(req.AppCode, req.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, gin.H{
		"app_code":   req.AppCode,
		"user_id":    req.UserID,
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (h *ConnectionHandler) Connect(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		log.Printf("ws connect rejected: missing token remote=%s", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "token is required")
		return
	}

	if !isWebSocketUpgrade(c.Request) {
		log.Printf("ws connect rejected: upgrade required remote=%s ua=%q connection=%q upgrade=%q", c.ClientIP(), c.Request.UserAgent(), c.GetHeader("Connection"), c.GetHeader("Upgrade"))
		response.Error(c, http.StatusBadRequest, "websocket upgrade required")
		return
	}

	claims, err := h.tokenService.Parse(tokenString)
	if err != nil {
		log.Printf("ws connect rejected: invalid token remote=%s err=%v", c.ClientIP(), err)
		response.Error(c, http.StatusUnauthorized, "invalid token")
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
		log.Printf("ws connect rejected: app_code=%s user_id=%s remote=%s err=%v status=%d", claims.AppCode, claims.UserID, c.ClientIP(), err, status)
		response.Error(c, status, err.Error())
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("ws connect accept failed: app_code=%s user_id=%s remote=%s err=%v", claims.AppCode, claims.UserID, c.ClientIP(), err)
		return
	}

	log.Printf("ws connect accepted: app_code=%s user_id=%s remote=%s", claims.AppCode, claims.UserID, c.ClientIP())

	client := h.hub.Register(conn, claims.AppCode, claims.UserID)
	client.Serve(context.Background())
}

func isWebSocketUpgrade(r *http.Request) bool {
	if !headerContainsToken(r.Header, "Connection", "upgrade") {
		return false
	}
	if !headerContainsToken(r.Header, "Upgrade", "websocket") {
		return false
	}
	if r.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}
	return true
}

func headerContainsToken(header http.Header, key, token string) bool {
	values := header.Values(key)
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}
