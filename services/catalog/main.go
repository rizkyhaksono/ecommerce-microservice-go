// @title           Catalog Service API
// @version         1.0.0
// @description     Category and Product microservice

// @host            localhost:9090
// @BasePath        /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/pkg/middleware"
	"ecommerce-microservice-go/pkg/psql"
	"ecommerce-microservice-go/services/catalog/handler"
	"ecommerce-microservice-go/services/catalog/repository"
	"ecommerce-microservice-go/services/catalog/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "ecommerce-microservice-go/services/catalog/docs"
)

func main() {
	env := getEnvOrDefault("GO_ENV", "development")
	var log *logger.Logger
	var err error
	if env == "development" {
		log, err = logger.NewDevelopmentLogger()
	} else {
		log, err = logger.NewLogger()
	}
	if err != nil {
		panic(fmt.Errorf("error initializing logger: %w", err))
	}
	defer func() { _ = log.Log.Sync() }()

	log.Info("Starting Catalog Service")

	db, err := psql.ConnectDB(log)
	if err != nil {
		log.Panic("Failed to connect to database", zap.Error(err))
	}

	if err := psql.AutoMigrate(db, log, &repository.Category{}, &repository.Product{}); err != nil {
		log.Panic("Failed to migrate database", zap.Error(err))
	}

	catRepo := repository.NewCategoryRepository(db, log)
	prodRepo := repository.NewProductRepository(db, log)
	catUC := usecase.NewCategoryUseCase(catRepo, log)
	prodUC := usecase.NewProductUseCase(prodRepo, log)
	h := handler.NewHandler(catUC, prodUC, log)

	if env != "development" {
		log.SetupGinWithZapLogger()
	} else {
		log.SetupGinWithZapLoggerInDevelopment()
	}

	router := gin.New()
	router.Use(gin.Recovery(), cors.Default())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CommonHeaders)
	router.Use(log.GinZapLogger())

	v1 := router.Group("/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "catalog"})
	})

	v1.GET("/catalog/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Category routes
	cat := v1.Group("/category")
	cat.GET("/", h.GetAllCategories)
	cat.GET("/:id", h.GetCategoryByID)
	catAuth := cat.Group("")
	catAuth.Use(middleware.AuthJWTMiddleware())
	{
		catAuth.POST("/", h.NewCategory)
		catAuth.PUT("/:id", h.UpdateCategory)
		catAuth.DELETE("/:id", h.DeleteCategory)
	}

	// Product routes
	prod := v1.Group("/product")
	prod.GET("/", h.GetAllProducts)
	prod.GET("/:id", h.GetProductByID)
	prod.GET("/category/:categoryId", h.GetProductsByCategory)
	prodAuth := prod.Group("")
	prodAuth.Use(middleware.AuthJWTMiddleware())
	{
		prodAuth.POST("/", h.NewProduct)
		prodAuth.PUT("/:id", h.UpdateProduct)
		prodAuth.DELETE("/:id", h.DeleteProduct)
	}

	port := getEnvOrDefault("SERVER_PORT", "8082")
	log.Info("Catalog Service starting", zap.String("port", port))
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Panic("Server failed", zap.Error(err))
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
