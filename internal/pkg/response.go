package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: msg,
	})
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, 40000, msg)
}

func Unauthorized(c *gin.Context, code int, msg string) {
	Error(c, http.StatusUnauthorized, code, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, 40300, msg)
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, 40400, msg)
}

func TooManyRequests(c *gin.Context, msg string) {
	Error(c, http.StatusTooManyRequests, 42900, msg)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, 50000, "服务器内部错误")
}
