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
	TargetCalories      float64 `json:"target_calories"`
	TargetProtein       float64 `json:"target_protein"`
	TargetFat           float64 `json:"target_fat"`
	TargetCarbohydrates float64 `json:"target_carbohydrates"`
	Phase               string  `json:"phase"`
	UpdatedAt           string  `json:"updated_at"`
}

// NutritionGoalNullableResponse は栄養目標取得のレスポンス（nilを許容）
type NutritionGoalNullableResponse struct {
	Goal *NutritionGoalResponse `json:"goal"`
}

// HandleGet はGET /api/nutrition/goalリクエストを処理
func (h *NutritionGoalHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// クエリパラメータから現在体重と目標体重を取得
	targetWeightStr := r.URL.Query().Get("target_weight")

	var currentWeight *float64
	if currentWeightStr := r.URL.Query().Get("current_weight"); currentWeightStr != "" {
		parsed, err := strconv.ParseFloat(currentWeightStr, 64)
		if err != nil {
			http.Error(w, "current_weightのパースに失敗しました", http.StatusBadRequest)
			return
		}
		// 範囲チェック
		if err := repository.ValidateWeightKg(parsed); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		currentWeight = &parsed
	}

	var targetWeight *float64

	if targetWeightStr != "" {
		parsed, err := strconv.ParseFloat(targetWeightStr, 64)
		if err != nil {
			http.Error(w, "target_weightのパースに失敗しました", http.StatusBadRequest)
			return
		}
		// 範囲チェック
		if err := repository.ValidateWeightKg(parsed); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		targetWeight = &parsed
	}

	// 目標体重が指定されていない場合はリポジトリから取得
	if targetWeight == nil {
		weightGoal, err := h.weightGoalRepo.GetGoal(r.Context(), userID)
		if err != nil {
			log.Printf("Error getting weight goal: userID=%s, error=%v", userID, err)
			http.Error(w, "目標体重の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		if weightGoal != nil {
			targetWeight = &weightGoal.TargetWeightKg
		}
	}

	goal, getGoalErr := h.nutritionGoalRepo.GetGoal(r.Context(), userID, currentWeight, targetWeight)
	if getGoalErr != nil {
		log.Printf("Error getting nutrition goal: userID=%s, error=%v", userID, getGoalErr)
		http.Error(w, "栄養目標の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	var response NutritionGoalNullableResponse
	if goal != nil {
		response.Goal = &NutritionGoalResponse{
			TargetCalories:      roundToOneDecimalForJSON(goal.TargetCalories),
			TargetProtein:       roundToOneDecimalForJSON(goal.TargetProtein),
			TargetFat:           roundToOneDecimalForJSON(goal.TargetFat),
			TargetCarbohydrates: roundToOneDecimalForJSON(goal.TargetCarbohydrates),
			Phase:               string(goal.Phase),
			UpdatedAt:           goal.UpdatedAt.Format(time.RFC3339),
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(response); err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// HandleSet はPUT /api/nutrition/goalリクエストを処理
func (h *NutritionGoalHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
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
		TargetCalories:      roundToOneDecimalForJSON(goal.TargetCalories),
		TargetProtein:       roundToOneDecimalForJSON(goal.TargetProtein),
		TargetFat:           roundToOneDecimalForJSON(goal.TargetFat),
		TargetCarbohydrates: roundToOneDecimalForJSON(goal.TargetCarbohydrates),
		Phase:               string(goal.Phase),
		UpdatedAt:           goal.UpdatedAt.Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(response); err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}
