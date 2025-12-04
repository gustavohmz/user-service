package usecase_test

import (
	"context"
	"testing"
	"time"

	"user-service/internal/domain"
	"user-service/internal/usecase"
	"user-service/pkg/errors"
	"github.com/google/uuid"
)

func TestGetUserUseCase_Execute_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	user := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Name:      "Gustavo Hernández",
		CreatedAt: time.Now(),
	}

	userRepo := &mockUserRepository{
		users: map[string]*domain.User{
			userID.String(): user,
		},
	}

	useCase := usecase.NewGetUserUseCase(userRepo)

	// Act
	response, err := useCase.Execute(context.Background(), userID.String())

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.User == nil {
		t.Fatal("Expected user in response, got nil")
	}

	if response.User.ID != userID.String() {
		t.Errorf("Expected user ID %s, got %s", userID.String(), response.User.ID)
	}

	if response.User.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, response.User.Email)
	}

	if response.User.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, response.User.Name)
	}
}

func TestGetUserUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	useCase := usecase.NewGetUserUseCase(userRepo)

	nonexistentID := uuid.New().String()

	// Act
	response, err := useCase.Execute(context.Background(), nonexistentID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nonexistent user, got nil")
	}

	if response != nil {
		t.Error("Expected nil response for nonexistent user")
	}

	// Verificar que el error es de usuario no encontrado
	if errWithCode, ok := err.(*errors.ErrorWithCode); ok {
		if errWithCode.Code != 404 {
			t.Errorf("Expected status code 404, got %d", errWithCode.Code)
		}
	}
}

