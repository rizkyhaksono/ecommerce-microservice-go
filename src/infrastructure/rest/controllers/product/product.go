package product

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	productDomain "github.com/gbrayhan/microservices-go/src/domain/product"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NewProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	SKU         string  `json:"sku" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock"`
	CategoryID  int     `json:"categoryId" binding:"required"`
	ImageURL    string  `json:"imageUrl"`
	IsActive    bool    `json:"isActive"`
}

type ResponseProduct struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SKU         string    `json:"sku"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  int       `json:"categoryId"`
	ImageURL    string    `json:"imageUrl"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

type IProductController interface {
	NewProduct(ctx *gin.Context)
	GetAllProducts(ctx *gin.Context)
	GetProductByID(ctx *gin.Context)
	GetProductsByCategory(ctx *gin.Context)
	UpdateProduct(ctx *gin.Context)
	DeleteProduct(ctx *gin.Context)
}

type Controller struct {
	productService productDomain.IProductService
	Logger         *logger.Logger
}

func NewProductController(service productDomain.IProductService, loggerInstance *logger.Logger) IProductController {
	return &Controller{productService: service, Logger: loggerInstance}
}

// NewProduct godoc
// @Summary      Create a new product
// @Description  Create a new product listing
// @Tags         Product
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body NewProductRequest true "Product details"
// @Success      200 {object} ResponseProduct
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      409 {object} controllers.MessageResponse
// @Router       /product/ [post]
func (c *Controller) NewProduct(ctx *gin.Context) {
	var request NewProductRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new product", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	p, err := c.productService.Create(&productDomain.Product{
		Name:        request.Name,
		Description: request.Description,
		SKU:         request.SKU,
		Price:       request.Price,
		Stock:       request.Stock,
		CategoryID:  request.CategoryID,
		ImageURL:    request.ImageURL,
		IsActive:    request.IsActive,
	})
	if err != nil {
		c.Logger.Error("Error creating product", zap.Error(err))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Product created", zap.Int("id", p.ID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(p))
}

// GetAllProducts godoc
// @Summary      Get all products
// @Description  Retrieve a list of all active products
// @Tags         Product
// @Produce      json
// @Success      200 {array} ResponseProduct
// @Failure      500 {object} controllers.MessageResponse
// @Router       /product/ [get]
func (c *Controller) GetAllProducts(ctx *gin.Context) {
	products, err := c.productService.GetAll()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, arrayDomainToResponseMapper(products))
}

// GetProductByID godoc
// @Summary      Get product by ID
// @Description  Retrieve a single product by its ID
// @Tags         Product
// @Produce      json
// @Param        id path int true "Product ID"
// @Success      200 {object} ResponseProduct
// @Failure      400 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /product/{id} [get]
func (c *Controller) GetProductByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid product id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	p, err := c.productService.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseMapper(p))
}

// GetProductsByCategory godoc
// @Summary      Get products by category
// @Description  Retrieve all active products belonging to a specific category
// @Tags         Product
// @Produce      json
// @Param        categoryId path int true "Category ID"
// @Success      200 {array} ResponseProduct
// @Failure      400 {object} controllers.MessageResponse
// @Router       /product/category/{categoryId} [get]
func (c *Controller) GetProductsByCategory(ctx *gin.Context) {
	categoryID, err := strconv.Atoi(ctx.Param("categoryId"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid category id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	products, err := c.productService.GetByCategory(categoryID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, arrayDomainToResponseMapper(products))
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Update product fields by ID
// @Tags         Product
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Param        request body map[string]interface{} true "Fields to update"
// @Success      200 {object} ResponseProduct
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /product/{id} [put]
func (c *Controller) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid product id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	var requestMap map[string]any
	if err := controllers.BindJSONMap(ctx, &requestMap); err != nil {
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	updated, err := c.productService.Update(id, requestMap)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseMapper(updated))
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Delete a product by ID
// @Tags         Product
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Success      200 {object} controllers.MessageResponse
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /product/{id} [delete]
func (c *Controller) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid product id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	if err := c.productService.Delete(id); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

// Mappers
func domainToResponseMapper(p *productDomain.Product) *ResponseProduct {
	return &ResponseProduct{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SKU:         p.SKU,
		Price:       p.Price,
		Stock:       p.Stock,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func arrayDomainToResponseMapper(products *[]productDomain.Product) *[]ResponseProduct {
	res := make([]ResponseProduct, len(*products))
	for i, p := range *products {
		res[i] = *domainToResponseMapper(&p)
	}
	return &res
}
