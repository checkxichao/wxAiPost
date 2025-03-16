// src/response/response.go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(ctx *gin.Context, message string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

func Error(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}
