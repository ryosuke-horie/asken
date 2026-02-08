package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// WeightGoalHandler は目標体重エンドポイントのハンドラー
type WeightGoalHandler struct {
	repository repository.WeightGoalRepository
}

// NewWeightGoalHandler は新しいWeightGoalHandlerを作成
func NewWeightGoalHandler(repository repository.WeightGoalRepository) *WeightGoalHandler {
	if repository == nil {
		panic("weight goal handler: repository must not be nil")
	}
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

// WeightGoalNullableResponse は目標体重取得のレスポンス（nilを許容）
type WeightGoalNullableResponse struct {
	Goal *WeightGoalResponse `json:"goal"`
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
		log.Printf("Error getting weight goal: userID=%s, error=%v", userID, err)
		http.Error(w, "目標体重の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	var response WeightGoalNullableResponse
	if goal != nil {
		response.Goal = &WeightGoalResponse{
			TargetWeightKg: roundToOneDecimalForJSON(goal.TargetWeightKg),
			UpdatedAt:      goal.UpdatedAt.Format(time.RFC3339),
		}
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, error=%v", userID, err)
	}
}

// HandleSet はPUT /api/weight/goalリクエストを処理
func (h *WeightGoalHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var req SetWeightGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if err := validateWeightKg(req.TargetWeightKg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	goal, err := h.repository.SetGoal(r.Context(), userID, req.TargetWeightKg)
	if err != nil {
		log.Printf("Error setting weight goal: userID=%s, error=%v", userID, err)
		http.Error(w, "目標体重の設定に失敗しました", http.StatusInternalServerError)
		return
	}

	response := WeightGoalResponse{
		TargetWeightKg: roundToOneDecimalForJSON(goal.TargetWeightKg),
		UpdatedAt:      goal.UpdatedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, error=%v", userID, err)
	}
}
