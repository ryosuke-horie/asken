//go:build !production

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDevAuthMiddleware_Authenticate(t *testing.T) {
	middleware := NewDevAuthMiddleware()

	t.Run("開発用トークンで認証成功すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+DevMockToken)
		rec := httptest.NewRecorder()

		var capturedUID string
		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUID = GetFirebaseUIDFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, DevMockUserID, capturedUID)
	})

	t.Run("無効なトークンは拒否すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("ハンドラーが呼び出されるべきではない")
		}))

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "無効な開発用トークンです")
	})

	t.Run("Authorizationヘッダーがない場合は拒否すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("ハンドラーが呼び出されるべきではない")
		}))

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "認証が必要です")
	})

	t.Run("Bearer形式以外は拒否すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Basic "+DevMockToken)
		rec := httptest.NewRecorder()

		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("ハンドラーが呼び出されるべきではない")
		}))

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "無効な認証形式です")
	})

	t.Run("Bearerのみでトークンがない場合は拒否すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer")
		rec := httptest.NewRecorder()

		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("ハンドラーが呼び出されるべきではない")
		}))

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "無効な認証形式です")
	})
}
