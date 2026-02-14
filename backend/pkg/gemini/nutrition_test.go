package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateNutrition_Success(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	calculator, err := NewNutritionCalculator(60 * time.Second)
	require.NoError(t, err)
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
	mockHTTPClient := &MockGeminiHTTPClient{}
	calculator := NewNutritionCalculatorWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	// 空の食材リスト
	foods := []FoodItem{}

	_, err := calculator.CalculateNutrition(ctx, foods)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "画像から食材を認識できませんでした")
}

func TestNutritionCalculator_CalculateNutrition_MockSuccess(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{
				Response: `[{
					"name": "刺身盛り合わせ",
					"quantity_value": 8,
					"quantity_unit": "切れ",
					"calories_kcal": 200,
					"protein_g": 20,
					"fat_g": 5,
					"carbohydrates_g": 10
				}]`,
			}, nil
		},
	}

	calculator := NewNutritionCalculatorWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
	}

	nutritionList, err := calculator.CalculateNutrition(ctx, foods)

	require.NoError(t, err)
	assert.Len(t, nutritionList, 1)
	assert.Equal(t, "刺身盛り合わせ", nutritionList[0].Name)
	assert.Equal(t, "8切れ", nutritionList[0].EstimatedAmount)
	assert.Equal(t, 200.0, nutritionList[0].Calories)
	assert.Equal(t, 20.0, nutritionList[0].Protein)
	assert.Equal(t, 5.0, nutritionList[0].Fat)
	assert.Equal(t, 10.0, nutritionList[0].Carbohydrates)
}

func TestNutritionCalculator_CalculateNutrition_MockAPIError(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return nil, assert.AnError
		},
	}

	calculator := NewNutritionCalculatorWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
	}

	_, err := calculator.CalculateNutrition(ctx, foods)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini API呼び出しエラー")
}

func TestNutritionCalculator_CalculateNutrition_MockInvalidJSON(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `{invalid json}`}, nil
		},
	}

	calculator := NewNutritionCalculatorWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
	}

	_, err := calculator.CalculateNutrition(ctx, foods)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "パースエラー")
}

func TestNutritionCalculator_CalculateNutrition_MockEmptyResponse(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `[]`}, nil
		},
	}

	calculator := NewNutritionCalculatorWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
	}

	nutritionList, err := calculator.CalculateNutrition(ctx, foods)

	require.NoError(t, err)
	assert.Empty(t, nutritionList)
}

func TestCalculateNutrition_Timeout(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// 非常に短いタイムアウトでテスト
	calculator, err := NewNutritionCalculator(1 * time.Millisecond)
	require.NoError(t, err)
	ctx := context.Background()

	foods := []FoodItem{
		{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
	}

	_, err = calculator.CalculateNutrition(ctx, foods)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}

func TestNewNutritionCalculator_EmptyAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	calculator, err := NewNutritionCalculator(60 * time.Second)

	require.Error(t, err)
	assert.Nil(t, calculator)
	assert.ErrorIs(t, err, ErrEmptyAPIKey)
}

func TestNewNutritionCalculatorWithAPIKey_EmptyAPIKey(t *testing.T) {
	calculator, err := NewNutritionCalculatorWithAPIKey("", 60*time.Second)

	require.Error(t, err)
	assert.Nil(t, calculator)
	assert.ErrorIs(t, err, ErrEmptyAPIKey)
}
