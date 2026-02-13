//go:build production

package middleware

import (
	"log"
	"net/http"
)

// DevAuthMiddleware は本番ビルドでは無効化された開発用認証ミドルウェア
type DevAuthMiddleware struct{}

// NewDevAuthMiddleware は本番ビルドでは常にログ出力して安全な実装を返す
func NewDevAuthMiddleware() *DevAuthMiddleware {
	log.Println("WARNING: DevAuthMiddleware is disabled in production build")
	return &DevAuthMiddleware{}
}

// Authenticate は本番ビルドでは常に認証を拒否する
func (m *DevAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "開発用認証は本番環境では無効です", http.StatusForbidden)
	})
}
