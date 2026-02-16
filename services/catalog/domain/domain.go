package domain

import "time"

type Category struct {
	ID          int
	Name        string
	Description string
	Slug        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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
