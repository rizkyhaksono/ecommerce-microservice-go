package user

import "time"

type User struct {
	ID           int
	UserName     string
	Email        string
	FirstName    string
	LastName     string
	Status       bool
	HashPassword string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type IUserService interface {
	GetAll() (*[]User, error)
	GetByID(id int) (*User, error)
	Create(user *User) (*User, error)
	Update(id int, userMap map[string]interface{}) (*User, error)
	Delete(id int) error
}
