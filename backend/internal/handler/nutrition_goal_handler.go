package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// NutritionGoalHandler は栄養目標エンドポイントのハンドラー
type NutritionGoalHandler struct {
	nutritionGoalRepo repository.NutritionGoalRepository
	weightGoalRepo    repository.WeightGoalRepository
}

// NewNutritionGoalHandler は新しいNutritionGoalHandlerを作成
func NewNutritionGoalHandler(nutritionGoalRepo repository.NutritionGoalRepository, weightGoalRepo repository.WeightGoalRepository) *NutritionGoalHandler {
	if nutritionGoalRepo == nil {
		panic("nutrition goal handler: nutritionGoalRepo must not be nil")
	}
	if weightGoalRepo == nil {
		panic("nutrition goal handler: weightGoalRepo must not be nil")
	}
	return &NutritionGoalHandler{
		nutritionGoalRepo: nutritionGoalRepo,
		weightGoalRepo:    weightGoalRepo,
	}
}

// SetNutritionGoalRequest は目標カロリー設定リクエスト
type SetNutritionGoalRequest struct {
	TargetCalories float64 `json:"target_calories"`
}

// NutritionGoalResponse は栄養目標のレスポンス
type NutritionGoalResponse struct {
	TargetCalories       float64            `json:"target_calories"`
	TargetProtein        float64            `json:"target_protein"`
	TargetFat            float64            `json:"target_fat"`
	TargetCarbohydrates  float64            `json:"target_carbohydrates"`
	Phase                string             `json:"phase"`
	MicronutrientTargets map[string]float64 `json:"micronutrient_targets,omitempty"`
	UpdatedAt            string             `json:"updated_at"`
}

// NutritionGoalNullableResponse は栄養目標取得のレスポンス（nilを許容）
type NutritionGoalNullableResponse struct {
	Goal *NutritionGoalResponse `json:"goal"`
}

// parseOptionalWeight はクエリパラメータの体重値をパース・検証する。
// 値が空の場合はnilを返す。パースエラー時はHTTPレスポンスを書き込みエラーを返す。
func parseOptionalWeight(w http.ResponseWriter, raw string, paramName string) (*float64, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		log.Printf("Error parsing %s: value=%s, error=%v", paramName, raw, err)
		http.Error(w, paramName+"のパースに失敗しました", http.StatusBadRequest)
		return nil, err
	}
	if err := repository.ValidateWeightKg(parsed); err != nil {
		log.Printf("Validation error for %s: value=%.1f, error=%v", paramName, parsed, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	return &parsed, nil
}

// HandleGet はGET /api/nutrition/goalリクエストを処理
func (h *NutritionGoalHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	currentWeight, err := parseOptionalWeight(w, r.URL.Query().Get("current_weight"), "current_weight")
	if err != nil {
		return
	}

	targetWeight, err := parseOptionalWeight(w, r.URL.Query().Get("target_weight"), "target_weight")
	if err != nil {
		return
	}

	// 目標体重が指定されていない場合はリポジトリから取得
	if targetWeight == nil {
		weightGoal, goalErr := h.weightGoalRepo.GetGoal(r.Context(), userID)
		if goalErr != nil {
			log.Printf("Error getting weight goal: userID=%s, error=%v", userID, goalErr)
			http.Error(w, "目標体重の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		if weightGoal != nil {
			targetWeight = &weightGoal.TargetWeightKg
		}
	}

	goal, err := h.nutritionGoalRepo.GetGoal(r.Context(), userID, currentWeight, targetWeight)
	if err != nil {
		log.Printf("Error getting nutrition goal: userID=%s, error=%v", userID, err)
		http.Error(w, "栄養目標の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	var response NutritionGoalNullableResponse
	if goal != nil {
		response.Goal = &NutritionGoalResponse{
			TargetCalories:       roundToOneDecimalForJSON(goal.TargetCalories),
			TargetProtein:        roundToOneDecimalForJSON(goal.TargetProtein),
			TargetFat:            roundToOneDecimalForJSON(goal.TargetFat),
			TargetCarbohydrates:  roundToOneDecimalForJSON(goal.TargetCarbohydrates),
			Phase:                string(goal.Phase),
			MicronutrientTargets: goal.MicronutrientTargets,
			UpdatedAt:            goal.UpdatedAt.Format(time.RFC3339),
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(response); err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("Error writing response: userID=%s, error=%v", userID, err)
	}
}

// HandleSet はPUT /api/nutrition/goalリクエストを処理
func (h *NutritionGoalHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var req SetNutritionGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if err := repository.ValidateTargetCalories(req.TargetCalories); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	goal, err := h.nutritionGoalRepo.SetGoal(r.Context(), userID, req.TargetCalories)
	if err != nil {
		log.Printf("Error setting nutrition goal: userID=%s, error=%v", userID, err)
		http.Error(w, "栄養目標の設定に失敗しました", http.StatusInternalServerError)
		return
	}

	response := NutritionGoalResponse{
		TargetCalories:       roundToOneDecimalForJSON(goal.TargetCalories),
		TargetProtein:        roundToOneDecimalForJSON(goal.TargetProtein),
		TargetFat:            roundToOneDecimalForJSON(goal.TargetFat),
		TargetCarbohydrates:  roundToOneDecimalForJSON(goal.TargetCarbohydrates),
		Phase:                string(goal.Phase),
		MicronutrientTargets: goal.MicronutrientTargets,
		UpdatedAt:            goal.UpdatedAt.Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(response); err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("Error writing response: userID=%s, error=%v", userID, err)
	}
}
