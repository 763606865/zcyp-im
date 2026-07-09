package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const requestStartKey = "response_request_start"

type Meta struct {
	Timestamp    string `json:"timestamp"`
	ResponseTime int64  `json:"response_time"`
}

type Envelope struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
	Meta    Meta   `json:"meta"`
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(requestStartKey, time.Now())
		c.Next()
	}
}

func Success(c *gin.Context, status int, data any, message string) {
	c.JSON(status, Envelope{
		Code:    status,
		Data:    data,
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{
		Code:    status,
		Data:    nil,
		Message: message,
		Meta:    buildMeta(c),
	})
}

func AbortError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Envelope{
		Code:    status,
		Data:    nil,
		Message: message,
		Meta:    buildMeta(c),
	})
}

func buildMeta(c *gin.Context) Meta {
	now := time.Now()
	meta := Meta{
		Timestamp:    now.Format(time.RFC3339Nano),
		ResponseTime: 0,
	}

	if value, ok := c.Get(requestStartKey); ok {
		if startedAt, ok := value.(time.Time); ok {
			meta.ResponseTime = now.Sub(startedAt).Milliseconds()
		}
	}

	return meta
}

func OK(c *gin.Context, data any) {
	Success(c, http.StatusOK, data, "")
}

func Created(c *gin.Context, data any) {
	Success(c, http.StatusCreated, data, "")
}
