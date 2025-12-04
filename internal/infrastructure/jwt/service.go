package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	domain "user-service/internal/domain"
)

type jwtService struct {
	secretKey string
	expiresIn time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey string, expiresInHours int) domain.JWTService {
	return &jwtService{
		secretKey: secretKey,
		expiresIn: time.Duration(expiresInHours) * time.Hour,
	}
}

func (s *jwtService) GenerateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("error al firmar token: %w", err)
	}

	return tokenString, nil
}

func (s *jwtService) ValidateToken(tokenString string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("error al parsear token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token inválido")
	}

	return claims.UserID, nil
}

