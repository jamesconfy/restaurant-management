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
		defer cancel()
		result, err := orderCollection.Find(context.TODO(), bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order items!"})
			return
		}

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
		defer cancel()

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

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

		c.JSON(http.StatusAccepted, insertedOrderItems)

	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItem models.OrderItem
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		orderItemId := c.Param("order_item_id")

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)

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
		defer cancel()

		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderItemId := c.Param("order_item_id")
		filter := bson.M{"order_item_id": orderItemId}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{Key: "quantity", Value: &orderItem.Quantity})
		}

		if orderItem.Unit_Price != nil {
			updateObj = append(updateObj, bson.E{Key: "unit_price", Value: &orderItem.Unit_Price})
		}

		if orderItem.Food_ID != nil {
			updateObj = append(updateObj, bson.E{Key: "food_id", Value: &orderItem.Food_ID})
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(ctx, filter, bson.D{{Key: "set", Value: updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update order item!"})
			return
		}

		c.JSON(http.StatusAccepted, result)

	}
}

func DeleteOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItem models.OrderItem
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		orderItemId := c.Param("order_item_id")

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find order item with that id"})
			return
		}

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
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}
	lookupFoodStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "food"}, {Key: "localField", Value: "food_id"}, {Key: "foreignField", Value: "food_id"}, {Key: "as", Value: "food"}}}}
	unwindFoodStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$food"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "order"}, {Key: "localField", Value: "order_id"}, {Key: "foreignField", Value: "order_id"}, {Key: "as", Value: "order"}}}}
	unwindOrderStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$order"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "table"}, {Key: "localField", Value: "order.table_id"}, {Key: "foreignField", Value: "table_id"}, {Key: "as", Value: "table"}}}}
	unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$table"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	projectStage := bson.D{{
		Key: "$project", Value: bson.D{
			{Key: "id", Value: 0},
			{Key: "amount", Value: "$food.price"},
			{Key: "total_count", Value: 1},
			{Key: "food_name", Value: "$food.name"},
			{Key: "food_image", Value: "$food.food_image"},
			{Key: "table_number", Value: "$table.table_number"},
			{Key: "table_id", Value: "$table.table_id"},
			{Key: "order_id", Value: "$order.order_id"},
			{Key: "price", Value: "$food.price"},
			{Key: "quantity", Value: 1},
		},
	}}

	groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "order_id", Value: "$order.order_id"}, {Key: "table_id", Value: "$table.table_id"}, {Key: "table_number", Value: "$table.table_number"}}}, {Key: "payment_due", Value: bson.D{{Key: "$sum", Value: "$amount"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}

	projectStage2 := bson.D{{
		Key: "$project", Value: bson.D{
			{Key: "id", Value: 0},
			{Key: "payment_due", Value: 1},
			{Key: "total_count", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
		},
	}}

	result, err := orderCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupFoodStage, unwindFoodStage, lookupOrderStage, unwindOrderStage, lookupTableStage, unwindTableStage, projectStage, groupStage, projectStage2})
	if err != nil {
		panic(err.Error())
	}

	if err = result.All(ctx, &OrderItems); err != nil {
		log.Fatal(err.Error())
	}
	return OrderItems, err
}
