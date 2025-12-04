package usecase

import (
	"context"

	"user-service/internal/domain"
	"user-service/pkg/errors"
)

type GetUserUseCase struct {
	userRepo domain.UserRepository
}

func NewGetUserUseCase(userRepo domain.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
	}
}

type GetUserResponse struct {
	User *UserDTO `json:"user"`
}

func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*GetUserResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.NewErrorWithCode(404, "Usuario no encontrado", errors.ErrUserNotFound)
	}

	return &GetUserResponse{
		User: &UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

