package routes

import (
	controller "restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/menu", controller.GetMenus())
	incomingRoutes.POST("/api/menus", controller.CreateMenu())
	incomingRoutes.GET("/api/menus/:menu_id", controller.GetMenu())
	incomingRoutes.PATCH("/api/menus/:menu_id", controller.UpdateMenu())
	incomingRoutes.DELETE("/api/menus/:menu_id", controller.DeleteMenu())
}
