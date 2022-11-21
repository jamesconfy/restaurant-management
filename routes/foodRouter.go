package routes

import (
	controller "github.com/jamesconfy/restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func FoodRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/foods", controller.GetFoods())
	incomingRoutes.POST("/api/foods", controller.CreateFood())
	incomingRoutes.GET("/api/foods/:food_id", controller.GetFood())
	incomingRoutes.PATCH("/api/foods/:food_id", controller.UpdateFood())
	incomingRoutes.DELETE("/api/foods/:food_id", controller.DeleteFood())
}
