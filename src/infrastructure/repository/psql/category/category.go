package category

import (
	"encoding/json"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	categoryDomain "github.com/gbrayhan/microservices-go/src/domain/category"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Category struct {
	ID          int       `gorm:"primaryKey"`
	Name        string    `gorm:"column:name;not null"`
	Description string    `gorm:"column:description"`
	Slug        string    `gorm:"column:slug;unique;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime:mili"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:mili"`
}

func (Category) TableName() string {
	return "categories"
}

type CategoryRepositoryInterface interface {
	GetAll() (*[]categoryDomain.Category, error)
	GetByID(id int) (*categoryDomain.Category, error)
	Create(category *categoryDomain.Category) (*categoryDomain.Category, error)
	Update(id int, categoryMap map[string]interface{}) (*categoryDomain.Category, error)
	Delete(id int) error
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewCategoryRepository(db *gorm.DB, loggerInstance *logger.Logger) CategoryRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]categoryDomain.Category, error) {
	var categories []Category
	if err := r.DB.Find(&categories).Error; err != nil {
		r.Logger.Error("Error getting all categories", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return arrayToDomainMapper(&categories), nil
}

func (r *Repository) GetByID(id int) (*categoryDomain.Category, error) {
	var cat Category
	err := r.DB.Where("id = ?", id).First(&cat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &categoryDomain.Category{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		r.Logger.Error("Error getting category by ID", zap.Error(err), zap.Int("id", id))
		return &categoryDomain.Category{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return cat.toDomainMapper(), nil
}

func (r *Repository) Create(catDomain *categoryDomain.Category) (*categoryDomain.Category, error) {
	cat := fromDomainMapper(catDomain)
	if err := r.DB.Create(cat).Error; err != nil {
		r.Logger.Error("Error creating category", zap.Error(err))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		if errUnmarshal := json.Unmarshal(byteErr, &newError); errUnmarshal != nil {
			return &categoryDomain.Category{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &categoryDomain.Category{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &categoryDomain.Category{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	return cat.toDomainMapper(), nil
}

func (r *Repository) Update(id int, categoryMap map[string]interface{}) (*categoryDomain.Category, error) {
	var cat Category
	cat.ID = id

	if err := r.DB.Model(&cat).Updates(categoryMap).Error; err != nil {
		r.Logger.Error("Error updating category", zap.Error(err), zap.Int("id", id))
		return &categoryDomain.Category{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Where("id = ?", id).First(&cat).Error; err != nil {
		return &categoryDomain.Category{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return cat.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&Category{}, id)
	if tx.Error != nil {
		r.Logger.Error("Error deleting category", zap.Error(tx.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

// Mappers
func (c *Category) toDomainMapper() *categoryDomain.Category {
	return &categoryDomain.Category{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Slug:        c.Slug,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func fromDomainMapper(c *categoryDomain.Category) *Category {
	return &Category{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Slug:        c.Slug,
	}
}

func arrayToDomainMapper(categories *[]Category) *[]categoryDomain.Category {
	result := make([]categoryDomain.Category, len(*categories))
	for i, c := range *categories {
		result[i] = *c.toDomainMapper()
	}
	return &result
}
