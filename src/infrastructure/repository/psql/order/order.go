package order

import (
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	orderDomain "github.com/gbrayhan/microservices-go/src/domain/order"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Order struct {
	ID              int         `gorm:"primaryKey"`
	UserID          int         `gorm:"column:user_id;not null"`
	Status          string      `gorm:"column:status;default:pending"`
	TotalAmount     float64     `gorm:"column:total_amount;not null"`
	ShippingAddress string      `gorm:"column:shipping_address"`
	Items           []OrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt       time.Time   `gorm:"autoCreateTime:mili"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime:mili"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID        int     `gorm:"primaryKey"`
	OrderID   int     `gorm:"column:order_id;not null"`
	ProductID int     `gorm:"column:product_id;not null"`
	Quantity  int     `gorm:"column:quantity;not null"`
	Price     float64 `gorm:"column:price;not null"`
	Subtotal  float64 `gorm:"column:subtotal;not null"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type OrderRepositoryInterface interface {
	Create(order *orderDomain.Order) (*orderDomain.Order, error)
	GetByID(id int) (*orderDomain.Order, error)
	GetByUserID(userID int) (*[]orderDomain.Order, error)
	UpdateStatus(id int, status orderDomain.OrderStatus) (*orderDomain.Order, error)
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewOrderRepository(db *gorm.DB, loggerInstance *logger.Logger) OrderRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) Create(oDomain *orderDomain.Order) (*orderDomain.Order, error) {
	o := fromDomainMapper(oDomain)
	if err := r.DB.Create(o).Error; err != nil {
		r.Logger.Error("Error creating order", zap.Error(err))
		return &orderDomain.Order{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	// Reload with items
	if err := r.DB.Preload("Items").First(o, o.ID).Error; err != nil {
		return &orderDomain.Order{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return o.toDomainMapper(), nil
}

func (r *Repository) GetByID(id int) (*orderDomain.Order, error) {
	var o Order
	err := r.DB.Preload("Items").Where("id = ?", id).First(&o).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &orderDomain.Order{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		r.Logger.Error("Error getting order by ID", zap.Error(err), zap.Int("id", id))
		return &orderDomain.Order{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return o.toDomainMapper(), nil
}

func (r *Repository) GetByUserID(userID int) (*[]orderDomain.Order, error) {
	var orders []Order
	if err := r.DB.Preload("Items").Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		r.Logger.Error("Error getting orders by user", zap.Error(err), zap.Int("userID", userID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return arrayToDomainMapper(&orders), nil
}

func (r *Repository) UpdateStatus(id int, status orderDomain.OrderStatus) (*orderDomain.Order, error) {
	var o Order
	if err := r.DB.Model(&Order{}).Where("id = ?", id).Update("status", string(status)).Error; err != nil {
		r.Logger.Error("Error updating order status", zap.Error(err), zap.Int("id", id))
		return &orderDomain.Order{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Preload("Items").Where("id = ?", id).First(&o).Error; err != nil {
		return &orderDomain.Order{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return o.toDomainMapper(), nil
}

// Mappers
func (o *Order) toDomainMapper() *orderDomain.Order {
	items := make([]orderDomain.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = orderDomain.OrderItem{
			ID:        item.ID,
			OrderID:   item.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  item.Subtotal,
		}
	}
	return &orderDomain.Order{
		ID:              o.ID,
		UserID:          o.UserID,
		Status:          orderDomain.OrderStatus(o.Status),
		TotalAmount:     o.TotalAmount,
		ShippingAddress: o.ShippingAddress,
		Items:           items,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

func fromDomainMapper(o *orderDomain.Order) *Order {
	items := make([]OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  item.Subtotal,
		}
	}
	return &Order{
		UserID:          o.UserID,
		Status:          string(o.Status),
		TotalAmount:     o.TotalAmount,
		ShippingAddress: o.ShippingAddress,
		Items:           items,
	}
}

func arrayToDomainMapper(orders *[]Order) *[]orderDomain.Order {
	result := make([]orderDomain.Order, len(*orders))
	for i, o := range *orders {
		result[i] = *o.toDomainMapper()
	}
	return &result
}
