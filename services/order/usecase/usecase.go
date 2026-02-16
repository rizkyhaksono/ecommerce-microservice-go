package usecase

import (
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/services/order/domain"
	"ecommerce-microservice-go/services/order/repository"

	"go.uber.org/zap"
)

type IOrderUseCase interface {
	GetAll() (*[]domain.Order, error)
	GetByID(id int) (*domain.Order, error)
	GetByUserID(userID int) (*[]domain.Order, error)
	Create(order *domain.Order) (*domain.Order, error)
	UpdateStatus(id int, status string) (*domain.Order, error)
}

type OrderUseCase struct {
	repo   repository.OrderRepositoryInterface
	Logger *logger.Logger
}

func NewOrderUseCase(r repository.OrderRepositoryInterface, l *logger.Logger) IOrderUseCase {
	return &OrderUseCase{repo: r, Logger: l}
}

func (s *OrderUseCase) GetAll() (*[]domain.Order, error) {
	s.Logger.Info("Getting all orders")
	return s.repo.GetAll()
}

func (s *OrderUseCase) GetByID(id int) (*domain.Order, error) {
	s.Logger.Info("Getting order by ID", zap.Int("id", id))
	return s.repo.GetByID(id)
}

func (s *OrderUseCase) GetByUserID(userID int) (*[]domain.Order, error) {
	s.Logger.Info("Getting orders by user ID", zap.Int("userID", userID))
	return s.repo.GetByUserID(userID)
}

func (s *OrderUseCase) Create(order *domain.Order) (*domain.Order, error) {
	s.Logger.Info("Creating order", zap.Int("userID", order.UserID))
	// Calculate subtotals and total
	var total float64
	for i := range order.Items {
		order.Items[i].Subtotal = float64(order.Items[i].Quantity) * order.Items[i].Price
		total += order.Items[i].Subtotal
	}
	order.TotalAmount = total
	order.Status = domain.OrderStatusPending
	return s.repo.Create(order)
}

func (s *OrderUseCase) UpdateStatus(id int, status string) (*domain.Order, error) {
	s.Logger.Info("Updating order status", zap.Int("id", id), zap.String("status", status))
	return s.repo.UpdateStatus(id, status)
}
