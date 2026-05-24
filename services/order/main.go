// @title           Order Service API
// @version         1.0.0
// @description     Order management microservice

// @host            localhost:9090
// @BasePath        /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/pkg/middleware"
	"ecommerce-microservice-go/pkg/observability"
	"ecommerce-microservice-go/pkg/psql"
	"ecommerce-microservice-go/services/order/handler"
	"ecommerce-microservice-go/services/order/repository"
	"ecommerce-microservice-go/services/order/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"

	_ "ecommerce-microservice-go/services/order/docs"
)

func main() {
	env := getEnvOrDefault("GO_ENV", "development")
	serviceName := getEnvOrDefault("OTEL_SERVICE_NAME", "order-service")

	// OpenTelemetry: traces + metrics + logs over OTLP to the collector.
	otelProviders, otelShutdown, err := observability.InitOTel(context.Background(), serviceName)
	if err != nil {
		panic(fmt.Errorf("error initializing OpenTelemetry: %w", err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = otelShutdown(ctx)
	}()

	log, err := logger.NewLoggerWithOTel(otelProviders.LoggerProvider, "ecommerce-microservice-go/services/order", env == "development")
	if err != nil {
		panic(fmt.Errorf("error initializing logger: %w", err))
	}
	defer func() { _ = log.Log.Sync() }()

	log.Info("Starting Order Service")

	db, err := psql.ConnectDB(log)
	if err != nil {
		log.Panic("Failed to connect to database", zap.Error(err))
	}

	if err := psql.AutoMigrate(db, log, &repository.Order{}, &repository.OrderItem{}); err != nil {
		log.Panic("Failed to migrate database", zap.Error(err))
	}

	orderRepo := repository.NewOrderRepository(db, log)
	orderUC := usecase.NewOrderUseCase(orderRepo, log)
	h := handler.NewHandler(orderUC, log)

	if env != "development" {
		log.SetupGinWithZapLogger()
	} else {
		log.SetupGinWithZapLoggerInDevelopment()
	}

	router := gin.New()
	router.Use(otelgin.Middleware(serviceName)) // first: creates the request span/context
	router.Use(gin.Recovery(), cors.Default())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CommonHeaders)
	router.Use(log.GinZapLogger())

	v1 := router.Group("/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "order"})
	})

	v1.GET("/order/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// All order routes require auth
	order := v1.Group("/order")
	order.Use(middleware.AuthJWTMiddleware())
	{
		order.GET("/", h.GetAllOrders)
		order.POST("/", h.NewOrder)
		order.GET("/:id", h.GetOrderByID)
		order.PUT("/:id/status", h.UpdateOrderStatus)
	}

	port := getEnvOrDefault("SERVER_PORT", "8083")
	log.Info("Order Service starting", zap.String("port", port))
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
