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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allTables []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		result, err := tableCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tables!"})
			return
		}
		defer cancel()

		if err = result.All(ctx, &allTables); err != nil {
			log.Fatal(err.Error())
		}
		c.JSON(http.StatusOK, allTables)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		defer cancel()

		table.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.ID = primitive.NewObjectID()
		table.Table_ID = table.ID.Hex()

		result, insertEr := tableCollection.InsertOne(ctx, &table)
		if insertEr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Table was not created!"})
			return
		}
		defer cancel()

		c.JSON(http.StatusAccepted, result)
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		tableId := c.Param("table_id")
		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot find that table"})
			return
		}
		c.JSON(http.StatusOK, table)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var updateObj primitive.D
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		tableId := c.Param("table_id")
		filter := bson.M{"table_id": tableId}

		if table.Number_Of_Guests != nil {
			updateObj = append(updateObj, bson.E{"number_of_guests", &table.Number_Of_Guests})
		}

		if table.Table_Number != nil {
			updateObj = append(updateObj, bson.E{"number_of_guests", &table.Table_Number})
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := tableCollection.UpdateOne(ctx, filter, bson.D{{"set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update table!"})
			return
		}
		defer cancel()
		c.JSON(http.StatusAccepted, result)
	}
}

func DeleteTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		tableId := c.Param("table_id")
		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find table with that id"})
			return
		}
		defer cancel()

		result := tableCollection.FindOneAndDelete(ctx, bson.M{"table_id": tableId}).Decode(&table)
		c.JSON(http.StatusAccepted, result)
	}
}
