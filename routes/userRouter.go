package routes

import (
	controller "github.com/jamesconfy/restaurant-management/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/users", controller.GetUsers())
	incomingRoutes.GET("/api/users/:user_id", controller.GetUser())
	incomingRoutes.POST("/api/register", controller.Register())
	incomingRoutes.POST("/api/login", controller.Login())
}
