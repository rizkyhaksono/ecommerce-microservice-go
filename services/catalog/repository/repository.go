package repository

import (
	"encoding/json"
	"time"

	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/services/catalog/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// --- Category GORM model ---
type Category struct {
	ID          int       `gorm:"primaryKey"`
	Name        string    `gorm:"column:name;not null"`
	Description string    `gorm:"column:description"`
	Slug        string    `gorm:"column:slug;unique;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime:mili"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:mili"`
}

func (Category) TableName() string { return "categories" }

// --- Product GORM model ---
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

func (Product) TableName() string { return "products" }

// --- Category Repository ---

type CategoryRepositoryInterface interface {
	GetAll() (*[]domain.Category, error)
	GetByID(id int) (*domain.Category, error)
	Create(c *domain.Category) (*domain.Category, error)
	Update(id int, m map[string]interface{}) (*domain.Category, error)
	Delete(id int) error
}

type CategoryRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewCategoryRepository(db *gorm.DB, l *logger.Logger) CategoryRepositoryInterface {
	return &CategoryRepository{DB: db, Logger: l}
}

func (r *CategoryRepository) GetAll() (*[]domain.Category, error) {
	var cats []Category
	if err := r.DB.Find(&cats).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	result := make([]domain.Category, len(cats))
	for i, c := range cats {
		result[i] = domain.Category{ID: c.ID, Name: c.Name, Description: c.Description, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	}
	return &result, nil
}

func (r *CategoryRepository) GetByID(id int) (*domain.Category, error) {
	var c Category
	if err := r.DB.Where("id = ?", id).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return &domain.Category{ID: c.ID, Name: c.Name, Description: c.Description, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}, nil
}

func (r *CategoryRepository) Create(d *domain.Category) (*domain.Category, error) {
	c := Category{Name: d.Name, Description: d.Description, Slug: d.Slug}
	if err := r.DB.Create(&c).Error; err != nil {
		byteErr, _ := json.Marshal(err)
		var ge domainErrors.GormErr
		if json.Unmarshal(byteErr, &ge) == nil && ge.Number == 1062 {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return &domain.Category{ID: c.ID, Name: c.Name, Description: c.Description, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}, nil
}

func (r *CategoryRepository) Update(id int, m map[string]interface{}) (*domain.Category, error) {
	var c Category
	c.ID = id
	if err := r.DB.Model(&c).Updates(m).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Where("id = ?", id).First(&c).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return &domain.Category{ID: c.ID, Name: c.Name, Description: c.Description, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}, nil
}

func (r *CategoryRepository) Delete(id int) error {
	tx := r.DB.Delete(&Category{}, id)
	if tx.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

// --- Product Repository ---

type ProductRepositoryInterface interface {
	GetAll() (*[]domain.Product, error)
	GetByID(id int) (*domain.Product, error)
	GetByCategory(categoryID int) (*[]domain.Product, error)
	Create(p *domain.Product) (*domain.Product, error)
	Update(id int, m map[string]interface{}) (*domain.Product, error)
	Delete(id int) error
}

type ProductRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewProductRepository(db *gorm.DB, l *logger.Logger) ProductRepositoryInterface {
	return &ProductRepository{DB: db, Logger: l}
}

func (r *ProductRepository) GetAll() (*[]domain.Product, error) {
	var products []Product
	if err := r.DB.Where("is_active = ?", true).Find(&products).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return productsToDomainn(products), nil
}

func (r *ProductRepository) GetByID(id int) (*domain.Product, error) {
	var p Product
	if err := r.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return productToDomain(&p), nil
}

func (r *ProductRepository) GetByCategory(categoryID int) (*[]domain.Product, error) {
	var products []Product
	if err := r.DB.Where("category_id = ? AND is_active = ?", categoryID, true).Find(&products).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return productsToDomainn(products), nil
}

func (r *ProductRepository) Create(d *domain.Product) (*domain.Product, error) {
	p := Product{Name: d.Name, Description: d.Description, SKU: d.SKU, Price: d.Price, Stock: d.Stock, CategoryID: d.CategoryID, ImageURL: d.ImageURL, IsActive: d.IsActive}
	if err := r.DB.Create(&p).Error; err != nil {
		r.Logger.Error("Error creating product", zap.Error(err))
		byteErr, _ := json.Marshal(err)
		var ge domainErrors.GormErr
		if json.Unmarshal(byteErr, &ge) == nil && ge.Number == 1062 {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return productToDomain(&p), nil
}

func (r *ProductRepository) Update(id int, m map[string]interface{}) (*domain.Product, error) {
	var p Product
	p.ID = id
	if err := r.DB.Model(&p).Updates(m).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Where("id = ?", id).First(&p).Error; err != nil {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return productToDomain(&p), nil
}

func (r *ProductRepository) Delete(id int) error {
	tx := r.DB.Delete(&Product{}, id)
	if tx.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

func productToDomain(p *Product) *domain.Product {
	return &domain.Product{ID: p.ID, Name: p.Name, Description: p.Description, SKU: p.SKU, Price: p.Price, Stock: p.Stock, CategoryID: p.CategoryID, ImageURL: p.ImageURL, IsActive: p.IsActive, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
}

func productsToDomainn(products []Product) *[]domain.Product {
	result := make([]domain.Product, len(products))
	for i, p := range products {
		result[i] = *productToDomain(&p)
	}
	return &result
}
