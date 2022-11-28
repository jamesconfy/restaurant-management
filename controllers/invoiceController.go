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

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

type InvoiceViewFormat struct {
	Invoice_ID       string
	Order_ID         string
	Payment_Method   string
	Payment_Status   *string
	Payment_Due      interface{}
	Payment_Due_Date time.Time
	Table_Number     interface{}
	Order_Details    interface{}
}

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allInvoices []bson.M
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching invoices!"})
			return
		}
		defer cancel()

		if err = result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err.Error())
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice models.Invoice
		var order models.Order
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		defer cancel()

		err := orderCollection.FindOne(ctx, bson.M{"table_id": invoice.Order_ID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Order was not found"})
			return
		}
		defer cancel()

		method := "CASH"
		if invoice.Payment_Method == nil {
			invoice.Payment_Method = &method
		}

		status := "PENDING"
		if invoice.Payment_Status == nil {
			invoice.Payment_Status = &status
		}

		invoice.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Payment_Due_Date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_ID = invoice.ID.Hex()

		result, insertEr := invoiceCollection.InsertOne(ctx, &invoice)
		if insertEr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invoice was not created!"})
			return
		}
		defer cancel()

		c.JSON(http.StatusAccepted, result)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot find that invoice item"})
			return
		}
		defer cancel()

		var invoiceView InvoiceViewFormat
		allOrdersItem, _ := ItemsByOrder(invoice.Order_ID)

		invoiceView.Order_ID = invoice.Order_ID
		invoiceView.Payment_Due_Date = invoice.Payment_Due_Date
		invoiceView.Payment_Method = "null"
		if invoice.Payment_Method != nil {
			invoiceView.Payment_Method = *invoice.Payment_Method
		}

		invoiceView.Invoice_ID = invoice.Invoice_ID
		invoiceView.Payment_Status = *&invoice.Payment_Status
		invoiceView.Payment_Due = allOrdersItem[0]["payment_due"]
		invoiceView.Table_Number = allOrdersItem[0]["table_number"]
		invoiceView.Order_Details = allOrdersItem[0]["invoice_items"]

		c.JSON(http.StatusOK, invoiceView)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice models.Invoice
		var updateObj primitive.D
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		invoiceId := c.Param("invoice_id")
		filter := bson.M{"invoice_id": invoiceId}

		method := "CASH"
		if invoice.Payment_Method != nil {
			updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_Method})
		} else {
			invoice.Payment_Method = &method
		}

		status := "PENDING"
		if invoice.Payment_Status != nil {
			updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_Status})
		} else {
			invoice.Payment_Status = &status
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := invoiceCollection.UpdateOne(ctx, filter, bson.D{{"set", updateObj}}, &opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update invoice!"})
			return
		}
		defer cancel()
		c.JSON(http.StatusAccepted, result)

	}
}

func DeleteInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot find invoice with that id"})
			return
		}
		defer cancel()

		result := invoiceCollection.FindOneAndDelete(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		c.JSON(http.StatusAccepted, result)
	}
}
