package gemini

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestMenus_Success(t *testing.T) {
	responseJSON := `{
		"suggestions": [
			{
				"title": "鶏むね肉と野菜の炒め物",
				"description": "高タンパク低脂質の定番メニュー",
				"reason": "タンパク質不足を補うために提案",
				"ingredients": [
					{"name": "鶏むね肉", "quantity": 200.0, "unit": "g"},
					{"name": "ブロッコリー", "quantity": 100.0, "unit": "g"}
				],
				"estimatedNutrition": {
					"calories": 350.0,
					"protein": 40.0,
					"fat": 8.0,
					"carbohydrates": 15.0
				}
			}
		]
	}`

	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: responseJSON}, nil
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)
	input := MenuSuggestionInput{
		MealType: "lunch",
		Count:    1,
		Ingredients: []IngredientContext{
			{ID: "ing-1", Name: "鶏むね肉", Quantity: 300, Unit: "g"},
		},
	}

	suggestions, err := suggester.SuggestMenus(context.Background(), input)

	require.NoError(t, err)
	require.Len(t, suggestions, 1)
	assert.Equal(t, "鶏むね肉と野菜の炒め物", suggestions[0].Title)
	assert.Equal(t, "高タンパク低脂質の定番メニュー", suggestions[0].Description)
	assert.Equal(t, "タンパク質不足を補うために提案", suggestions[0].Reason)
	assert.Len(t, suggestions[0].Ingredients, 2)
	assert.Equal(t, float64(350), suggestions[0].EstimatedNutrition.Calories)
	assert.Equal(t, float64(40), suggestions[0].EstimatedNutrition.Protein)
}

func TestSuggestMenus_APIError(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return nil, fmt.Errorf("API呼び出し失敗")
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)
	input := MenuSuggestionInput{MealType: "lunch", Count: 1}

	_, err := suggester.SuggestMenus(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini APIコールエラー")
}

func TestSuggestMenus_InvalidJSON(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: "not valid json"}, nil
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)
	input := MenuSuggestionInput{MealType: "dinner", Count: 2}

	_, err := suggester.SuggestMenus(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "パースエラー")
}

func TestGenerateRecipe_Success(t *testing.T) {
	recipeJSON := `{"recipe": "1. 鶏肉を一口大に切る\n2. フライパンで炒める\n3. 塩コショウで味付け"}`

	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: recipeJSON}, nil
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)
	ingredients := []GeminiIngredient{
		{Name: "鶏むね肉", Quantity: 200, Unit: "g"},
	}

	recipe, err := suggester.GenerateRecipe(context.Background(), "鶏むね肉のソテー", ingredients)

	require.NoError(t, err)
	assert.Contains(t, recipe, "鶏肉を一口大に切る")
}

func TestGenerateRecipe_APIError(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return nil, fmt.Errorf("タイムアウト")
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)

	_, err := suggester.GenerateRecipe(context.Background(), "テスト料理", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "APIコールエラー")
}

func TestGenerateRecipe_InvalidJSON(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `{"invalid_field": "value"}`}, nil
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)

	// recipe フィールドなしでもエラーにならずに空文字を返すことを確認
	recipe, err := suggester.GenerateRecipe(context.Background(), "テスト料理", nil)

	require.NoError(t, err)
	assert.Equal(t, "", recipe)
}

func TestBuildSuggestionPrompt_ContainsMealType(t *testing.T) {
	tests := []struct {
		mealType string
		label    string
	}{
		{"breakfast", "朝食"},
		{"lunch", "昼食"},
		{"dinner", "夕食"},
		{"snack", "間食"},
	}

	for _, tt := range tests {
		t.Run(tt.mealType, func(t *testing.T) {
			input := MenuSuggestionInput{
				MealType: tt.mealType,
				Count:    3,
			}
			prompt := buildSuggestionPrompt(input)
			assert.Contains(t, prompt, tt.label)
		})
	}
}

func TestBuildSuggestionPrompt_ContainsIngredients(t *testing.T) {
	input := MenuSuggestionInput{
		MealType: "lunch",
		Count:    2,
		Ingredients: []IngredientContext{
			{Name: "豚肉", Quantity: 200, Unit: "g"},
			{Name: "キャベツ", Quantity: 150, Unit: "g"},
		},
	}

	prompt := buildSuggestionPrompt(input)

	assert.Contains(t, prompt, "豚肉")
	assert.Contains(t, prompt, "キャベツ")
}

func TestBuildSuggestionPrompt_NoIngredients(t *testing.T) {
	input := MenuSuggestionInput{
		MealType:    "dinner",
		Count:       1,
		Ingredients: []IngredientContext{},
	}

	prompt := buildSuggestionPrompt(input)

	assert.Contains(t, prompt, "食材なし")
}

func TestBuildSuggestionPrompt_WithNutritionGoal(t *testing.T) {
	input := MenuSuggestionInput{
		MealType: "lunch",
		Count:    2,
		NutritionGoal: &NutritionGoalContext{
			TargetCalories:      2500,
			TargetProtein:       180,
			TargetFat:           70,
			TargetCarbohydrates: 300,
			Phase:               "増量",
		},
	}

	prompt := buildSuggestionPrompt(input)

	assert.Contains(t, prompt, "2500")
	assert.Contains(t, prompt, "増量")
}

func TestBuildRecipePrompt_ContainsTitle(t *testing.T) {
	ingredients := []GeminiIngredient{
		{Name: "鶏むね肉", Quantity: 200, Unit: "g"},
	}

	prompt := buildRecipePrompt("鶏むね肉のソテー", ingredients)

	assert.Contains(t, prompt, "鶏むね肉のソテー")
	assert.Contains(t, prompt, "鶏むね肉")
	assert.Contains(t, prompt, "200")
}

func TestSuggestMenus_WithCodeBlock(t *testing.T) {
	// Gemini がコードブロックで囲んで返す場合
	responseJSON := "```json\n{\"suggestions\": []}\n```"

	mockClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: responseJSON}, nil
		},
	}
	suggester := NewMenuSuggesterWithHTTPClient(mockClient)
	input := MenuSuggestionInput{MealType: "lunch", Count: 1}

	suggestions, err := suggester.SuggestMenus(context.Background(), input)

	require.NoError(t, err)
	assert.Empty(t, suggestions)
}
