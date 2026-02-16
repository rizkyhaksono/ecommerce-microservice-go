package product

import (
	"encoding/json"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	productDomain "github.com/gbrayhan/microservices-go/src/domain/product"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Product struct {
	ID          int       `gorm:"primaryKey"`
	Name        string    `gorm:"column:name;not null"`
	Description string    `gorm:"column:description"`
	SKU         string    `gorm:"column:sku;unique;not null"`
	Price       float64   `gorm:"column:price;not null"`
	Stock       int       `gorm:"column:stock;default:0"`
	CategoryID  int       `gorm:"column:category_id;not null"`
	ImageURL    string    `gorm:"column:image_url"`
	IsActive    bool      `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime:mili"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:mili"`
}

func (Product) TableName() string {
	return "products"
}

type ProductRepositoryInterface interface {
	GetAll() (*[]productDomain.Product, error)
	GetByID(id int) (*productDomain.Product, error)
	GetByCategory(categoryID int) (*[]productDomain.Product, error)
	Create(product *productDomain.Product) (*productDomain.Product, error)
	Update(id int, productMap map[string]interface{}) (*productDomain.Product, error)
	Delete(id int) error
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewProductRepository(db *gorm.DB, loggerInstance *logger.Logger) ProductRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]productDomain.Product, error) {
	var products []Product
	if err := r.DB.Where("is_active = ?", true).Find(&products).Error; err != nil {
		r.Logger.Error("Error getting all products", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return arrayToDomainMapper(&products), nil
}

func (r *Repository) GetByID(id int) (*productDomain.Product, error) {
	var p Product
	err := r.DB.Where("id = ?", id).First(&p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &productDomain.Product{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		r.Logger.Error("Error getting product by ID", zap.Error(err), zap.Int("id", id))
		return &productDomain.Product{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return p.toDomainMapper(), nil
}

func (r *Repository) GetByCategory(categoryID int) (*[]productDomain.Product, error) {
	var products []Product
	if err := r.DB.Where("category_id = ? AND is_active = ?", categoryID, true).Find(&products).Error; err != nil {
		r.Logger.Error("Error getting products by category", zap.Error(err), zap.Int("categoryID", categoryID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return arrayToDomainMapper(&products), nil
}

func (r *Repository) Create(pDomain *productDomain.Product) (*productDomain.Product, error) {
	p := fromDomainMapper(pDomain)
	if err := r.DB.Create(p).Error; err != nil {
		r.Logger.Error("Error creating product", zap.Error(err))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		if errUnmarshal := json.Unmarshal(byteErr, &newError); errUnmarshal != nil {
			return &productDomain.Product{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &productDomain.Product{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &productDomain.Product{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	return p.toDomainMapper(), nil
}

func (r *Repository) Update(id int, productMap map[string]interface{}) (*productDomain.Product, error) {
	var p Product
	p.ID = id

	if err := r.DB.Model(&p).Updates(productMap).Error; err != nil {
		r.Logger.Error("Error updating product", zap.Error(err), zap.Int("id", id))
		return &productDomain.Product{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Where("id = ?", id).First(&p).Error; err != nil {
		return &productDomain.Product{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return p.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&Product{}, id)
	if tx.Error != nil {
		r.Logger.Error("Error deleting product", zap.Error(tx.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

// Mappers
func (p *Product) toDomainMapper() *productDomain.Product {
	return &productDomain.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SKU:         p.SKU,
		Price:       p.Price,
		Stock:       p.Stock,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func fromDomainMapper(p *productDomain.Product) *Product {
	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SKU:         p.SKU,
		Price:       p.Price,
		Stock:       p.Stock,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		IsActive:    p.IsActive,
	}
}

func arrayToDomainMapper(products *[]Product) *[]productDomain.Product {
	result := make([]productDomain.Product, len(*products))
	for i, p := range *products {
		result[i] = *p.toDomainMapper()
	}
	return &result
}
