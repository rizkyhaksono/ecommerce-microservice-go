package category

import (
	categoryDomain "github.com/gbrayhan/microservices-go/src/domain/category"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/category"
	"go.uber.org/zap"
)

type ICategoryUseCase interface {
	GetAll() (*[]categoryDomain.Category, error)
	GetByID(id int) (*categoryDomain.Category, error)
	Create(category *categoryDomain.Category) (*categoryDomain.Category, error)
	Update(id int, categoryMap map[string]interface{}) (*categoryDomain.Category, error)
	Delete(id int) error
}

type CategoryUseCase struct {
	categoryRepository category.CategoryRepositoryInterface
	Logger             *logger.Logger
}

func NewCategoryUseCase(repo category.CategoryRepositoryInterface, logger *logger.Logger) ICategoryUseCase {
	return &CategoryUseCase{categoryRepository: repo, Logger: logger}
}

func (s *CategoryUseCase) GetAll() (*[]categoryDomain.Category, error) {
	s.Logger.Info("Getting all categories")
	return s.categoryRepository.GetAll()
}

func (s *CategoryUseCase) GetByID(id int) (*categoryDomain.Category, error) {
	s.Logger.Info("Getting category by ID", zap.Int("id", id))
	return s.categoryRepository.GetByID(id)
}

func (s *CategoryUseCase) Create(cat *categoryDomain.Category) (*categoryDomain.Category, error) {
	s.Logger.Info("Creating new category", zap.String("name", cat.Name))
	return s.categoryRepository.Create(cat)
}

func (s *CategoryUseCase) Update(id int, categoryMap map[string]interface{}) (*categoryDomain.Category, error) {
	s.Logger.Info("Updating category", zap.Int("id", id))
	return s.categoryRepository.Update(id, categoryMap)
}

func (s *CategoryUseCase) Delete(id int) error {
	s.Logger.Info("Deleting category", zap.Int("id", id))
	return s.categoryRepository.Delete(id)
}
