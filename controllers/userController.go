package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/jamesconfy/restaurant-management/database"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client , "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func Register() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func HashPassword(password string) string {

}

func VerifyPassword(userPassword, providedPassword string) bool {

}
