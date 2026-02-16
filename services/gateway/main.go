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
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ServiceConfig struct {
	UserURL    string
	CatalogURL string
	OrderURL   string
}

func main() {
	log := initLogger()
	defer func() { _ = log.Sync() }()

	log.Info("Starting API Gateway")

	cfg := ServiceConfig{
		UserURL:    getEnvOrDefault("USER_SERVICE_URL", "http://localhost:9091"),
		CatalogURL: getEnvOrDefault("CATALOG_SERVICE_URL", "http://localhost:9092"),
		OrderURL:   getEnvOrDefault("ORDER_SERVICE_URL", "http://localhost:9093"),
	}

	env := getEnvOrDefault("GO_ENV", "development")
	if env == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(zapLoggerMiddleware(log))

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
	userProxy := createReverseProxy(cfg.UserURL, log)
	v1.Any("/auth/*path", proxyHandler(userProxy))
	v1.Any("/user/*path", proxyHandler(userProxy))

	// Catalog Service routes
	catalogProxy := createReverseProxy(cfg.CatalogURL, log)
	v1.Any("/category/*path", proxyHandler(catalogProxy))
	v1.Any("/product/*path", proxyHandler(catalogProxy))
	v1.Any("/catalog/*path", proxyHandler(catalogProxy))

	// Order Service routes
	orderProxy := createReverseProxy(cfg.OrderURL, log)
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

func initLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	return zap.New(core)
}

func zapLoggerMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
