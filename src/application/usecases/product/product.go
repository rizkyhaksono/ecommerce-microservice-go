package product

import (
	productDomain "github.com/gbrayhan/microservices-go/src/domain/product"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/product"
	"go.uber.org/zap"
)

type IProductUseCase interface {
	GetAll() (*[]productDomain.Product, error)
	GetByID(id int) (*productDomain.Product, error)
	GetByCategory(categoryID int) (*[]productDomain.Product, error)
	Create(product *productDomain.Product) (*productDomain.Product, error)
	Update(id int, productMap map[string]interface{}) (*productDomain.Product, error)
	Delete(id int) error
}

type ProductUseCase struct {
	productRepository product.ProductRepositoryInterface
	Logger            *logger.Logger
}

func NewProductUseCase(repo product.ProductRepositoryInterface, logger *logger.Logger) IProductUseCase {
	return &ProductUseCase{productRepository: repo, Logger: logger}
}

func (s *ProductUseCase) GetAll() (*[]productDomain.Product, error) {
	s.Logger.Info("Getting all products")
	return s.productRepository.GetAll()
}

func (s *ProductUseCase) GetByID(id int) (*productDomain.Product, error) {
	s.Logger.Info("Getting product by ID", zap.Int("id", id))
	return s.productRepository.GetByID(id)
}

func (s *ProductUseCase) GetByCategory(categoryID int) (*[]productDomain.Product, error) {
	s.Logger.Info("Getting products by category", zap.Int("categoryID", categoryID))
	return s.productRepository.GetByCategory(categoryID)
}

func (s *ProductUseCase) Create(p *productDomain.Product) (*productDomain.Product, error) {
	s.Logger.Info("Creating new product", zap.String("name", p.Name))
	return s.productRepository.Create(p)
}

func (s *ProductUseCase) Update(id int, productMap map[string]interface{}) (*productDomain.Product, error) {
	s.Logger.Info("Updating product", zap.Int("id", id))
	return s.productRepository.Update(id, productMap)
}

func (s *ProductUseCase) Delete(id int) error {
	s.Logger.Info("Deleting product", zap.Int("id", id))
	return s.productRepository.Delete(id)
}
