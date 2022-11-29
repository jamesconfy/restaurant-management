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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allOrders []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		result, err := orderCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching orders!"})
			return
		}

		if err = result.All(ctx, &allOrders); err != nil {
			log.Fatal(err.Error())
		}
		c.JSON(http.StatusOK, allOrders)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		var table models.Table
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_ID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Table was not found"})
			return
		}

		order.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.Order_ID = order.ID.Hex()

		result, insertEr := orderCollection.InsertOne(ctx, &order)
		if insertEr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Order was not created!"})
			return
		}

		c.JSON(http.StatusAccepted, result)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		orderId := c.Param("order_id")

		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot find that order"})
			return
		}
		c.JSON(http.StatusOK, order)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		var table models.Table
		var updateObj primitive.D
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderId := c.Param("order_id")
		filter := bson.M{"order_id": orderId}

		if order.Table_ID != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_ID}).Decode(&table)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Table does not exist"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: order.Table_ID})
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(ctx, filter, bson.D{{Key: "set", Value: updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update order!"})
			return
		}
		defer cancel()
		c.JSON(http.StatusAccepted, result)
	}
}

func DeleteOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		orderId := c.Param("order_id")

		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find order with that id"})
			return
		}

		result := orderCollection.FindOneAndDelete(ctx, bson.M{"order_id": orderId}).Decode(&order)
		c.JSON(http.StatusAccepted, result)
	}
}

func OrderItemCreator(order models.Order) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()
	order.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_ID = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)
	return order.Order_ID
}
