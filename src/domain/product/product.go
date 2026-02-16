package product

import "time"

type Product struct {
	ID          int
	Name        string
	Description string
	SKU         string
	Price       float64
	Stock       int
	CategoryID  int
	ImageURL    string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type IProductService interface {
	GetAll() (*[]Product, error)
	GetByID(id int) (*Product, error)
	GetByCategory(categoryID int) (*[]Product, error)
	Create(product *Product) (*Product, error)
	Update(id int, productMap map[string]interface{}) (*Product, error)
	Delete(id int) error
}
