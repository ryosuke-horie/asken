package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockTokenVerifier はテスト用のTokenVerifier実装
type mockTokenVerifier struct {
	verifyFunc func(ctx context.Context, idToken string) (string, error)
}

func (m *mockTokenVerifier) VerifyAndGetUID(ctx context.Context, idToken string) (string, error) {
	return m.verifyFunc(ctx, idToken)
}

func TestSetFirebaseUIDToContext(t *testing.T) {
	t.Run("Firebase UIDをコンテキストに設定して取得できるべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := SetFirebaseUIDToContext(req.Context(), "test-firebase-uid")
		req = req.WithContext(ctx)

		uid := GetFirebaseUIDFromContext(req.Context())
		assert.Equal(t, "test-firebase-uid", uid)
	})

	t.Run("空のUIDを設定できるべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := SetFirebaseUIDToContext(req.Context(), "")
		req = req.WithContext(ctx)

		uid := GetFirebaseUIDFromContext(req.Context())
		assert.Equal(t, "", uid)
	})
}

func TestGetFirebaseUIDFromContext(t *testing.T) {
	t.Run("UIDが設定されていない場合は空文字を返すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

		uid := GetFirebaseUIDFromContext(req.Context())
		assert.Equal(t, "", uid)
	})
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	t.Run("Authorizationヘッダーなしで401を返すべき", func(t *testing.T) {
		verifier := &mockTokenVerifier{}
		m := NewAuthMiddleware(verifier)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()

		m.Authenticate(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, nextCalled)
	})

	t.Run("Bearer形式でないヘッダーで401を返すべき", func(t *testing.T) {
		verifier := &mockTokenVerifier{}
		m := NewAuthMiddleware(verifier)

		tests := []struct {
			name   string
			header string
		}{
			{"Basic認証形式", "Basic dXNlcjpwYXNz"},
			{"Bearer以外のプレフィックス", "Token abc123"},
			{"スペースなし", "Bearerabc123"},
			{"トークン部分が複数", "Bearer abc 123"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					nextCalled = true
				})

				req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
				req.Header.Set("Authorization", tt.header)
				w := httptest.NewRecorder()

				m.Authenticate(next).ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
				assert.False(t, nextCalled)
			})
		}
	})

	t.Run("トークン検証失敗で401を返すべき", func(t *testing.T) {
		verifier := &mockTokenVerifier{
			verifyFunc: func(ctx context.Context, idToken string) (string, error) {
				return "", errors.New("invalid token")
			},
		}
		m := NewAuthMiddleware(verifier)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		m.Authenticate(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, nextCalled)
	})

	t.Run("有効なトークンでnext handlerが呼ばれcontextにUIDが設定されるべき", func(t *testing.T) {
		expectedUID := "test-uid-12345"
		verifier := &mockTokenVerifier{
			verifyFunc: func(ctx context.Context, idToken string) (string, error) {
				assert.Equal(t, "valid-token", idToken)
				return expectedUID, nil
			},
		}
		m := NewAuthMiddleware(verifier)

		var capturedUID string
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			capturedUID = GetFirebaseUIDFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		m.Authenticate(next).ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, expectedUID, capturedUID)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIsDevMode(t *testing.T) {
	t.Run("APP_ENV=developmentでtrueを返すべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Setenv("APP_ENV", "development")
		assert.True(t, IsDevMode())
	})

	t.Run("APP_ENV未設定でfalseを返すべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Unsetenv("APP_ENV")
		assert.False(t, IsDevMode())
	})

	t.Run("APP_ENV=productionでfalseを返すべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Setenv("APP_ENV", "production")
		assert.False(t, IsDevMode())
	})

	t.Run("APP_ENVが他の値でfalseを返すべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Setenv("APP_ENV", "staging")
		assert.False(t, IsDevMode())
	})
}
