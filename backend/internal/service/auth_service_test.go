package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_HashPassword(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	t.Run("パスワードをハッシュ化すべき", func(t *testing.T) {
		password := "securePassword123"

		hash, err := authService.HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
	})

	t.Run("同じパスワードでも異なるハッシュを生成すべき", func(t *testing.T) {
		password := "securePassword123"

		hash1, err1 := authService.HashPassword(password)
		hash2, err2 := authService.HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestAuthService_VerifyPassword(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	t.Run("正しいパスワードを検証すべき", func(t *testing.T) {
		password := "securePassword123"
		hash, err := authService.HashPassword(password)
		require.NoError(t, err)

		result := authService.VerifyPassword(hash, password)

		assert.True(t, result)
	})

	t.Run("誤ったパスワードを拒否すべき", func(t *testing.T) {
		password := "securePassword123"
		wrongPassword := "wrongPassword"
		hash, err := authService.HashPassword(password)
		require.NoError(t, err)

		result := authService.VerifyPassword(hash, wrongPassword)

		assert.False(t, result)
	})
}

func TestAuthService_GenerateToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	t.Run("有効なJWTトークンを生成すべき", func(t *testing.T) {
		userID := uuid.New()

		token, err := authService.GenerateToken(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	t.Run("有効なトークンを検証すべき", func(t *testing.T) {
		userID := uuid.New()
		token, err := authService.GenerateToken(userID)
		require.NoError(t, err)

		claims, err := authService.ValidateToken(token)

		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("無効なトークンを拒否すべき", func(t *testing.T) {
		invalidToken := "invalid.token.here"

		claims, err := authService.ValidateToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("異なるシークレットで署名されたトークンを拒否すべき", func(t *testing.T) {
		otherAuthService := NewAuthService("other-secret", 24*time.Hour)
		userID := uuid.New()
		token, err := otherAuthService.GenerateToken(userID)
		require.NoError(t, err)

		claims, err := authService.ValidateToken(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("期限切れのトークンを拒否すべき", func(t *testing.T) {
		shortLivedAuthService := NewAuthService("test-secret", -1*time.Hour)
		userID := uuid.New()
		token, err := shortLivedAuthService.GenerateToken(userID)
		require.NoError(t, err)

		claims, err := authService.ValidateToken(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
