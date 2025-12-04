package usecase_test

import (
	"context"
	"testing"

	"user-service/internal/domain"
	"user-service/internal/usecase"
	"user-service/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginUseCase_Execute_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	
	user := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Gustavo Hernández",
	}

	userRepo := &mockUserRepository{
		users: map[string]*domain.User{
			"test@example.com": user,
		},
	}
	jwtService := &mockJWTService{}

	useCase := usecase.NewLoginUseCase(userRepo, jwtService)

	req := usecase.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Act
	response, err := useCase.Execute(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Token == "" {
		t.Error("Expected token, got empty string")
	}
}

func TestLoginUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	jwtService := &mockJWTService{}

	useCase := usecase.NewLoginUseCase(userRepo, jwtService)

	req := usecase.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Act
	response, err := useCase.Execute(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent user, got nil")
	}

	if response != nil {
		t.Error("Expected nil response for nonexistent user")
	}

	// Verificar que el error es de credenciales inválidas
	if errWithCode, ok := err.(*errors.ErrorWithCode); ok {
		if errWithCode.Code != 401 {
			t.Errorf("Expected status code 401, got %d", errWithCode.Code)
		}
	}
}

func TestLoginUseCase_Execute_InvalidPassword(t *testing.T) {
	// Arrange
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	
	user := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Gustavo Hernández",
	}

	userRepo := &mockUserRepository{
		users: map[string]*domain.User{
			"test@example.com": user,
		},
	}
	jwtService := &mockJWTService{}

	useCase := usecase.NewLoginUseCase(userRepo, jwtService)

	req := usecase.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	// Act
	response, err := useCase.Execute(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}

	if response != nil {
		t.Error("Expected nil response for wrong password")
	}

	// Verificar que el error es de credenciales inválidas
	if errWithCode, ok := err.(*errors.ErrorWithCode); ok {
		if errWithCode.Code != 401 {
			t.Errorf("Expected status code 401, got %d", errWithCode.Code)
		}
	}
}

