package routes

import (
	"restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/users", controller.GetUsers())
	incomingRoutes.GET("/api/users/:user_id", controller.GetUser())
	incomingRoutes.POST("/api/register", controller.Register())
	incomingRoutes.POST("/api/login", controller.Login())
}
