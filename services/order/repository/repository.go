package repository

import (
	"time"

	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/services/order/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GORM models
type Order struct {
	ID          int         `gorm:"primaryKey"`
	UserID      int         `gorm:"column:user_id;not null"`
	Status      string      `gorm:"column:status;default:pending"`
	TotalAmount float64     `gorm:"column:total_amount;default:0"`
	Items       []OrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt   time.Time   `gorm:"autoCreateTime:mili"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime:mili"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID        int     `gorm:"primaryKey"`
	OrderID   int     `gorm:"column:order_id;not null"`
	ProductID int     `gorm:"column:product_id;not null"`
	Quantity  int     `gorm:"column:quantity;not null"`
	Price     float64 `gorm:"column:price;not null"`
	Subtotal  float64 `gorm:"column:subtotal;not null"`
}

func (OrderItem) TableName() string { return "order_items" }

// Interfaces

type OrderRepositoryInterface interface {
	GetAll() (*[]domain.Order, error)
	GetByID(id int) (*domain.Order, error)
	GetByUserID(userID int) (*[]domain.Order, error)
	Create(order *domain.Order) (*domain.Order, error)
	UpdateStatus(id int, status string) (*domain.Order, error)
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewOrderRepository(db *gorm.DB, l *logger.Logger) OrderRepositoryInterface {
	return &Repository{DB: db, Logger: l}
}

func (r *Repository) GetAll() (*[]domain.Order, error) {
	var orders []Order
	if err := r.DB.Preload("Items").Find(&orders).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return ordersToDomain(orders), nil
}

func (r *Repository) GetByID(id int) (*domain.Order, error) {
	var o Order
	if err := r.DB.Preload("Items").Where("id = ?", id).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return orderToDomain(&o), nil
}

func (r *Repository) GetByUserID(userID int) (*[]domain.Order, error) {
	var orders []Order
	if err := r.DB.Preload("Items").Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return ordersToDomain(orders), nil
}

func (r *Repository) Create(d *domain.Order) (*domain.Order, error) {
	o := fromDomain(d)
	if err := r.DB.Create(o).Error; err != nil {
		r.Logger.Error("Error creating order", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	// Reload with items
	var created Order
	r.DB.Preload("Items").Where("id = ?", o.ID).First(&created)
	return orderToDomain(&created), nil
}

func (r *Repository) UpdateStatus(id int, status string) (*domain.Order, error) {
	var o Order
	if err := r.DB.Where("id = ?", id).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Model(&o).Update("status", status).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.DB.Preload("Items").Where("id = ?", id).First(&o)
	return orderToDomain(&o), nil
}

// Mappers
func orderToDomain(o *Order) *domain.Order {
	items := make([]domain.OrderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = domain.OrderItem{ID: it.ID, OrderID: it.OrderID, ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Subtotal: it.Subtotal}
	}
	return &domain.Order{ID: o.ID, UserID: o.UserID, Status: domain.OrderStatus(o.Status), TotalAmount: o.TotalAmount, Items: items, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt}
}

func ordersToDomain(orders []Order) *[]domain.Order {
	result := make([]domain.Order, len(orders))
	for i, o := range orders {
		result[i] = *orderToDomain(&o)
	}
	return &result
}

func fromDomain(d *domain.Order) *Order {
	items := make([]OrderItem, len(d.Items))
	for i, it := range d.Items {
		items[i] = OrderItem{ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Subtotal: it.Subtotal}
	}
	return &Order{UserID: d.UserID, Status: string(d.Status), TotalAmount: d.TotalAmount, Items: items}
}
