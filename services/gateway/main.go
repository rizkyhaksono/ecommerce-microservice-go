// @title           Ecommerce API Gateway
// @version         1.0.0
// @description     API Gateway that routes requests to microservices (User, Catalog, Order)

// @host            localhost:9090
// @BasePath        /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/pkg/observability"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type ServiceConfig struct {
	UserURL    string
	CatalogURL string
	OrderURL   string
}

func main() {
	env := getEnvOrDefault("GO_ENV", "development")
	serviceName := getEnvOrDefault("OTEL_SERVICE_NAME", "gateway")

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

	log, err := logger.NewLoggerWithOTel(otelProviders.LoggerProvider, "ecommerce-microservice-go/services/gateway", env == "development")
	if err != nil {
		panic(fmt.Errorf("error initializing logger: %w", err))
	}
	defer func() { _ = log.Log.Sync() }()

	log.Info("Starting API Gateway")

	cfg := ServiceConfig{
		UserURL:    getEnvOrDefault("USER_SERVICE_URL", "http://localhost:9091"),
		CatalogURL: getEnvOrDefault("CATALOG_SERVICE_URL", "http://localhost:9092"),
		OrderURL:   getEnvOrDefault("ORDER_SERVICE_URL", "http://localhost:9093"),
	}

	if env == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(otelgin.Middleware(serviceName)) // first: creates the request span/context
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(log.GinZapLogger())

	// Root Handler
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Ecommerce Microservices API Gateway",
			"status":  "running",
			"services": gin.H{
				"user":    "/v1/health",
				"catalog": "/v1/health",
				"order":   "/v1/health",
			},
			"docs": gin.H{
				"user":    "/v1/user/docs/index.html",
				"catalog": "/v1/catalog/docs/index.html",
				"order":   "/v1/order/docs/index.html",
			},
		})
	})

	v1 := router.Group("/v1")

	// Health check
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "gateway",
		})
	})

	// User Service routes
	userProxy := createReverseProxy(cfg.UserURL, log.Log)
	v1.Any("/auth/*path", proxyHandler(userProxy))
	v1.Any("/user/*path", proxyHandler(userProxy))

	// Catalog Service routes
	catalogProxy := createReverseProxy(cfg.CatalogURL, log.Log)
	v1.Any("/category/*path", proxyHandler(catalogProxy))
	v1.Any("/product/*path", proxyHandler(catalogProxy))
	v1.Any("/catalog/*path", proxyHandler(catalogProxy))

	// Order Service routes
	orderProxy := createReverseProxy(cfg.OrderURL, log.Log)
	v1.Any("/order/*path", proxyHandler(orderProxy))

	port := getEnvOrDefault("SERVER_PORT", "9090")
	log.Info("API Gateway starting", zap.String("port", port), zap.String("userService", cfg.UserURL), zap.String("catalogService", cfg.CatalogURL), zap.String("orderService", cfg.OrderURL))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Gateway failed to start", zap.Error(err))
	}
}

func createReverseProxy(target string, log *zap.Logger) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal("Invalid service URL", zap.String("target", target), zap.Error(err))
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// Inject W3C traceparent on the downstream call so the target service's
	// otelgin middleware continues the same trace (gateway -> service).
	proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("Proxy error", zap.String("target", target), zap.String("path", r.URL.Path), zap.Error(err))
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error": "service unavailable"}`))
	}
	return proxy
}

func proxyHandler(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rebuild the URL path: strip the /v1 prefix group and re-add the full path
		// Gin's *path captures everything after the route group
		// The reverse proxy target already has /v1 in its path
		c.Request.URL.Path = "/v1" + c.Request.URL.Path[len("/v1"):]
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
