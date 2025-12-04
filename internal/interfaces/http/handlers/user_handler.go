package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"user-service/internal/interfaces/http/dto"
	"user-service/internal/usecase"
	"user-service/pkg/errors"
)

type UserHandler struct {
	createUserUseCase *usecase.CreateUserUseCase
	loginUseCase      *usecase.LoginUseCase
	getUserUseCase    *usecase.GetUserUseCase
}

func NewUserHandler(
	createUserUseCase *usecase.CreateUserUseCase,
	loginUseCase *usecase.LoginUseCase,
	getUserUseCase *usecase.GetUserUseCase,
) *UserHandler {
	return &UserHandler{
		createUserUseCase: createUserUseCase,
		loginUseCase:     loginUseCase,
		getUserUseCase:   getUserUseCase,
	}
}

// @Summary Crear usuario
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "Datos del usuario"
// @Success 201 {object} usecase.CreateUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "datos inválidos",
			Message: "Verifique que todos los campos requeridos estén presentes y sean válidos: " + err.Error(),
		})
		return
	}

	useCaseReq := usecase.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	response, err := h.createUserUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Credenciales"
// @Success 200 {object} usecase.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "datos inválidos",
			Message: "Email y contraseña son requeridos: " + err.Error(),
		})
		return
	}

	useCaseReq := usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := h.loginUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Obtener usuario
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} usecase.GetUserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "no autorizado",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "error interno",
		})
		return
	}

	response, err := h.getUserUseCase.Execute(c.Request.Context(), userIDStr)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func handleError(c *gin.Context, err error) {
	if errWithCode, ok := err.(*errors.ErrorWithCode); ok {
		c.JSON(errWithCode.Code, dto.ErrorResponse{
			Error:   errWithCode.Message,
			Message: errWithCode.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error:   "error interno del servidor",
		Message: err.Error(),
	})
}

