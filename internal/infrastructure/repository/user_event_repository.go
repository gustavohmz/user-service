package repository

import (
	"context"
	"fmt"

	"user-service/internal/domain"
	"gorm.io/gorm"
)

type userEventRepository struct {
	db *gorm.DB
}

func NewUserEventRepository(db *gorm.DB) domain.UserEventRepository {
	return &userEventRepository{db: db}
}

func (r *userEventRepository) Create(ctx context.Context, event *domain.UserEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("error al crear evento: %w", err)
	}
	return nil
}

