package routes

import (
	controller "restaurant-management/controller"

	"github.com/gin-gonic/gin"
)

func InvoiceRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/api/invoices", controller.GetInvoices())
	incomingRoutes.POST("/api/invoices", controller.CreateInvoice())
	incomingRoutes.GET("/api/invoices/:invoice_id", controller.GetInvoice())
	incomingRoutes.PATCH("/api/invoices/:invoice_id", controller.UpdateInvoice())
	incomingRoutes.DELETE("/api/invoices/:invoice_id", controller.DeleteInvoice())
}
