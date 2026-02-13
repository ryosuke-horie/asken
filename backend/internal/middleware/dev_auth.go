//go:build !production

package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
)

// Development auth constants - must match iOS MockFirebaseAuthService
const (
	DevMockToken  = "dev-mock-token"
	DevMockUserID = "dev-mock-user"
)

// DevAuthMiddleware は開発環境用の認証バイパスミドルウェア
type DevAuthMiddleware struct{}

// NewDevAuthMiddleware は DevAuthMiddleware を作成する
func NewDevAuthMiddleware() *DevAuthMiddleware {
	return &DevAuthMiddleware{}
}

// Authenticate は開発用トークンを検証し、固定のUIDを設定する
func (m *DevAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "無効な認証形式です", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != DevMockToken {
			maskedToken := token
			if len(token) > 8 {
				maskedToken = token[:4] + "..." + token[len(token)-4:]
			}
			log.Printf("[DEV] Invalid dev token received: %s", maskedToken)
			http.Error(w, "無効な開発用トークンです", http.StatusUnauthorized)
			return
		}

		log.Printf("[DEV] Mock auth successful for path: %s", r.URL.Path)
		ctx := context.WithValue(r.Context(), firebaseUIDContextKey, DevMockUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
