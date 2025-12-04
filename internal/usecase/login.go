package usecase

import (
	"context"

	"user-service/internal/domain"
	"user-service/pkg/errors"
)

type LoginUseCase struct {
	userRepo   domain.UserRepository
	jwtService domain.JWTService
}

func NewLoginUseCase(
	userRepo domain.UserRepository,
	jwtService domain.JWTService,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.NewErrorWithCode(401, "Credenciales inválidas", errors.ErrInvalidCredentials)
	}

	if !user.VerifyPassword(req.Password) {
		return nil, errors.NewErrorWithCode(401, "Credenciales inválidas", errors.ErrInvalidCredentials)
	}

	token, err := uc.jwtService.GenerateToken(user.ID.String())
	if err != nil {
		return nil, errors.NewErrorWithCode(500, "Error al generar token", err)
	}

	return &LoginResponse{
		Token: token,
	}, nil
}

