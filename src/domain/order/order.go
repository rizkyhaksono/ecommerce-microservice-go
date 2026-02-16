package order

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              int
	UserID          int
	Status          OrderStatus
	TotalAmount     float64
	ShippingAddress string
	Items           []OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
	Price     float64
	Subtotal  float64
}

type IOrderService interface {
	Create(order *Order) (*Order, error)
	GetByID(id int) (*Order, error)
	GetByUserID(userID int) (*[]Order, error)
	UpdateStatus(id int, status OrderStatus) (*Order, error)
}
