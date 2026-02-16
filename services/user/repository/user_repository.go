package repository

import (
	"encoding/json"
	"os"
	"time"

	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	userDomain "ecommerce-microservice-go/services/user/domain"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           int       `gorm:"primaryKey"`
	UserName     string    `gorm:"column:user_name"`
	Email        string    `gorm:"column:email;unique"`
	FirstName    string    `gorm:"column:first_name"`
	LastName     string    `gorm:"column:last_name"`
	Status       bool      `gorm:"column:status"`
	HashPassword string    `gorm:"column:hash_password"`
	CreatedAt    time.Time `gorm:"autoCreateTime:mili"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime:mili"`
}

func (User) TableName() string {
	return "users"
}

type UserRepositoryInterface interface {
	GetAll() (*[]userDomain.User, error)
	GetByID(id int) (*userDomain.User, error)
	GetByEmail(email string) (*userDomain.User, error)
	Create(user *userDomain.User) (*userDomain.User, error)
	Update(id int, userMap map[string]interface{}) (*userDomain.User, error)
	Delete(id int) error
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewUserRepository(db *gorm.DB, loggerInstance *logger.Logger) UserRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]userDomain.User, error) {
	var users []User
	if err := r.DB.Find(&users).Error; err != nil {
		r.Logger.Error("Error getting all users", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return arrayToDomainMapper(&users), nil
}

func (r *Repository) GetByID(id int) (*userDomain.User, error) {
	var u User
	err := r.DB.Where("id = ?", id).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return u.toDomainMapper(), nil
}

func (r *Repository) GetByEmail(email string) (*userDomain.User, error) {
	var u User
	err := r.DB.Where("email = ?", email).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return u.toDomainMapper(), nil
}

func (r *Repository) Create(uDomain *userDomain.User) (*userDomain.User, error) {
	u := fromDomainMapper(uDomain)
	txResult := r.DB.Create(u)
	if txResult.Error != nil {
		byteErr, _ := json.Marshal(txResult.Error)
		var newError domainErrors.GormErr
		if errUnmarshal := json.Unmarshal(byteErr, &newError); errUnmarshal != nil {
			return &userDomain.User{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	return u.toDomainMapper(), nil
}

func (r *Repository) Update(id int, userMap map[string]interface{}) (*userDomain.User, error) {
	var u User
	u.ID = id
	if err := r.DB.Model(&u).Updates(userMap).Error; err != nil {
		return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if err := r.DB.Where("id = ?", id).First(&u).Error; err != nil {
		return &userDomain.User{}, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return u.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&User{}, id)
	if tx.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

// SeedInitialUser seeds the initial admin user from env vars
func SeedInitialUser(db *gorm.DB, loggerInstance *logger.Logger) error {
	email := os.Getenv("START_USER_EMAIL")
	pw := os.Getenv("START_USER_PW")
	if email == "" || pw == "" {
		loggerInstance.Info("Initial user seed skipped: START_USER_EMAIL or START_USER_PW not set")
		return nil
	}
	var existing User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		loggerInstance.Info("Initial user already exists, skipping seed", zap.String("email", email))
		return nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Create(&User{Email: email, HashPassword: string(hashedPassword)}).Error
}

// Mappers
func (u *User) toDomainMapper() *userDomain.User {
	return &userDomain.User{
		ID: u.ID, UserName: u.UserName, Email: u.Email,
		FirstName: u.FirstName, LastName: u.LastName, Status: u.Status,
		HashPassword: u.HashPassword, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func fromDomainMapper(u *userDomain.User) *User {
	return &User{
		ID: u.ID, UserName: u.UserName, Email: u.Email,
		FirstName: u.FirstName, LastName: u.LastName, Status: u.Status,
		HashPassword: u.HashPassword,
	}
}

func arrayToDomainMapper(users *[]User) *[]userDomain.User {
	result := make([]userDomain.User, len(*users))
	for i, u := range *users {
		result[i] = *u.toDomainMapper()
	}
	return &result
}
