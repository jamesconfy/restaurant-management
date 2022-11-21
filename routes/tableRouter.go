package routes

import (
	controller "restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/tables", controller.GetTables())
	incomingRoutes.POST("/api/tables", controller.CreateTable())
	incomingRoutes.GET("/api/tables/:table_id", controller.GetTable())
	incomingRoutes.PATCH("/api/tables/:table_id", controller.UpdateTable())
	incomingRoutes.DELETE("/api/tables/:table_id", controller.DeleteTable())
}
