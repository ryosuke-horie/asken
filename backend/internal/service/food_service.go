package service

import (
	"context"
	"fmt"

	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
)

// GeminiClient はGemini APIクライアントのインターフェース（モック可能）
type GeminiClient interface {
	ClassifyFoods(ctx context.Context, imagePath string) ([]gemini.FoodItem, error)
	CalculateNutrition(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error)
}

// AnalysisResult は分析結果を表す構造体
type AnalysisResult struct {
	Foods              []gemini.NutritionInfo `json:"foods"`
	TotalCalories      float64                `json:"total_calories"`
	TotalProtein       float64                `json:"total_protein"`
	TotalFat           float64                `json:"total_fat"`
	TotalCarbohydrates float64                `json:"total_carbohydrates"`
}

// FoodService は食品分析サービス
type FoodService struct {
	geminiClient GeminiClient
}

// NewFoodService は新しいFoodServiceを作成
func NewFoodService(geminiClient GeminiClient) *FoodService {
	return &FoodService{
		geminiClient: geminiClient,
	}
}

// AnalyzeFoodImage は画像から食材を分析し、栄養素を計算する
func (s *FoodService) AnalyzeFoodImage(ctx context.Context, imagePath string) (*AnalysisResult, error) {
	// Step 1: 食材分類
	foods, err := s.geminiClient.ClassifyFoods(ctx, imagePath)
	if err != nil {
		return nil, fmt.Errorf("食材分類エラー: %w", err)
	}

	// Step 2: 栄養素計算
	nutritionList, err := s.geminiClient.CalculateNutrition(ctx, foods)
	if err != nil {
		return nil, fmt.Errorf("栄養素計算エラー: %w", err)
	}

	// 合計カロリー・栄養素を計算
	totalCal, totalPro, totalFat, totalCarbs := calculateTotals(nutritionList)

	// AnalysisResultを返却
	return &AnalysisResult{
		Foods:              nutritionList,
		TotalCalories:      totalCal,
		TotalProtein:       totalPro,
		TotalFat:           totalFat,
		TotalCarbohydrates: totalCarbs,
	}, nil
}

// calculateTotals は栄養素の合計を計算する
func calculateTotals(nutritionList []gemini.NutritionInfo) (totalCal, totalPro, totalFat, totalCarbs float64) {
	for _, nutrition := range nutritionList {
		totalCal += nutrition.Calories
		totalPro += nutrition.Protein
		totalFat += nutrition.Fat
		totalCarbs += nutrition.Carbohydrates
	}
	return
}
