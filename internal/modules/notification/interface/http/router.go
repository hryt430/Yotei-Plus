package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes は通知関連のルートを登録する
func RegisterRoutes(r *gin.RouterGroup, handler *NotificationHandler) {
	notificationRoutes := r.Group("/notifications")
	{
		// 通知作成
		notificationRoutes.POST("", handler.Create)

		// ユーザーIDによる通知一覧取得
		notificationRoutes.GET("/user/:userID", handler.GetAll)

		// ユーザーIDによる未読通知一覧取得
		notificationRoutes.GET("/user/:userID/unread", handler.GetUnread)

		// 通知ID指定の取得
		notificationRoutes.GET("/:id", handler.GetByID)

		// 通知を既読にする
		notificationRoutes.PUT("/:id/read", handler.MarkAsRead)
	}
}
