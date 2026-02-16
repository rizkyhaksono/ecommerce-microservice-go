package usecase

import (
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/services/catalog/domain"
	"ecommerce-microservice-go/services/catalog/repository"

	"go.uber.org/zap"
)

// --- Category UseCase ---

type ICategoryUseCase interface {
	GetAll() (*[]domain.Category, error)
	GetByID(id int) (*domain.Category, error)
	Create(c *domain.Category) (*domain.Category, error)
	Update(id int, m map[string]interface{}) (*domain.Category, error)
	Delete(id int) error
}

type CategoryUseCase struct {
	repo   repository.CategoryRepositoryInterface
	Logger *logger.Logger
}

func NewCategoryUseCase(r repository.CategoryRepositoryInterface, l *logger.Logger) ICategoryUseCase {
	return &CategoryUseCase{repo: r, Logger: l}
}

func (s *CategoryUseCase) GetAll() (*[]domain.Category, error) {
	s.Logger.Info("Getting all categories")
	return s.repo.GetAll()
}
func (s *CategoryUseCase) GetByID(id int) (*domain.Category, error) {
	s.Logger.Info("Getting category by ID", zap.Int("id", id))
	return s.repo.GetByID(id)
}
func (s *CategoryUseCase) Create(c *domain.Category) (*domain.Category, error) {
	s.Logger.Info("Creating category", zap.String("name", c.Name))
	return s.repo.Create(c)
}
func (s *CategoryUseCase) Update(id int, m map[string]interface{}) (*domain.Category, error) {
	s.Logger.Info("Updating category", zap.Int("id", id))
	return s.repo.Update(id, m)
}
func (s *CategoryUseCase) Delete(id int) error {
	s.Logger.Info("Deleting category", zap.Int("id", id))
	return s.repo.Delete(id)
}

// --- Product UseCase ---

type IProductUseCase interface {
	GetAll() (*[]domain.Product, error)
	GetByID(id int) (*domain.Product, error)
	GetByCategory(categoryID int) (*[]domain.Product, error)
	Create(p *domain.Product) (*domain.Product, error)
	Update(id int, m map[string]interface{}) (*domain.Product, error)
	Delete(id int) error
}

type ProductUseCase struct {
	repo   repository.ProductRepositoryInterface
	Logger *logger.Logger
}

func NewProductUseCase(r repository.ProductRepositoryInterface, l *logger.Logger) IProductUseCase {
	return &ProductUseCase{repo: r, Logger: l}
}

func (s *ProductUseCase) GetAll() (*[]domain.Product, error) {
	s.Logger.Info("Getting all products")
	return s.repo.GetAll()
}
func (s *ProductUseCase) GetByID(id int) (*domain.Product, error) {
	s.Logger.Info("Getting product by ID", zap.Int("id", id))
	return s.repo.GetByID(id)
}
func (s *ProductUseCase) GetByCategory(categoryID int) (*[]domain.Product, error) {
	s.Logger.Info("Getting products by category", zap.Int("categoryID", categoryID))
	return s.repo.GetByCategory(categoryID)
}
func (s *ProductUseCase) Create(p *domain.Product) (*domain.Product, error) {
	s.Logger.Info("Creating product", zap.String("name", p.Name))
	return s.repo.Create(p)
}
func (s *ProductUseCase) Update(id int, m map[string]interface{}) (*domain.Product, error) {
	s.Logger.Info("Updating product", zap.Int("id", id))
	return s.repo.Update(id, m)
}
func (s *ProductUseCase) Delete(id int) error {
	s.Logger.Info("Deleting product", zap.Int("id", id))
	return s.repo.Delete(id)
}
