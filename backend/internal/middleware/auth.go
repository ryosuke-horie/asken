package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
)

type contextKey string

const firebaseUIDContextKey contextKey = "firebaseUID"

// Authenticator は認証ミドルウェアの共通インターフェース
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

// AuthMiddleware は Firebase 認証を行うミドルウェア
type AuthMiddleware struct {
	firebaseAuthService *service.FirebaseAuthService
}

func NewAuthMiddleware(firebaseAuthService *service.FirebaseAuthService) *AuthMiddleware {
	return &AuthMiddleware{
		firebaseAuthService: firebaseAuthService,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
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

		idToken := parts[1]
		token, err := m.firebaseAuthService.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("Firebase ID token validation failed: %v", err)
			http.Error(w, "無効なトークンです", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), firebaseUIDContextKey, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetFirebaseUIDFromContext はコンテキストから Firebase UID を取得する
func GetFirebaseUIDFromContext(ctx context.Context) string {
	uid, ok := ctx.Value(firebaseUIDContextKey).(string)
	if !ok {
		return ""
	}
	return uid
}

// SetFirebaseUIDToContext はテスト用に Firebase UID をコンテキストに設定する
func SetFirebaseUIDToContext(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, firebaseUIDContextKey, uid)
}
