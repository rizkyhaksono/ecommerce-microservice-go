package di

import (
	"sync"

	authUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/auth"
	categoryUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/category"
	orderUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/order"
	productUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/product"
	userUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/user"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/category"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/order"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/product"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/user"
	authController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/auth"
	categoryController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/category"
	orderController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/order"
	productController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/product"
	userController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/user"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
	"gorm.io/gorm"
)

// ApplicationContext holds all application dependencies and services
type ApplicationContext struct {
	DB                 *gorm.DB
	Logger             *logger.Logger
	AuthController     authController.IAuthController
	UserController     userController.IUserController
	CategoryController categoryController.ICategoryController
	ProductController  productController.IProductController
	OrderController    orderController.IOrderController
	JWTService         security.IJWTService
	UserRepository     user.UserRepositoryInterface
	CategoryRepository category.CategoryRepositoryInterface
	ProductRepository  product.ProductRepositoryInterface
	OrderRepository    order.OrderRepositoryInterface
	AuthUseCase        authUseCase.IAuthUseCase
	UserUseCase        userUseCase.IUserUseCase
	CategoryUseCase    categoryUseCase.ICategoryUseCase
	ProductUseCase     productUseCase.IProductUseCase
	OrderUseCase       orderUseCase.IOrderUseCase
}

var (
	loggerInstance *logger.Logger
	loggerOnce     sync.Once
)

func GetLogger() *logger.Logger {
	loggerOnce.Do(func() {
		loggerInstance, _ = logger.NewLogger()
	})
	return loggerInstance
}

// SetupDependencies creates a new application context with all dependencies
func SetupDependencies(loggerInstance *logger.Logger) (*ApplicationContext, error) {
	// Initialize database with logger
	db, err := psql.InitPSQLDB(loggerInstance)
	if err != nil {
		return nil, err
	}

	// Initialize JWT service
	jwtService := security.NewJWTService()

	// Initialize repositories
	userRepo := user.NewUserRepository(db, loggerInstance)
	categoryRepo := category.NewCategoryRepository(db, loggerInstance)
	productRepo := product.NewProductRepository(db, loggerInstance)
	orderRepo := order.NewOrderRepository(db, loggerInstance)

	// Initialize use cases
	authUC := authUseCase.NewAuthUseCase(userRepo, jwtService, loggerInstance)
	userUC := userUseCase.NewUserUseCase(userRepo, loggerInstance)
	categoryUC := categoryUseCase.NewCategoryUseCase(categoryRepo, loggerInstance)
	productUC := productUseCase.NewProductUseCase(productRepo, loggerInstance)
	orderUC := orderUseCase.NewOrderUseCase(orderRepo, loggerInstance)

	// Initialize controllers
	authCtrl := authController.NewAuthController(authUC, loggerInstance)
	userCtrl := userController.NewUserController(userUC, loggerInstance)
	categoryCtrl := categoryController.NewCategoryController(categoryUC, loggerInstance)
	productCtrl := productController.NewProductController(productUC, loggerInstance)
	orderCtrl := orderController.NewOrderController(orderUC, loggerInstance)

	return &ApplicationContext{
		DB:                 db,
		Logger:             loggerInstance,
		AuthController:     authCtrl,
		UserController:     userCtrl,
		CategoryController: categoryCtrl,
		ProductController:  productCtrl,
		OrderController:    orderCtrl,
		JWTService:         jwtService,
		UserRepository:     userRepo,
		CategoryRepository: categoryRepo,
		ProductRepository:  productRepo,
		OrderRepository:    orderRepo,
		AuthUseCase:        authUC,
		UserUseCase:        userUC,
		CategoryUseCase:    categoryUC,
		ProductUseCase:     productUC,
		OrderUseCase:       orderUC,
	}, nil
}

// NewTestApplicationContext creates an application context for testing with mocked dependencies
func NewTestApplicationContext(
	mockUserRepo user.UserRepositoryInterface,
	mockJWTService security.IJWTService,
	loggerInstance *logger.Logger,
) *ApplicationContext {
	authUC := authUseCase.NewAuthUseCase(mockUserRepo, mockJWTService, loggerInstance)
	userUC := userUseCase.NewUserUseCase(mockUserRepo, loggerInstance)

	authCtrl := authController.NewAuthController(authUC, loggerInstance)
	userCtrl := userController.NewUserController(userUC, loggerInstance)

	return &ApplicationContext{
		Logger:         loggerInstance,
		AuthController: authCtrl,
		UserController: userCtrl,
		JWTService:     mockJWTService,
		UserRepository: mockUserRepo,
		AuthUseCase:    authUC,
		UserUseCase:    userUC,
	}
}
