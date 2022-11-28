package controllers

import (
	"context"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jamesconfy/restaurant-management/database"
	"github.com/jamesconfy/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		var allFoods []bson.M
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 20
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{{"id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"sum", 1}}}, {"$data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{{
			"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			},
		}}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while listing all the foods!"})
			return
		}
		defer cancel()

		if err = result.All(ctx, &allFoods); err != nil {
			log.Fatal(err.Error())
		}
		c.JSON(http.StatusOK, allFoods)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_ID}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Menu was not found"})
			return
		}

		defer cancel()
		food.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_ID = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, insertEr := foodCollection.InsertOne(ctx, &food)
		if insertEr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Food was not created!"})
			return
		}

		defer cancel()
		c.JSON(http.StatusAccepted, result)
	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot find that food item"})
			return
		}
		c.JSON(http.StatusOK, food)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var food models.Food
		var menu models.Menu
		var updateObj primitive.D
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		foodId := c.Param("food_id")
		filter := bson.M{"food_id": foodId}

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{"name", food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{"price", food.Price})
		}

		if food.Food_Image != nil {
			updateObj = append(updateObj, bson.E{"food_image", food.Food_Image})
		}

		if food.Menu_ID != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_ID}).Decode(&menu)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Menu item not found"})
				return
			}

			updateObj = append(updateObj, bson.E{"menu_id", food.Menu_ID})
		}
		defer cancel()

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := menuCollection.UpdateOne(ctx, filter, bson.D{{"set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update food!"})
			return
		}
		defer cancel()
		c.JSON(http.StatusAccepted, result)
	}
}

func DeleteFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find food with that id"})
			return
		}
		defer cancel()

		result := foodCollection.FindOneAndDelete(ctx, bson.M{"food_id": foodId}).Decode(&food)
		c.JSON(http.StatusAccepted, result)
	}
}

func rounded(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	ouptut := math.Pow(10, float64(precision))
	return float64(rounded(num * ouptut))
}
