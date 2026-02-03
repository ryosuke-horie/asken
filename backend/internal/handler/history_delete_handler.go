package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// HistoryDeleteHandler は履歴削除エンドポイントのハンドラー
type HistoryDeleteHandler struct {
	repository repository.AnalysisRepository
}

// NewHistoryDeleteHandler は新しいHistoryDeleteHandlerを作成
func NewHistoryDeleteHandler(repository repository.AnalysisRepository) *HistoryDeleteHandler {
	return &HistoryDeleteHandler{
		repository: repository,
	}
}

// Handle はDELETE /api/history/:idリクエストを処理
func (h *HistoryDeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history delete request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// DELETEメソッドのみ許可
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		log.Printf("Authentication failed for %s: %s %s - no Firebase UID in context", r.RemoteAddr, r.Method, r.URL.Path)
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// URLからhistory_idを抽出
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	historyIDStr := pathParts[3]
	historyID, err := uuid.Parse(historyIDStr)
	if err != nil {
		log.Printf("Invalid UUID: %s, error: %v", historyIDStr, err)
		http.Error(w, "Invalid history ID", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting history ID: %s, userID: %s", historyID, userID)

	// リポジトリから履歴を削除（userIDでスコープ）
	err = h.repository.DeleteHistory(r.Context(), userID, historyID)
	if err != nil {
		log.Printf("Error deleting history: %v", err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "History not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete history", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("History deleted successfully: %s", historyID)

	// 成功レスポンスを返却（204 No Content）
	w.WriteHeader(http.StatusNoContent)
	log.Printf("Delete response sent successfully for ID: %s", historyID)
}
