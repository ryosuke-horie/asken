package handler

import (
	"log"
	"net/http"
)

// HealthHandler はヘルスチェックエンドポイントのハンドラー
type HealthHandler struct{}

// NewHealthHandler は新しいHealthHandlerを作成
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Handle はGET /api/healthリクエストを処理
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		log.Printf("Health check response write error: %v", err)
	}
}
