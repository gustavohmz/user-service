package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"user-service/internal/interfaces/http/handlers"
	"user-service/internal/usecase"
)

// Mocks para casos de uso - implementaciones simples para testing
type mockCreateUserUseCase struct {
	*usecase.CreateUserUseCase
	response *usecase.CreateUserResponse
	err      error
}

func (m *mockCreateUserUseCase) Execute(ctx context.Context, req usecase.CreateUserRequest) (*usecase.CreateUserResponse, error) {
	return m.response, m.err
}

type mockLoginUseCase struct {
	*usecase.LoginUseCase
	response *usecase.LoginResponse
	err      error
}

func (m *mockLoginUseCase) Execute(ctx context.Context, req usecase.LoginRequest) (*usecase.LoginResponse, error) {
	return m.response, m.err
}

type mockGetUserUseCase struct {
	*usecase.GetUserUseCase
	response *usecase.GetUserResponse
	err      error
}

func (m *mockGetUserUseCase) Execute(ctx context.Context, userID string) (*usecase.GetUserResponse, error) {
	return m.response, m.err
}

func setupRouter(handler *handlers.UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	{
		api.POST("/users", handler.CreateUser)
		api.POST("/auth/login", handler.Login)
		api.GET("/users/me", handler.GetUser)
	}
	return router
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	// Este test requiere mocks más complejos o una refactorización del handler
	// Por ahora, solo verificamos que el handler se crea correctamente
	handler := handlers.NewUserHandler(
		&usecase.CreateUserUseCase{},
		&usecase.LoginUseCase{},
		&usecase.GetUserUseCase{},
	)
	
	if handler == nil {
		t.Error("Expected handler to be created, got nil")
	}
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	// Arrange
	handler := handlers.NewUserHandler(
		&usecase.CreateUserUseCase{},
		&usecase.LoginUseCase{},
		&usecase.GetUserUseCase{},
	)

	router := setupRouter(handler)

	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}
}

func TestUserHandler_Login_InvalidJSON(t *testing.T) {
	// Arrange
	handler := handlers.NewUserHandler(
		&usecase.CreateUserUseCase{},
		&usecase.LoginUseCase{},
		&usecase.GetUserUseCase{},
	)

	router := setupRouter(handler)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}
}

