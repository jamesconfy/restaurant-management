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

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

type OrderItemPack struct {
	Table_ID    *string
	Order_Items []models.OrderItem
}

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allOrderItems []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		result, err := orderCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order items!"})
			return
		}
		defer cancel()

		if err = result.All(ctx, &allOrderItems); err != nil {
			log.Fatal(err.Error())
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order
		var orderItemPack OrderItemPack
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		defer cancel()

		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Table_ID = orderItemPack.Table_ID

		orderItemsToBeInserted := []interface{}{}
		orderId := OrderItemCreator(order)

		for _, orderItem := range orderItemPack.Order_Items {
			orderItem.Order_ID = &orderId

			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}
			defer cancel()

			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_Item_ID = orderItem.ID.Hex()
			num := toFixed(*orderItem.Unit_Price, 2)
			orderItem.Unit_Price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}
		// order.Order_Date, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		insertedOrderItems, insertEr := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
		if insertEr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Order item was not created!"})
			return
		}
		defer cancel()

		c.JSON(http.StatusAccepted, insertedOrderItems)

	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItem models.OrderItem
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		orderItemId := c.Param("order_item_id")

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot find that order item"})
			return
		}
		c.JSON(http.StatusOK, orderItem)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItem models.OrderItem
		var updateObj primitive.D
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		orderItemId := c.Param("order_item_id")
		filter := bson.M{"order_item_id": orderItemId}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", &orderItem.Quantity})
		}

		if orderItem.Unit_Price != nil {
			updateObj = append(updateObj, bson.E{"unit_price", &orderItem.Unit_Price})
		}

		if orderItem.Food_ID != nil {
			updateObj = append(updateObj, bson.E{"food_id", &orderItem.Food_ID})
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(ctx, filter, bson.D{{"set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update order item!"})
			return
		}
		defer cancel()
		c.JSON(http.StatusAccepted, result)

	}
}

func DeleteOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		orderItemId := c.Param("order_item_id")
		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find order item with that id"})
			return
		}
		defer cancel()

		result := orderItemCollection.FindOneAndDelete(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		c.JSON(http.StatusAccepted, result)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while listing order item by order id"})
			return
		}

		c.JSON(http.StatusOK, allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}
	lookupFoodStage := bson.D{{"$lookup", bson.D{{"from", "food"}, {"localField", "food_id"}, {"foreignField", "food_id"}, {"as", "food"}}}}
	unwindFoodStage := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupOrderStage := bson.D{{"$lookup", bson.D{{"from", "order"}, {"localField", "order_id"}, {"foreignField", "order_id"}, {"as", "order"}}}}
	unwindOrderStage := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupTableStage := bson.D{{"$lookup", bson.D{{"from", "table"}, {"localField", "order.table_id"}, {"foreignField", "table_id"}, {"as", "table"}}}}
	unwindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullAndEmptyArrays", true}}}}

	projectStage := bson.D{{
		"$project", bson.D{
			{"id", 0},
			{"amount", "$food.price"},
			{"total_count", 1},
			{"food_name", "$food.name"},
			{"food_image", "$food.food_image"},
			{"table_number", "$table.table_number"},
			{"table_id", "$table.table_id"},
			{"order_id", "$order.order_id"},
			{"price", "$food.price"},
			{"quantity", 1},
		},
	}}

	groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"order_id", "$order.order_id"}, {"table_id", "$table.table_id"}, {"table_number", "$table.table_number"}}}, {"payment_due", bson.D{{"$sum", "$amount"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"order_items", bson.D{{"$push", "$$ROOT"}}}}}}

	projectStage2 := bson.D{{
		"$project", bson.D{
			{"id", 0},
			{"payment_due", 1},
			{"total_count", 1},
			{"table_number", "$_id.table_number"},
			{"order_items", 1},
		},
	}}

	result, err := orderCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupFoodStage, unwindFoodStage, lookupOrderStage, unwindOrderStage, lookupTableStage, unwindTableStage, projectStage, groupStage, projectStage2})
	if err != nil {
		panic(err.Error())
	}
	defer cancel()

	if err = result.All(ctx, &OrderItems); err != nil {
		log.Fatal(err.Error())
	}
	return OrderItems, err
}
