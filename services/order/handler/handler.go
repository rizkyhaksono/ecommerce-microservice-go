package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"ecommerce-microservice-go/pkg/controllers"
	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/services/order/domain"
	"ecommerce-microservice-go/services/order/usecase"

	"github.com/gin-gonic/gin"
)

type OrderItemRequest struct {
	ProductID int     `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
}

type NewOrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required"`
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
	ID          int                 `json:"id"`
	UserID      int                 `json:"userId"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"totalAmount"`
	Items       []ResponseOrderItem `json:"items"`
	CreatedAt   time.Time           `json:"createdAt,omitempty"`
	UpdatedAt   time.Time           `json:"updatedAt,omitempty"`
}

type Handler struct {
	orderUC usecase.IOrderUseCase
	Logger  *logger.Logger
}

func NewHandler(uc usecase.IOrderUseCase, l *logger.Logger) *Handler {
	return &Handler{orderUC: uc, Logger: l}
}

// GetAllOrders godoc
// @Summary      Get all orders
// @Tags         Order
// @Security     BearerAuth
// @Success      200 {array} ResponseOrder
// @Router       /order/ [get]
func (h *Handler) GetAllOrders(ctx *gin.Context) {
	orders, err := h.orderUC.GetAll()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ordersToResponse(orders))
}

// GetOrderByID godoc
// @Summary      Get order by ID
// @Tags         Order
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Success      200 {object} ResponseOrder
// @Router       /order/{id} [get]
func (h *Handler) GetOrderByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	o, err := h.orderUC.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, orderToResponse(o))
}

// NewOrder godoc
// @Summary      Create order
// @Tags         Order
// @Security     BearerAuth
// @Param        request body NewOrderRequest true "Order"
// @Success      200 {object} ResponseOrder
// @Router       /order/ [post]
func (h *Handler) NewOrder(ctx *gin.Context) {
	var req NewOrderRequest
	if err := controllers.BindJSON(ctx, &req); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}

	// Extract user ID from JWT context
	userIDVal, exists := ctx.Get("userId")
	if !exists {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("user id not found in token"), domainErrors.NotAuthenticated))
		return
	}
	userID := int(userIDVal.(float64))

	items := make([]domain.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price}
	}

	o, err := h.orderUC.Create(&domain.Order{UserID: userID, Items: items})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, orderToResponse(o))
}

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Tags         Order
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Param        request body UpdateStatusRequest true "Status"
// @Success      200 {object} ResponseOrder
// @Router       /order/{id}/status [put]
func (h *Handler) UpdateOrderStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid id"), domainErrors.ValidationError))
		return
	}
	var req UpdateStatusRequest
	if err := controllers.BindJSON(ctx, &req); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	o, err := h.orderUC.UpdateStatus(id, req.Status)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, orderToResponse(o))
}

// Mappers
func orderToResponse(o *domain.Order) ResponseOrder {
	items := make([]ResponseOrderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = ResponseOrderItem{ID: it.ID, ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Subtotal: it.Subtotal}
	}
	return ResponseOrder{ID: o.ID, UserID: o.UserID, Status: string(o.Status), TotalAmount: o.TotalAmount, Items: items, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt}
}

func ordersToResponse(orders *[]domain.Order) []ResponseOrder {
	res := make([]ResponseOrder, len(*orders))
	for i, o := range *orders {
		res[i] = orderToResponse(&o)
	}
	return res
}
