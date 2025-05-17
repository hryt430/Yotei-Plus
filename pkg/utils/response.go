package utils

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse はエラーレスポンスを生成
func ErrorResponse(message string) gin.H {
	return gin.H{
		"success": false,
		"error":   message,
	}
}

// SuccessResponse は成功レスポンスを生成
func SuccessResponse(message string, data interface{}) gin.H {
	return gin.H{
		"success": true,
		"message": message,
		"data":    data,
	}
}
