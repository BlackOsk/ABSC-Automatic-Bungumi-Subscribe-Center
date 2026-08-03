package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

// Success 成功统一回包 (Code: 0)
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Message: "success",
		Code:    200,
		Data:    data,
	})
}

// Fail 失败统一回包 (Code: > 0)
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}
