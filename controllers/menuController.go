package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jamesconfy/restaurant-management/database"
	"github.com/jamesconfy/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allMenus []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		result, err := menuCollection.Find(context.TODO(), bson.M{}) //.Decode(&foods)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while listing all the menus!"})
			return
		}

		if err = result.All(ctx, &allMenus); err != nil {
			log.Fatal(err.Error())
		}
		c.JSON(http.StatusOK, allMenus)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		menu.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID()
		menu.Menu_ID = menu.ID.Hex()

		result, insertEr := menuCollection.InsertOne(ctx, &menu)
		if insertEr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Menu was not created!"})
			return
		}

		c.JSON(http.StatusAccepted, result)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		menuId := c.Param("menu_id")

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot find that menu"})
			return
		}
		c.JSON(http.StatusOK, menu)
	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		var updateObj primitive.D
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id": menuId}

		if menu.Start_Date != nil && menu.End_Date != nil {
			if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Kindly provide an acceptable time format"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})
		}

		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
		}

		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updated_at})

		// menuCollection.UpdateByID(ctx, menuId, updateObj)
		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := menuCollection.UpdateOne(ctx, filter, bson.D{{Key: "set", Value: updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update menu!"})
			return
		}

		c.JSON(http.StatusAccepted, result)
	}
}

func DeleteMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		menuId := c.Param("menu_id")

		err := menuCollection.FindOne(ctx, bson.M{"food_id": menuId}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find menu with that id"})
			return
		}

		result := menuCollection.FindOneAndDelete(ctx, bson.M{"food_id": menuId}).Decode(&menu)
		c.JSON(http.StatusAccepted, result)
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(check) && end.After(start)
}
