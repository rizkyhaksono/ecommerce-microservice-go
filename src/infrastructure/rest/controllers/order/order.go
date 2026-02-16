package order

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	orderDomain "github.com/gbrayhan/microservices-go/src/domain/order"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderItemRequest struct {
	ProductID int     `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
}

type NewOrderRequest struct {
	ShippingAddress string             `json:"shippingAddress" binding:"required"`
	Items           []OrderItemRequest `json:"items" binding:"required,dive"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type ResponseOrderItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}

type ResponseOrder struct {
	ID              int                 `json:"id"`
	UserID          int                 `json:"userId"`
	Status          string              `json:"status"`
	TotalAmount     float64             `json:"totalAmount"`
	ShippingAddress string              `json:"shippingAddress"`
	Items           []ResponseOrderItem `json:"items"`
	CreatedAt       time.Time           `json:"createdAt,omitempty"`
	UpdatedAt       time.Time           `json:"updatedAt,omitempty"`
}

type IOrderController interface {
	CreateOrder(ctx *gin.Context)
	GetOrderByID(ctx *gin.Context)
	GetMyOrders(ctx *gin.Context)
	UpdateOrderStatus(ctx *gin.Context)
}

type Controller struct {
	orderService orderDomain.IOrderService
	Logger       *logger.Logger
}

func NewOrderController(service orderDomain.IOrderService, loggerInstance *logger.Logger) IOrderController {
	return &Controller{orderService: service, Logger: loggerInstance}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order with items. UserID is extracted from the JWT token.
// @Tags         Order
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body NewOrderRequest true "Order details with items"
// @Success      200 {object} ResponseOrder
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Router       /order/ [post]
func (c *Controller) CreateOrder(ctx *gin.Context) {
	var request NewOrderRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new order", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	// Get user ID from JWT context
	userIDVal, exists := ctx.Get("userId")
	if !exists {
		appError := domainErrors.NewAppError(errors.New("user not authenticated"), domainErrors.NotAuthenticated)
		_ = ctx.Error(appError)
		return
	}
	userID := int(userIDVal.(float64))

	// Build domain order
	items := make([]orderDomain.OrderItem, len(request.Items))
	for i, item := range request.Items {
		items[i] = orderDomain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	order, err := c.orderService.Create(&orderDomain.Order{
		UserID:          userID,
		ShippingAddress: request.ShippingAddress,
		Items:           items,
	})
	if err != nil {
		c.Logger.Error("Error creating order", zap.Error(err))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Order created", zap.Int("id", order.ID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(order))
}

// GetOrderByID godoc
// @Summary      Get order by ID
// @Description  Retrieve a single order with its items by ID
// @Tags         Order
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Success      200 {object} ResponseOrder
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /order/{id} [get]
func (c *Controller) GetOrderByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid order id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	order, err := c.orderService.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseMapper(order))
}

// GetMyOrders godoc
// @Summary      Get current user's orders
// @Description  Retrieve all orders for the authenticated user
// @Tags         Order
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} ResponseOrder
// @Failure      401 {object} controllers.MessageResponse
// @Router       /order/ [get]
func (c *Controller) GetMyOrders(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("userId")
	if !exists {
		appError := domainErrors.NewAppError(errors.New("user not authenticated"), domainErrors.NotAuthenticated)
		_ = ctx.Error(appError)
		return
	}
	userID := int(userIDVal.(float64))

	orders, err := c.orderService.GetByUserID(userID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, arrayDomainToResponseMapper(orders))
}

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Description  Update the status of an order (pending, paid, shipped, delivered, cancelled)
// @Tags         Order
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Param        request body UpdateStatusRequest true "New status"
// @Success      200 {object} ResponseOrder
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /order/{id}/status [put]
func (c *Controller) UpdateOrderStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		appError := domainErrors.NewAppError(errors.New("invalid order id"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	var request UpdateStatusRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending": true, "paid": true, "shipped": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[request.Status] {
		appError := domainErrors.NewAppError(errors.New("invalid status: must be pending, paid, shipped, delivered, or cancelled"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	order, err := c.orderService.UpdateStatus(id, orderDomain.OrderStatus(request.Status))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseMapper(order))
}

// Mappers
func domainToResponseMapper(o *orderDomain.Order) *ResponseOrder {
	items := make([]ResponseOrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = ResponseOrderItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  item.Subtotal,
		}
	}
	return &ResponseOrder{
		ID:              o.ID,
		UserID:          o.UserID,
		Status:          string(o.Status),
		TotalAmount:     o.TotalAmount,
		ShippingAddress: o.ShippingAddress,
		Items:           items,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

func arrayDomainToResponseMapper(orders *[]orderDomain.Order) *[]ResponseOrder {
	res := make([]ResponseOrder, len(*orders))
	for i, o := range *orders {
		res[i] = *domainToResponseMapper(&o)
	}
	return &res
}
