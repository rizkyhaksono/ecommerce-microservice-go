// @title           User Service API
// @version         1.0.0
// @description     User and Authentication microservice

// @host            localhost:9090
// @BasePath        /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/pkg/middleware"
	"ecommerce-microservice-go/pkg/psql"
	"ecommerce-microservice-go/pkg/security"
	"ecommerce-microservice-go/services/user/handler"
	"ecommerce-microservice-go/services/user/repository"
	"ecommerce-microservice-go/services/user/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "ecommerce-microservice-go/services/user/docs"
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

	log.Info("Starting User Service")

	// Connect to database
	db, err := psql.ConnectDB(log)
	if err != nil {
		log.Panic("Failed to connect to database", zap.Error(err))
	}

	// Auto-migrate
	if err := psql.AutoMigrate(db, log, &repository.User{}); err != nil {
		log.Panic("Failed to migrate database", zap.Error(err))
	}

	// Seed initial user
	if err := repository.SeedInitialUser(db, log); err != nil {
		log.Warn("Failed to seed initial user", zap.Error(err))
	}

	// Dependencies
	userRepo := repository.NewUserRepository(db, log)
	jwtService := security.NewJWTService()
	authUC := usecase.NewAuthUseCase(userRepo, jwtService, log)
	userUC := usecase.NewUserUseCase(userRepo, log)
	h := handler.NewHandler(authUC, userUC, log)

	// Router
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

	// Health
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user"})
	})

	v1.GET("/user/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/register", h.Register)
	auth.POST("/access-token", h.GetAccessTokenByRefreshToken)

	// User routes (protected)
	user := v1.Group("/user")
	user.Use(middleware.AuthJWTMiddleware())
	{
		user.GET("/", h.GetAllUsers)
		user.POST("/", h.NewUser)
		user.GET("/:id", h.GetUserByID)
		user.PUT("/:id", h.UpdateUser)
		user.DELETE("/:id", h.DeleteUser)
	}

	// Start server
	port := getEnvOrDefault("SERVER_PORT", "8081")
	log.Info("User Service starting", zap.String("port", port))
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
