package usecase

import (
	"context"
	"strings"
	"time"

	"user-service/internal/domain"
	"user-service/pkg/errors"
)

type CreateUserUseCase struct {
	userRepo      domain.UserRepository
	pldService    domain.PLDService
	eventPublisher domain.EventPublisher
	jwtService    domain.JWTService
}

func NewCreateUserUseCase(
	userRepo domain.UserRepository,
	pldService domain.PLDService,
	eventPublisher domain.EventPublisher,
	jwtService domain.JWTService,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:      userRepo,
		pldService:    pldService,
		eventPublisher: eventPublisher,
		jwtService:    jwtService,
	}
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

type CreateUserResponse struct {
	User  *UserDTO `json:"user"`
	Token string   `json:"token"`
}

type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	existingUser, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.NewErrorWithCode(409, "El usuario ya existe", errors.ErrUserAlreadyExists)
	}

	firstName, lastName := splitName(req.Name)
	inBlacklist, err := uc.pldService.CheckBlacklist(ctx, firstName, lastName, req.Email)
	if err != nil {
		return nil, errors.NewErrorWithCode(500, "Error al verificar PLD", err)
	}
	if inBlacklist {
		return nil, errors.NewErrorWithCode(403, "Usuario en lista negra", errors.ErrUserInBlacklist)
	}

	user := &domain.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, errors.NewErrorWithCode(500, "Error al procesar contraseña", err)
	}

	if err := user.Validate(); err != nil {
		return nil, errors.NewErrorWithCode(400, "Datos inválidos", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, errors.NewErrorWithCode(500, "Error al crear usuario", err)
	}

	token, err := uc.jwtService.GenerateToken(user.ID.String())
	if err != nil {
		return nil, errors.NewErrorWithCode(500, "Error al generar token", err)
	}

	go func() {
		eventCtx := context.Background()
		uc.eventPublisher.PublishUserCreated(
			eventCtx,
			user.ID.String(),
			user.Email,
			user.CreatedAt.Unix(),
		)
	}()

	return &CreateUserResponse{
		User: &UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}, nil
}

func splitName(name string) (firstName, lastName string) {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	firstName = parts[0]
	lastName = strings.Join(parts[1:], " ")
	return firstName, lastName
}

