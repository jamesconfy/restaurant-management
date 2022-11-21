package routes

import (
	controller "restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func OrderItemsRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/orderItems", controller.GetOrderItems())
	incomingRoutes.POST("/api/orderItems", controller.CreateOrderItem())
	incomingRoutes.GET("/api/orderItems/:orderItem_id", controller.GetOrderItem())
	incomingRoutes.PATCH("/api/orderItems/:orderItem_id", controller.UpdateOrderItem())
	incomingRoutes.DELETE("/api/orderItems/:orderItem_id", controller.DeleteOrderItem())
	incomingRoutes.GET("/api/orderItems/order/:order_id", controller.GetOrderItemsByOrder())
}
