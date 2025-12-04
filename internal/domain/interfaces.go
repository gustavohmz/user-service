package domain

import (
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type PLDService interface {
	CheckBlacklist(ctx context.Context, firstName, lastName, email string) (bool, error)
}

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, userID, email string, createdAt int64) error
}

type EventConsumer interface {
	ConsumeUserCreated(ctx context.Context, handler func(userID, email string, createdAt int64) error) error
}

type JWTService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(tokenString string) (string, error)
}

type UserEventRepository interface {
	Create(ctx context.Context, event *UserEvent) error
}

