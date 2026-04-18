package domain_test

import (
	"architecture/backend/internal/modules/identity/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMockUser_BIVT_23_SP_3_21(t *testing.T) {
	user := domain.User{
		ID:           uuid.New(),
		Email:        "test21@bivt23sp3.ru",
		PasswordHash: "hashedpassword",
		Name:         "Тестовый Пользователь",
		Group:        "БИВТ-23-СП-3",
		Number:       21,
		CreatedAt:    time.Now(),
	}
	require.Equal(t, "БИВТ-23-СП-3", user.Group)
	require.Equal(t, 21, user.Number)
	require.Equal(t, "test21@bivt23sp3.ru", user.Email)
}
