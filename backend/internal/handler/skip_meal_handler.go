package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// SkipMealHandler は「食べなかった」エンドポイントのハンドラー
type SkipMealHandler struct {
	repository repository.AnalysisRepository
}

// NewSkipMealHandler は新しいSkipMealHandlerを作成
func NewSkipMealHandler(repository repository.AnalysisRepository) *SkipMealHandler {
	return &SkipMealHandler{repository: repository}
}

// SkipMealRequest はリクエストボディ
type SkipMealRequest struct {
	MealType string `json:"meal_type"`
	MealDate string `json:"meal_date"`
}

// SkipMealResponse はレスポンス
type SkipMealResponse struct {
	ID string `json:"id"`
}

// Handle はPOST /api/meals/skipリクエストを処理
func (h *SkipMealHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req SkipMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.MealType == "" {
		http.Error(w, "meal_typeは必須です", http.StatusBadRequest)
		return
	}

	validMealTypes := map[string]bool{"breakfast": true, "lunch": true, "dinner": true, "snack": true}
	if !validMealTypes[req.MealType] {
		http.Error(w, "無効なmeal_typeです", http.StatusBadRequest)
		return
	}

	if req.MealDate == "" {
		http.Error(w, "meal_dateは必須です", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", req.MealDate); err != nil {
		http.Error(w, "meal_dateはYYYY-MM-DD形式で指定してください", http.StatusBadRequest)
		return
	}

	recordID, err := h.repository.CreateSkippedMeal(r.Context(), req.MealType, req.MealDate, &userID)
	if err != nil {
		log.Printf("「食べなかった」記録の保存に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "記録の保存に失敗しました", http.StatusInternalServerError)
		return
	}

	response := SkipMealResponse{
		ID: recordID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}
