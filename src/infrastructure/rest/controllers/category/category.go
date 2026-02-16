package category

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	categoryDomain "github.com/gbrayhan/microservices-go/src/domain/category"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

type ICategoryController interface {
	NewCategory(ctx *gin.Context)
	GetAllCategories(ctx *gin.Context)
	GetCategoryByID(ctx *gin.Context)
	UpdateCategory(ctx *gin.Context)
	DeleteCategory(ctx *gin.Context)
}

type Controller struct {
	categoryService categoryDomain.ICategoryService
	Logger          *logger.Logger
}

func NewCategoryController(service categoryDomain.ICategoryService, loggerInstance *logger.Logger) ICategoryController {
	return &Controller{categoryService: service, Logger: loggerInstance}
}

// NewCategory godoc
// @Summary      Create a new category
// @Description  Create a new product category
// @Tags         Category
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body NewCategoryRequest true "Category details"
// @Success      200 {object} ResponseCategory
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      409 {object} controllers.MessageResponse
// @Router       /category/ [post]
func (c *Controller) NewCategory(ctx *gin.Context) {
	var request NewCategoryRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new category", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	cat, err := c.categoryService.Create(&categoryDomain.Category{
		Name:        request.Name,
		Description: request.Description,
		Slug:        request.Slug,
	})
	if err != nil {
		c.Logger.Error("Error creating category", zap.Error(err))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Category created", zap.Int("id", cat.ID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(cat))
}

// GetAllCategories godoc
// @Summary      Get all categories
// @Description  Retrieve a list of all product categories
// @Tags         Category
// @Produce      json
// @Success      200 {array} ResponseCategory
// @Failure      500 {object} controllers.MessageResponse
// @Router       /category/ [get]
func (c *Controller) GetAllCategories(ctx *gin.Context) {
	categories, err := c.categoryService.GetAll()
	if err != nil {
		c.Logger.Error("Error getting all categories", zap.Error(err))
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, arrayDomainToResponseMapper(categories))
}

// GetCategoryByID godoc
// @Summary      Get category by ID
// @Description  Retrieve a single category by its ID
// @Tags         Category
// @Produce      json
// @Param        id path int true "Category ID"
// @Success      200 {object} ResponseCategory
// @Failure      400 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /category/{id} [get]
func (c *Controller) GetCategoryByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid category id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	cat, err := c.categoryService.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseMapper(cat))
}

// UpdateCategory godoc
// @Summary      Update a category
// @Description  Update category fields by ID
// @Tags         Category
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Param        request body map[string]interface{} true "Fields to update"
// @Success      200 {object} ResponseCategory
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /category/{id} [put]
func (c *Controller) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid category id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	var requestMap map[string]any
	if err := controllers.BindJSONMap(ctx, &requestMap); err != nil {
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	updated, err := c.categoryService.Update(id, requestMap)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseMapper(updated))
}

// DeleteCategory godoc
// @Summary      Delete a category
// @Description  Delete a category by ID
// @Tags         Category
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} controllers.MessageResponse
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /category/{id} [delete]
func (c *Controller) DeleteCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid category id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	if err := c.categoryService.Delete(id); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

// Mappers
func domainToResponseMapper(c *categoryDomain.Category) *ResponseCategory {
	return &ResponseCategory{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Slug:        c.Slug,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func arrayDomainToResponseMapper(categories *[]categoryDomain.Category) *[]ResponseCategory {
	res := make([]ResponseCategory, len(*categories))
	for i, c := range *categories {
		res[i] = *domainToResponseMapper(&c)
	}
	return &res
}
