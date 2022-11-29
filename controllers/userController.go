package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jamesconfy/restaurant-management/database"
	helper "github.com/jamesconfy/restaurant-management/helpers"
	"github.com/jamesconfy/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allUsers []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 20
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		// startIndex, err := strconv.Atoi(c.Query("startIndex"))
		// if err != nil {

		// }

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while getting all users"})
			return
		}

		if err := result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		userId := c.Param("user_id")

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot find that user"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists!"})
			return
		}

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Phone number already exists1"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password
		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllToken(*user.Email, user.User_ID, *user.First_Name, *user.Last_Name)
		user.Token = &token
		user.Refresh_Token = &refreshToken

		result, err := userCollection.InsertOne(ctx, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not created!"})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var foundUser models.User
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "We can't seem to find that user in our database!"})
			return
		}

		if !VerifyPassword(*foundUser.Password, *user.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect password, please try again"})
			return
		}

		token, refreshToken, _ := helper.GenerateAllToken(*foundUser.Email, foundUser.User_ID, *foundUser.First_Name, *foundUser.Last_Name)
		err = helper.UpdateToken(token, refreshToken, foundUser.User_ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong!"})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err.Error())
	}

	return string(bytes)
}

func VerifyPassword(userPassword, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	check := true
	if err != nil {
		check = false
	}

	return check
}
