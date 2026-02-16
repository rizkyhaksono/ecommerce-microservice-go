package security

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	domainErrors "ecommerce-microservice-go/pkg/errors"

	"github.com/golang-jwt/jwt/v4"
)

const (
	Access  = "access"
	Refresh = "refresh"
)

type AppToken struct {
	Token          string    `json:"token"`
	TokenType      string    `json:"type"`
	ExpirationTime time.Time `json:"expirationTime"`
}

type Claims struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	jwt.RegisteredClaims
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTime    int64
	RefreshTime   int64
}

type IJWTService interface {
	GenerateJWTToken(userID int, tokenType string) (*AppToken, error)
	GetClaimsAndVerifyToken(tokenString string, tokenType string) (jwt.MapClaims, error)
}

type JWTService struct {
	config JWTConfig
}

func NewJWTService() IJWTService {
	return &JWTService{config: loadJWTConfig()}
}

func NewJWTServiceWithConfig(config JWTConfig) IJWTService {
	return &JWTService{config: config}
}

func loadJWTConfig() JWTConfig {
	return JWTConfig{
		AccessSecret:  getEnvOrDefault("JWT_ACCESS_SECRET_KEY", "default_access_secret"),
		RefreshSecret: getEnvOrDefault("JWT_REFRESH_SECRET_KEY", "default_refresh_secret"),
		AccessTime:    getEnvAsInt64OrDefault("JWT_ACCESS_TIME_MINUTE", 60),
		RefreshTime:   getEnvAsInt64OrDefault("JWT_REFRESH_TIME_HOUR", 24),
	}
}

func (s *JWTService) GenerateJWTToken(userID int, tokenType string) (*AppToken, error) {
	var secretKey string
	var duration time.Duration

	switch tokenType {
	case Access:
		secretKey = s.config.AccessSecret
		duration = time.Duration(s.config.AccessTime) * time.Minute
	case Refresh:
		secretKey = s.config.RefreshSecret
		duration = time.Duration(s.config.RefreshTime) * time.Hour
	default:
		return nil, errors.New("invalid token type")
	}

	now := time.Now()
	exp := now.Add(duration)

	tokenClaims := &Claims{
		ID:   userID,
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims).SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &AppToken{Token: tokenStr, TokenType: tokenType, ExpirationTime: exp}, nil
}

func (s *JWTService) GetClaimsAndVerifyToken(tokenString string, tokenType string) (jwt.MapClaims, error) {
	var secretKey string
	if tokenType == Refresh {
		secretKey = s.config.RefreshSecret
	} else {
		secretKey = s.config.AccessSecret
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, domainErrors.NewAppError(
				fmt.Errorf("unexpected signing method: %v", token.Header["alg"]),
				domainErrors.NotAuthenticated,
			)
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.NotAuthenticated)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, domainErrors.NewAppError(errors.New("invalid claims type or token not valid"), domainErrors.NotAuthenticated)
	}

	if claims["type"] != tokenType {
		return nil, domainErrors.NewAppError(errors.New("invalid token type"), domainErrors.NotAuthenticated)
	}

	expVal, ok := claims["exp"]
	if !ok || expVal == nil {
		return nil, domainErrors.NewAppError(errors.New("token missing expiration (exp) claim"), domainErrors.NotAuthenticated)
	}
	timeExpire, ok := expVal.(float64)
	if !ok {
		return nil, domainErrors.NewAppError(errors.New("token exp claim is not a float64"), domainErrors.NotAuthenticated)
	}
	if time.Now().Unix() > int64(timeExpire) {
		return nil, domainErrors.NewAppError(errors.New("token expired"), domainErrors.NotAuthenticated)
	}

	idVal, ok := claims["id"]
	if !ok || idVal == nil {
		return nil, domainErrors.NewAppError(errors.New("token missing id claim"), domainErrors.NotAuthenticated)
	}
	switch idVal.(type) {
	case float64, int64, int:
	default:
		return nil, domainErrors.NewAppError(errors.New("token id claim is not a number"), domainErrors.NotAuthenticated)
	}

	return claims, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
