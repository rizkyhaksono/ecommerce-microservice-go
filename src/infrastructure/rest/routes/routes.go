package routes

import (
	"net/http"

	"github.com/gbrayhan/microservices-go/src/infrastructure/di"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func ApplicationRouter(router *gin.Engine, appContext *di.ApplicationContext) {
	// Swagger UI route — redirect /docs/ to /docs/index.html
	router.GET("/docs/*any", func(c *gin.Context) {
		any := c.Param("any")
		if any == "/" || any == "" {
			c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	v1 := router.Group("/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Service is running",
		})
	})

	AuthRoutes(v1, appContext.AuthController)
	UserRoutes(v1, appContext.UserController)
	CategoryRoutes(v1, appContext.CategoryController)
	ProductRoutes(v1, appContext.ProductController)
	OrderRoutes(v1, appContext.OrderController)
}

