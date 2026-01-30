package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateNutrition_Success(t *testing.T) {
	skipIfNoGeminiCLI(t)

	calculator := NewNutritionCalculator(60 * time.Second)
	ctx := context.Background()

	// テスト用の食材リスト
	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
		{Name: "サラダ", EstimatedAmount: "1皿"},
	}

	nutritionList, err := calculator.CalculateNutrition(ctx, foods)

	require.NoError(t, err)
	assert.NotEmpty(t, nutritionList)
	assert.Equal(t, len(foods), len(nutritionList))

	// 各栄養素情報が適切な構造を持っているか確認
	for _, nutrition := range nutritionList {
		assert.NotEmpty(t, nutrition.Name, "食材名が空です")
		assert.NotEmpty(t, nutrition.EstimatedAmount, "推定量が空です")
		assert.Greater(t, nutrition.Calories, 0.0, "カロリーが0以下です")
		assert.GreaterOrEqual(t, nutrition.Protein, 0.0, "タンパク質が負の値です")
		assert.GreaterOrEqual(t, nutrition.Fat, 0.0, "脂質が負の値です")
		assert.GreaterOrEqual(t, nutrition.Carbohydrates, 0.0, "炭水化物が負の値です")
	}
}

func TestCalculateNutrition_EmptyFoods(t *testing.T) {
	calculator := NewNutritionCalculator(60 * time.Second)
	ctx := context.Background()

	// 空の食材リスト
	foods := []FoodItem{}

	_, err := calculator.CalculateNutrition(ctx, foods)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "画像から食材を認識できませんでした")
}

func TestCalculateNutrition_InvalidResponse(t *testing.T) {
	// このテストは実際のGemini CLIの挙動に依存するためスキップ
	t.Skip("Gemini CLIの挙動に依存するためスキップ")
}

func TestCalculateNutrition_Timeout(t *testing.T) {
	skipIfNoGeminiCLI(t)

	// 非常に短いタイムアウトでテスト
	calculator := NewNutritionCalculator(1 * time.Millisecond)
	ctx := context.Background()

	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
	}

	_, err := calculator.CalculateNutrition(ctx, foods)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}
