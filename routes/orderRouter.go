package routes

import (
	controller "github.com/jamesconfy/restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func OrderRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/orders", controller.GetOrders())
	incomingRoutes.POST("/api/orders", controller.CreateOrder())
	incomingRoutes.GET("/api/orders/:order_id", controller.GetOrder())
	incomingRoutes.PATCH("/api/orders/:order_id", controller.UpdateOrder())
	incomingRoutes.DELETE("/api/orders/:order_id", controller.DeleteOrder())
}
