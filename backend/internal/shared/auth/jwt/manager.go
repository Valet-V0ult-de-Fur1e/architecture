package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secret []byte
}

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (m *Manager) Generate(userID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return signedToken, nil
}

func (m *Manager) Parse(tokenRaw string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenRaw, &claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return m.secret, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse jwt: %w", err)
	}

	tokenClaims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid jwt claims")
	}

	userID, err := uuid.Parse(tokenClaims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid jwt user id: %w", err)
	}

	return userID, nil
}
