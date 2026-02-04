package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// NutritionEstimator は栄養素推定のインターフェース
type NutritionEstimator interface {
	EstimateSingleFood(ctx context.Context, foodName string, quantity int) (*gemini.NutritionInfo, error)
}

// NutritionEstimateHandler は栄養素推定エンドポイントのハンドラー
type NutritionEstimateHandler struct {
	estimator NutritionEstimator
}

// NewNutritionEstimateHandler は新しいNutritionEstimateHandlerを作成
func NewNutritionEstimateHandler(estimator NutritionEstimator) *NutritionEstimateHandler {
	return &NutritionEstimateHandler{
		estimator: estimator,
	}
}

// NutritionEstimateRequest はリクエストボディの構造体
type NutritionEstimateRequest struct {
	FoodName string `json:"food_name"`
	Quantity int    `json:"quantity"`
}

// NutritionEstimateResponse はレスポンスの構造体
type NutritionEstimateResponse struct {
	Name            string  `json:"name"`
	EstimatedAmount string  `json:"estimated_amount"`
	CaloriesKcal    float64 `json:"calories_kcal"`
	ProteinG        float64 `json:"protein_g"`
	FatG            float64 `json:"fat_g"`
	CarbohydratesG  float64 `json:"carbohydrates_g"`
}

// Handle はPOST /api/nutrition/estimateリクエストを処理
func (h *NutritionEstimateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NutritionEstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// バリデーション
	if req.FoodName == "" {
		http.Error(w, "食品名を入力してください", http.StatusBadRequest)
		return
	}
	if len(req.FoodName) > 200 {
		http.Error(w, "食品名は200文字以内で入力してください", http.StatusBadRequest)
		return
	}
	if req.Quantity < 1 {
		req.Quantity = 1
	}
	if req.Quantity > 100 {
		http.Error(w, "数量は100以下で入力してください", http.StatusBadRequest)
		return
	}

	log.Printf("Nutrition estimate request: food_name=%s, quantity=%d", req.FoodName, req.Quantity)

	// 栄養素を推定
	nutrition, err := h.estimator.EstimateSingleFood(r.Context(), req.FoodName, req.Quantity)
	if err != nil {
		log.Printf("Error estimating nutrition: %v", err)
		http.Error(w, "栄養素の推定に失敗しました", http.StatusInternalServerError)
		return
	}

	// レスポンスを作成
	response := NutritionEstimateResponse{
		Name:            nutrition.Name,
		EstimatedAmount: nutrition.EstimatedAmount,
		CaloriesKcal:    nutrition.Calories,
		ProteinG:        nutrition.Protein,
		FatG:            nutrition.Fat,
		CarbohydratesG:  nutrition.Carbohydrates,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Nutrition estimate response sent: %+v", response)
}
