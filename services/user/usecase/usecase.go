package usecase

import (
	"errors"
	"time"

	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	"ecommerce-microservice-go/pkg/security"
	userDomain "ecommerce-microservice-go/services/user/domain"
	"ecommerce-microservice-go/services/user/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// --- User UseCase ---

type IUserUseCase interface {
	GetAll() (*[]userDomain.User, error)
	GetByID(id int) (*userDomain.User, error)
	Create(user *userDomain.User) (*userDomain.User, error)
	Update(id int, userMap map[string]interface{}) (*userDomain.User, error)
	Delete(id int) error
}

type UserUseCase struct {
	userRepository repository.UserRepositoryInterface
	Logger         *logger.Logger
}

func NewUserUseCase(repo repository.UserRepositoryInterface, l *logger.Logger) IUserUseCase {
	return &UserUseCase{userRepository: repo, Logger: l}
}

func (s *UserUseCase) GetAll() (*[]userDomain.User, error) {
	s.Logger.Info("Getting all users")
	return s.userRepository.GetAll()
}

func (s *UserUseCase) GetByID(id int) (*userDomain.User, error) {
	s.Logger.Info("Getting user by ID", zap.Int("id", id))
	return s.userRepository.GetByID(id)
}

func (s *UserUseCase) Create(u *userDomain.User) (*userDomain.User, error) {
	s.Logger.Info("Creating new user", zap.String("email", u.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(u.HashPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.HashPassword = string(hash)
	return s.userRepository.Create(u)
}

func (s *UserUseCase) Update(id int, userMap map[string]interface{}) (*userDomain.User, error) {
	s.Logger.Info("Updating user", zap.Int("id", id))
	return s.userRepository.Update(id, userMap)
}

func (s *UserUseCase) Delete(id int) error {
	s.Logger.Info("Deleting user", zap.Int("id", id))
	return s.userRepository.Delete(id)
}

// --- Auth UseCase ---

type IAuthUseCase interface {
	Login(email, password string) (*userDomain.User, *AuthTokens, error)
	AccessTokenByRefreshToken(refreshToken string) (*userDomain.User, *AuthTokens, error)
}

type AuthUseCase struct {
	UserRepository repository.UserRepositoryInterface
	JWTService     security.IJWTService
	Logger         *logger.Logger
}

func NewAuthUseCase(repo repository.UserRepositoryInterface, jwt security.IJWTService, l *logger.Logger) IAuthUseCase {
	return &AuthUseCase{UserRepository: repo, JWTService: jwt, Logger: l}
}

type AuthTokens struct {
	AccessToken               string
	RefreshToken              string
	ExpirationAccessDateTime  time.Time
	ExpirationRefreshDateTime time.Time
}

func (s *AuthUseCase) Login(email, password string) (*userDomain.User, *AuthTokens, error) {
	s.Logger.Info("User login attempt", zap.String("email", email))
	user, err := s.UserRepository.GetByEmail(email)
	if err != nil {
		return nil, nil, err
	}
	if user.ID == 0 {
		return nil, nil, domainErrors.NewAppError(errors.New("email or password does not match"), domainErrors.NotAuthenticated)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password)) != nil {
		return nil, nil, domainErrors.NewAppError(errors.New("email or password does not match"), domainErrors.NotAuthenticated)
	}

	accessToken, err := s.JWTService.GenerateJWTToken(user.ID, "access")
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := s.JWTService.GenerateJWTToken(user.ID, "refresh")
	if err != nil {
		return nil, nil, err
	}

	return user, &AuthTokens{
		AccessToken:               accessToken.Token,
		RefreshToken:              refreshToken.Token,
		ExpirationAccessDateTime:  accessToken.ExpirationTime,
		ExpirationRefreshDateTime: refreshToken.ExpirationTime,
	}, nil
}

func (s *AuthUseCase) AccessTokenByRefreshToken(refreshToken string) (*userDomain.User, *AuthTokens, error) {
	s.Logger.Info("Refreshing access token")
	claimsMap, err := s.JWTService.GetClaimsAndVerifyToken(refreshToken, "refresh")
	if err != nil {
		return nil, nil, err
	}
	userID := int(claimsMap["id"].(float64))
	user, err := s.UserRepository.GetByID(userID)
	if err != nil {
		return nil, nil, err
	}

	accessToken, err := s.JWTService.GenerateJWTToken(user.ID, "access")
	if err != nil {
		return nil, nil, err
	}

	expTime := int64(claimsMap["exp"].(float64))

	return user, &AuthTokens{
		AccessToken:               accessToken.Token,
		RefreshToken:              refreshToken,
		ExpirationAccessDateTime:  accessToken.ExpirationTime,
		ExpirationRefreshDateTime: time.Unix(expTime, 0),
	}, nil
}
