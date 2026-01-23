package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// HistoryHandler は履歴取得エンドポイントのハンドラー
type HistoryHandler struct {
	repository repository.AnalysisRepository
}

// NewHistoryHandler は新しいHistoryHandlerを作成
func NewHistoryHandler(repository repository.AnalysisRepository) *HistoryHandler {
	return &HistoryHandler{
		repository: repository,
	}
}

// HandleList はGET /api/historyリクエストを処理（履歴一覧取得）
func (h *HistoryHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history list request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// クエリパラメータからpage, limitを取得
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	log.Printf("Fetching history list: page=%d, limit=%d", page, limit)

	// リポジトリから履歴一覧を取得
	items, total, err := h.repository.GetHistoryList(r.Context(), page, limit)
	if err != nil {
		log.Printf("Error getting history list: %v", err)
		http.Error(w, "Failed to get history list", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d items, total=%d", len(items), total)

	// レスポンスを生成
	response := map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	// JSONレスポンスを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("History list response sent successfully")
}

// HandleDetail はGET /api/history/:idリクエストを処理（履歴詳細取得）
func (h *HistoryHandler) HandleDetail(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history detail request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	log.Printf("Getting history detail for ID: %s", historyID)

	// リポジトリから履歴詳細を取得
	detail, err := h.repository.GetHistoryDetail(r.Context(), historyID)
	if err != nil {
		log.Printf("Error getting history detail: %v", err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "History not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get history detail", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("History detail retrieved successfully")

	// JSONレスポンスを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("History detail response sent successfully for ID: %s", historyID)
}
