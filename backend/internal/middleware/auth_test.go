package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	authService := service.NewAuthService("test-secret", 24*time.Hour)

	t.Run("有効なトークンでリクエストを許可すべき", func(t *testing.T) {
		userID := uuid.New()
		token, err := authService.GenerateToken(userID)
		require.NoError(t, err)

		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			ctxUserID := GetUserIDFromContext(r.Context())
			assert.Equal(t, userID, ctxUserID)
			w.WriteHeader(http.StatusOK)
		})

		middleware := NewAuthMiddleware(authService)
		wrappedHandler := middleware.Authenticate(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authorizationヘッダーがない場合は拒否すべき", func(t *testing.T) {
		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware := NewAuthMiddleware(authService)
		wrappedHandler := middleware.Authenticate(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("無効なトークンを拒否すべき", func(t *testing.T) {
		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware := NewAuthMiddleware(authService)
		wrappedHandler := middleware.Authenticate(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Bearerプレフィックスがない場合は拒否すべき", func(t *testing.T) {
		userID := uuid.New()
		token, err := authService.GenerateToken(userID)
		require.NoError(t, err)

		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware := NewAuthMiddleware(authService)
		wrappedHandler := middleware.Authenticate(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("期限切れのトークンを拒否すべき", func(t *testing.T) {
		expiredAuthService := service.NewAuthService("test-secret", -1*time.Hour)
		userID := uuid.New()
		token, err := expiredAuthService.GenerateToken(userID)
		require.NoError(t, err)

		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware := NewAuthMiddleware(authService)
		wrappedHandler := middleware.Authenticate(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
