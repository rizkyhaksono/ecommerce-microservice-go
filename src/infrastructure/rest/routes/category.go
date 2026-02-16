package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/category"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func CategoryRoutes(router *gin.RouterGroup, controller category.ICategoryController) {
	c := router.Group("/category")
	// Public: list and detail
	c.GET("/", controller.GetAllCategories)
	c.GET("/:id", controller.GetCategoryByID)
	// Protected: create, update, delete
	c.Use(middlewares.AuthJWTMiddleware())
	{
		c.POST("/", controller.NewCategory)
		c.PUT("/:id", controller.UpdateCategory)
		c.DELETE("/:id", controller.DeleteCategory)
	}
}
