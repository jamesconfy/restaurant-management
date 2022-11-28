package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jamesconfy/restaurant-management/database"
	"github.com/jamesconfy/restaurant-management/middleware"
	"github.com/jamesconfy/restaurant-management/routes"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Authentication())

	routes.UserRoutes(router)
	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.InvoiceRoutes(router)
	routes.OrderItemsRoutes(router)

	router.Run("" + port)

}
