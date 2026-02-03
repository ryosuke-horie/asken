package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// StatusHandler は分析ステータス取得エンドポイントのハンドラー
type StatusHandler struct {
	repository repository.AnalysisRepository
}

// NewStatusHandler は新しいStatusHandlerを作成
func NewStatusHandler(repository repository.AnalysisRepository) *StatusHandler {
	return &StatusHandler{
		repository: repository,
	}
}

// Handle はGET /api/analyze/:idリクエストを処理
func (h *StatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received status request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// 1. URLからanalysis_idを抽出
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	analysisIDStr := pathParts[3]
	analysisID, err := uuid.Parse(analysisIDStr)
	if err != nil {
		log.Printf("Invalid UUID: %s, error: %v", analysisIDStr, err)
		http.Error(w, "Invalid analysis ID", http.StatusBadRequest)
		return
	}

	log.Printf("Getting status for analysis ID: %s, userID: %s", analysisID, userID)

	// 2. リポジトリからリクエストを取得（userIDでスコープ）
	request, err := h.repository.GetRequest(r.Context(), userID, analysisID)
	if err != nil {
		log.Printf("Error getting request: %v", err)
		http.Error(w, "Analysis request not found", http.StatusNotFound)
		return
	}

	log.Printf("Request found with status: %s", request.Status)

	// 3. ステータスに応じてレスポンスを生成
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var response map[string]interface{}

	switch request.Status {
	case repository.StatusPending:
		response = map[string]interface{}{
			"status":  "pending",
			"message": "分析リクエストを受け付けました",
		}

	case repository.StatusProcessing:
		response = map[string]interface{}{
			"status":  "processing",
			"message": "分析処理中です",
		}

	case repository.StatusCompleted:
		// 結果を取得（userIDでスコープ）
		result, err := h.repository.GetResult(r.Context(), userID, analysisID)
		if err != nil {
			log.Printf("Error getting result: %v", err)
			http.Error(w, "Failed to get analysis result", http.StatusInternalServerError)
			return
		}

		response = map[string]interface{}{
			"status": "completed",
			"result": result,
		}

	case repository.StatusFailed:
		response = map[string]interface{}{
			"status": "failed",
			"error":  request.ErrorMessage,
		}

	default:
		log.Printf("Unknown status: %s", request.Status)
		http.Error(w, "Unknown status", http.StatusInternalServerError)
		return
	}

	// 4. JSONレスポンスを返却
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Status response sent successfully for analysis ID: %s", analysisID)
}
