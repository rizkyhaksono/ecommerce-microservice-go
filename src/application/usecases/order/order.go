package order

import (
	orderDomain "github.com/gbrayhan/microservices-go/src/domain/order"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/order"
	"go.uber.org/zap"
)

type IOrderUseCase interface {
	Create(order *orderDomain.Order) (*orderDomain.Order, error)
	GetByID(id int) (*orderDomain.Order, error)
	GetByUserID(userID int) (*[]orderDomain.Order, error)
	UpdateStatus(id int, status orderDomain.OrderStatus) (*orderDomain.Order, error)
}

type OrderUseCase struct {
	orderRepository order.OrderRepositoryInterface
	Logger          *logger.Logger
}

func NewOrderUseCase(repo order.OrderRepositoryInterface, logger *logger.Logger) IOrderUseCase {
	return &OrderUseCase{orderRepository: repo, Logger: logger}
}

func (s *OrderUseCase) Create(o *orderDomain.Order) (*orderDomain.Order, error) {
	s.Logger.Info("Creating new order", zap.Int("userID", o.UserID))

	// Calculate subtotals and total
	var total float64
	for i := range o.Items {
		o.Items[i].Subtotal = float64(o.Items[i].Quantity) * o.Items[i].Price
		total += o.Items[i].Subtotal
	}
	o.TotalAmount = total
	o.Status = orderDomain.OrderStatusPending

	return s.orderRepository.Create(o)
}

func (s *OrderUseCase) GetByID(id int) (*orderDomain.Order, error) {
	s.Logger.Info("Getting order by ID", zap.Int("id", id))
	return s.orderRepository.GetByID(id)
}

func (s *OrderUseCase) GetByUserID(userID int) (*[]orderDomain.Order, error) {
	s.Logger.Info("Getting orders by user ID", zap.Int("userID", userID))
	return s.orderRepository.GetByUserID(userID)
}

func (s *OrderUseCase) UpdateStatus(id int, status orderDomain.OrderStatus) (*orderDomain.Order, error) {
	s.Logger.Info("Updating order status", zap.Int("id", id), zap.String("status", string(status)))
	return s.orderRepository.UpdateStatus(id, status)
}
