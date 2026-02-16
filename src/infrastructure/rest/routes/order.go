package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/order"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func OrderRoutes(router *gin.RouterGroup, controller order.IOrderController) {
	o := router.Group("/order")
	o.Use(middlewares.AuthJWTMiddleware())
	{
		o.POST("/", controller.CreateOrder)
		o.GET("/", controller.GetMyOrders)
		o.GET("/:id", controller.GetOrderByID)
		o.PUT("/:id/status", controller.UpdateOrderStatus)
	}
}
