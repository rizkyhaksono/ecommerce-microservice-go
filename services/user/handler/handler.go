package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"ecommerce-microservice-go/pkg/controllers"
	domainErrors "ecommerce-microservice-go/pkg/errors"
	"ecommerce-microservice-go/pkg/logger"
	userDomain "ecommerce-microservice-go/services/user/domain"
	"ecommerce-microservice-go/services/user/usecase"

	"github.com/gin-gonic/gin"
)

// Request/Response types

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AccessTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type NewUserRequest struct {
	UserName  string `json:"userName" binding:"required"`
	Email     string `json:"email" binding:"required"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password" binding:"required"`
	Status    bool   `json:"status"`
}

type UserData struct {
	UserName  string `json:"userName"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Status    bool   `json:"status"`
	ID        int    `json:"id"`
}

type SecurityData struct {
	JWTAccessToken            string    `json:"jwtAccessToken"`
	JWTRefreshToken           string    `json:"jwtRefreshToken"`
	ExpirationAccessDateTime  time.Time `json:"expirationAccessDateTime"`
	ExpirationRefreshDateTime time.Time `json:"expirationRefreshDateTime"`
}

type LoginResponse struct {
	Data     UserData     `json:"data"`
	Security SecurityData `json:"security"`
}

type ResponseUser struct {
	ID        int       `json:"id"`
	UserName  string    `json:"userName"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type Handler struct {
	authUseCase usecase.IAuthUseCase
	userUseCase usecase.IUserUseCase
	Logger      *logger.Logger
}

func NewHandler(auth usecase.IAuthUseCase, user usecase.IUserUseCase, l *logger.Logger) *Handler {
	return &Handler{authUseCase: auth, userUseCase: user, Logger: l}
}

// --- Auth handlers ---

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user account (Public)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body NewUserRequest true "User registration details"
// @Success      200 {object} ResponseUser
// @Failure      400 {object} controllers.MessageResponse
// @Router       /auth/register [post]
func (h *Handler) Register(ctx *gin.Context) {
	var request NewUserRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	u, err := h.userUseCase.Create(&userDomain.User{
		UserName: request.UserName, Email: request.Email,
		FirstName: request.FirstName, LastName: request.LastName,
		HashPassword: request.Password, Status: true, // Auto-active for registration
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseUser(u))
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password, returns JWT tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Router       /auth/login [post]
func (h *Handler) Login(ctx *gin.Context) {
	var request LoginRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	user, tokens, err := h.authUseCase.Login(request.Email, request.Password)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, LoginResponse{
		Data:     toUserData(user),
		Security: toSecurityData(tokens),
	})
}

// GetAccessTokenByRefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using a valid refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body AccessTokenRequest true "Refresh token"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} controllers.MessageResponse
// @Failure      401 {object} controllers.MessageResponse
// @Router       /auth/access-token [post]
func (h *Handler) GetAccessTokenByRefreshToken(ctx *gin.Context) {
	var request AccessTokenRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	user, tokens, err := h.authUseCase.AccessTokenByRefreshToken(request.RefreshToken)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, LoginResponse{
		Data:     toUserData(user),
		Security: toSecurityData(tokens),
	})
}

// --- User handlers ---

// NewUser godoc
// @Summary      Create a new user
// @Description  Create a new user account
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body NewUserRequest true "User details"
// @Success      200 {object} ResponseUser
// @Failure      400 {object} controllers.MessageResponse
// @Router       /user/ [post]
func (h *Handler) NewUser(ctx *gin.Context) {
	var request NewUserRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	u, err := h.userUseCase.Create(&userDomain.User{
		UserName: request.UserName, Email: request.Email,
		FirstName: request.FirstName, LastName: request.LastName,
		HashPassword: request.Password, Status: request.Status,
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseUser(u))
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieve a list of all users
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} ResponseUser
// @Failure      500 {object} controllers.MessageResponse
// @Router       /user/ [get]
func (h *Handler) GetAllUsers(ctx *gin.Context) {
	users, err := h.userUseCase.GetAll()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, arrayDomainToResponse(users))
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieve a single user by their ID
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} ResponseUser
// @Failure      400 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /user/{id} [get]
func (h *Handler) GetUserByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid user id"), domainErrors.ValidationError))
		return
	}
	u, err := h.userUseCase.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseUser(u))
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Update user fields by ID
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        request body map[string]interface{} true "Fields to update"
// @Success      200 {object} ResponseUser
// @Failure      400 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /user/{id} [put]
func (h *Handler) UpdateUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid user id"), domainErrors.ValidationError))
		return
	}
	var requestMap map[string]any
	if err := controllers.BindJSONMap(ctx, &requestMap); err != nil {
		_ = ctx.Error(domainErrors.NewAppError(err, domainErrors.ValidationError))
		return
	}
	updated, err := h.userUseCase.Update(id, requestMap)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, domainToResponseUser(updated))
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user by ID
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} controllers.MessageResponse
// @Failure      400 {object} controllers.MessageResponse
// @Failure      404 {object} controllers.MessageResponse
// @Router       /user/{id} [delete]
func (h *Handler) DeleteUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppError(errors.New("invalid user id"), domainErrors.ValidationError))
		return
	}
	if err := h.userUseCase.Delete(id); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

// Mappers
func domainToResponseUser(u *userDomain.User) ResponseUser {
	return ResponseUser{
		ID: u.ID, UserName: u.UserName, Email: u.Email,
		FirstName: u.FirstName, LastName: u.LastName, Status: u.Status,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func arrayDomainToResponse(users *[]userDomain.User) []ResponseUser {
	res := make([]ResponseUser, len(*users))
	for i, u := range *users {
		res[i] = domainToResponseUser(&u)
	}
	return res
}

func toUserData(u *userDomain.User) UserData {
	return UserData{UserName: u.UserName, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName, Status: u.Status, ID: u.ID}
}

func toSecurityData(t *usecase.AuthTokens) SecurityData {
	return SecurityData{
		JWTAccessToken: t.AccessToken, JWTRefreshToken: t.RefreshToken,
		ExpirationAccessDateTime: t.ExpirationAccessDateTime, ExpirationRefreshDateTime: t.ExpirationRefreshDateTime,
	}
}
