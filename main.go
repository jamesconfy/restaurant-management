package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jamesconfy/restaurant-management/database"
	"github.com/jamesconfy/restaurant-management/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	// router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.InvoiceRoutes(router)
	routes.OrderItemsRoutes(router)

	router.Run(":" + port)

}
