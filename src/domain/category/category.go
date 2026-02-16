package category

import "time"

type Category struct {
	ID          int
	Name        string
	Description string
	Slug        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ICategoryService interface {
	GetAll() (*[]Category, error)
	GetByID(id int) (*Category, error)
	Create(category *Category) (*Category, error)
	Update(id int, categoryMap map[string]interface{}) (*Category, error)
	Delete(id int) error
}
