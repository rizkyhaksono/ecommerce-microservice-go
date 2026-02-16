package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"ecommerce-microservice-go/pkg/controllers"
	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/services/catalog/domain"
	"ecommerce-microservice-go/services/catalog/usecase"

	"github.com/gin-gonic/gin"
)

type NewCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Slug        string `json:"slug" binding:"required"`
}

type ResponseCategory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

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

type Handler struct {
	catUC  usecase.ICategoryUseCase
	prodUC usecase.IProductUseCase
	Logger *logger.Logger
}

func NewHandler(c usecase.ICategoryUseCase, p usecase.IProductUseCase, l *logger.Logger) *Handler {
	return &Handler{catUC: c, prodUC: p, Logger: l}
}

// --- Category handlers ---

// GetAllCategories godoc
// @Summary      Get all categories
// @Tags         Category
// @Produce      json
// @Success      200 {array} ResponseCategory
// @Router       /category/ [get]
func (h *Handler) GetAllCategories(ctx *gin.Context) {
	cats, err := h.catUC.GetAll()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	res := make([]ResponseCategory, len(*cats))
	for i, c := range *cats {
		res[i] = catToResponse(&c)
	}
	ctx.JSON(http.StatusOK, res)
}

// GetCategoryByID godoc
// @Summary      Get category by ID
// @Tags         Category
// @Param        id path int true "Category ID"
// @Success      200 {object} ResponseCategory
// @Router       /category/{id} [get]
func (h *Handler) GetCategoryByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	c, err := h.catUC.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, catToResponse(c))
}

// NewCategory godoc
// @Summary      Create category
// @Tags         Category
// @Security     BearerAuth
// @Param        request body NewCategoryRequest true "Category"
// @Success      200 {object} ResponseCategory
// @Router       /category/ [post]
func (h *Handler) NewCategory(ctx *gin.Context) {
	var req NewCategoryRequest
	if err := controllers.BindJSON(ctx, &req); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	c, err := h.catUC.Create(&domain.Category{Name: req.Name, Description: req.Description, Slug: req.Slug})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, catToResponse(c))
}

// UpdateCategory godoc
// @Summary      Update category
// @Tags         Category
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Param        request body map[string]interface{} true "Fields"
// @Success      200 {object} ResponseCategory
// @Router       /category/{id} [put]
func (h *Handler) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	var m map[string]any
	if err := controllers.BindJSONMap(ctx, &m); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	c, err := h.catUC.Update(id, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, catToResponse(c))
}

// DeleteCategory godoc
// @Summary      Delete category
// @Tags         Category
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} controllers.MessageResponse
// @Router       /category/{id} [delete]
func (h *Handler) DeleteCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	if err := h.catUC.Delete(id); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

// --- Product handlers ---

// GetAllProducts godoc
// @Summary      Get all products
// @Tags         Product
// @Success      200 {array} ResponseProduct
// @Router       /product/ [get]
func (h *Handler) GetAllProducts(ctx *gin.Context) {
	products, err := h.prodUC.GetAll()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, productsToResponse(products))
}

// GetProductByID godoc
// @Summary      Get product by ID
// @Tags         Product
// @Param        id path int true "Product ID"
// @Success      200 {object} ResponseProduct
// @Router       /product/{id} [get]
func (h *Handler) GetProductByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	p, err := h.prodUC.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, prodToResponse(p))
}

// GetProductsByCategory godoc
// @Summary      Get products by category
// @Tags         Product
// @Param        categoryId path int true "Category ID"
// @Success      200 {array} ResponseProduct
// @Router       /product/category/{categoryId} [get]
func (h *Handler) GetProductsByCategory(ctx *gin.Context) {
	catID, err := strconv.Atoi(ctx.Param("categoryId"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid category id"), domainErrors.ValidationError))
		return
	}
	products, err := h.prodUC.GetByCategory(catID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, productsToResponse(products))
}

// NewProduct godoc
// @Summary      Create product
// @Tags         Product
// @Security     BearerAuth
// @Param        request body NewProductRequest true "Product"
// @Success      200 {object} ResponseProduct
// @Router       /product/ [post]
func (h *Handler) NewProduct(ctx *gin.Context) {
	var req NewProductRequest
	if err := controllers.BindJSON(ctx, &req); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	p, err := h.prodUC.Create(&domain.Product{
		Name: req.Name, Description: req.Description, SKU: req.SKU,
		Price: req.Price, Stock: req.Stock, CategoryID: req.CategoryID,
		ImageURL: req.ImageURL, IsActive: req.IsActive,
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, prodToResponse(p))
}

// UpdateProduct godoc
// @Summary      Update product
// @Tags         Product
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Param        request body map[string]interface{} true "Fields"
// @Success      200 {object} ResponseProduct
// @Router       /product/{id} [put]
func (h *Handler) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	var m map[string]any
	if err := controllers.BindJSONMap(ctx, &m); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	p, err := h.prodUC.Update(id, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, prodToResponse(p))
}

// DeleteProduct godoc
// @Summary      Delete product
// @Tags         Product
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Success      200 {object} controllers.MessageResponse
// @Router       /product/{id} [delete]
func (h *Handler) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	if err := h.prodUC.Delete(id); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

// Mappers
func catToResponse(c *domain.Category) ResponseCategory {
	return ResponseCategory{ID: c.ID, Name: c.Name, Description: c.Description, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

func prodToResponse(p *domain.Product) ResponseProduct {
	return ResponseProduct{ID: p.ID, Name: p.Name, Description: p.Description, SKU: p.SKU, Price: p.Price, Stock: p.Stock, CategoryID: p.CategoryID, ImageURL: p.ImageURL, IsActive: p.IsActive, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
}

func productsToResponse(ps *[]domain.Product) []ResponseProduct {
	res := make([]ResponseProduct, len(*ps))
	for i, p := range *ps {
		res[i] = prodToResponse(&p)
	}
	return res
}
