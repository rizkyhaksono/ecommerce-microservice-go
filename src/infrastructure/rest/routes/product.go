package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/product"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func ProductRoutes(router *gin.RouterGroup, controller product.IProductController) {
	p := router.Group("/product")
	// Public: list and detail
	p.GET("/", controller.GetAllProducts)
	p.GET("/:id", controller.GetProductByID)
	p.GET("/category/:categoryId", controller.GetProductsByCategory)
	// Protected: create, update, delete
	p.Use(middlewares.AuthJWTMiddleware())
	{
		p.POST("/", controller.NewProduct)
		p.PUT("/:id", controller.UpdateProduct)
		p.DELETE("/:id", controller.DeleteProduct)
	}
}
