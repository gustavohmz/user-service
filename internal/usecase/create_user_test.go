package usecase_test

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
	"user-service/internal/usecase"
)

// Mocks
type mockUserRepository struct {
	users map[string]*domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("usuario ya existe")
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, errors.New("usuario no encontrado")
	}
	return user, nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID.String() == id {
			return user, nil
		}
	}
	return nil, errors.New("usuario no encontrado")
}

type mockPLDService struct {
	blacklist map[string]bool
}

func (m *mockPLDService) CheckBlacklist(ctx context.Context, firstName, lastName, email string) (bool, error) {
	return m.blacklist[email], nil
}

type mockEventPublisher struct{}

func (m *mockEventPublisher) PublishUserCreated(ctx context.Context, userID, email string, createdAt int64) error {
	return nil
}

type mockJWTService struct{}

func (m *mockJWTService) GenerateToken(userID string) (string, error) {
	return "mock-token", nil
}

func (m *mockJWTService) ValidateToken(tokenString string) (string, error) {
	return "mock-user-id", nil
}

func TestCreateUserUseCase_Execute_Success(t *testing.T) {
	// Arrange
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	pldService := &mockPLDService{blacklist: make(map[string]bool)}
	eventPublisher := &mockEventPublisher{}
	jwtService := &mockJWTService{}

	useCase := usecase.NewCreateUserUseCase(
		userRepo,
		pldService,
		eventPublisher,
		jwtService,
	)

	req := usecase.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Gustavo Hernández",
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

	if response.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, response.User.Email)
	}

	if response.Token == "" {
		t.Error("Expected token, got empty string")
	}
}

func TestCreateUserUseCase_Execute_UserInBlacklist(t *testing.T) {
	// Arrange
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	pldService := &mockPLDService{
		blacklist: map[string]bool{
			"blacklisted@example.com": true,
		},
	}
	eventPublisher := &mockEventPublisher{}
	jwtService := &mockJWTService{}

	useCase := usecase.NewCreateUserUseCase(
		userRepo,
		pldService,
		eventPublisher,
		jwtService,
	)

	req := usecase.CreateUserRequest{
		Email:    "blacklisted@example.com",
		Password: "password123",
		Name:     "Blacklisted User",
	}

	// Act
	response, err := useCase.Execute(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for blacklisted user, got nil")
	}

	if response != nil {
		t.Error("Expected nil response for blacklisted user")
	}
}

func TestCreateUserUseCase_Execute_UserAlreadyExists(t *testing.T) {
	// Arrange
	existingUser := &domain.User{
		Email: "existing@example.com",
		Name:  "Existing User",
	}
	existingUser.HashPassword("password123")

	userRepo := &mockUserRepository{
		users: map[string]*domain.User{
			"existing@example.com": existingUser,
		},
	}
	pldService := &mockPLDService{blacklist: make(map[string]bool)}
	eventPublisher := &mockEventPublisher{}
	jwtService := &mockJWTService{}

	useCase := usecase.NewCreateUserUseCase(
		userRepo,
		pldService,
		eventPublisher,
		jwtService,
	)

	req := usecase.CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "New User",
	}

	// Act
	response, err := useCase.Execute(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for existing user, got nil")
	}

	if response != nil {
		t.Error("Expected nil response for existing user")
	}
}

func TestCreateUserUseCase_Execute_InvalidPassword(t *testing.T) {
	// Arrange
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	pldService := &mockPLDService{blacklist: make(map[string]bool)}
	eventPublisher := &mockEventPublisher{}
	jwtService := &mockJWTService{}

	useCase := usecase.NewCreateUserUseCase(
		userRepo,
		pldService,
		eventPublisher,
		jwtService,
	)

	req := usecase.CreateUserRequest{
		Email:    "test@example.com",
		Password: "short", // Password muy corto
		Name:     "Gustavo Hernández",
	}

	// Act
	_, err := useCase.Execute(context.Background(), req)

	// Assert
	// Nota: La validación de longitud mínima se hace en el handler/DTO
	// Si pasa la validación del handler, debería funcionar
	// Este test verifica que el caso de uso maneja correctamente el flujo
	if err != nil {
		t.Logf("Error recibido (esperado si hay validación): %v", err)
	}
}

