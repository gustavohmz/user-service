package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"` // Hash bcrypt
	Name      string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("email es requerido")
	}
	if len(u.Password) == 0 {
		return fmt.Errorf("password es requerido")
	}
	if u.Name == "" {
		return fmt.Errorf("nombre es requerido")
	}
	if len(u.Email) < 5 {
		return fmt.Errorf("email inválido")
	}
	return nil
}

