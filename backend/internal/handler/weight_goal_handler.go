package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// WeightGoalHandler は目標体重エンドポイントのハンドラー
type WeightGoalHandler struct {
	repository repository.WeightRecordRepository
}

// NewWeightGoalHandler は新しいWeightGoalHandlerを作成
func NewWeightGoalHandler(repository repository.WeightRecordRepository) *WeightGoalHandler {
	return &WeightGoalHandler{repository: repository}
}

// SetWeightGoalRequest は目標体重設定リクエスト
type SetWeightGoalRequest struct {
	TargetWeightKg float64 `json:"target_weight_kg"`
}

// WeightGoalResponse は目標体重のレスポンス
type WeightGoalResponse struct {
	TargetWeightKg float64 `json:"target_weight_kg"`
	UpdatedAt      string  `json:"updated_at"`
}

// HandleGet はGET /api/weight/goalリクエストを処理
func (h *WeightGoalHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	goal, err := h.repository.GetGoal(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting weight goal: %v", err)
		http.Error(w, "目標体重の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if goal == nil {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"goal": nil}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	response := WeightGoalResponse{
		TargetWeightKg: goal.TargetWeightKg,
		UpdatedAt:      goal.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// HandleSet はPUT /api/weight/goalリクエストを処理
func (h *WeightGoalHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req SetWeightGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if req.TargetWeightKg < 20.0 || req.TargetWeightKg > 300.0 {
		http.Error(w, fmt.Sprintf("target_weight_kgは20.0〜300.0の範囲で指定してください"), http.StatusBadRequest)
		return
	}

	goal, err := h.repository.SetGoal(r.Context(), userID, req.TargetWeightKg)
	if err != nil {
		log.Printf("Error setting weight goal: %v", err)
		http.Error(w, "目標体重の設定に失敗しました", http.StatusInternalServerError)
		return
	}

	response := WeightGoalResponse{
		TargetWeightKg: goal.TargetWeightKg,
		UpdatedAt:      goal.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
