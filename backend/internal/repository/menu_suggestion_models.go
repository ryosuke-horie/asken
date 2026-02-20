package repository

import (
	"context"
	"time"
)

// MenuSuggestionStatus はメニューサジェストのステータスを表す型
type MenuSuggestionStatus string

const (
	MenuStatusSuggested MenuSuggestionStatus = "suggested"
	MenuStatusAccepted  MenuSuggestionStatus = "accepted"
	MenuStatusDismissed MenuSuggestionStatus = "dismissed"
)

// MenuSuggestionIngredient はサジェストで使用する食材を表す構造体
type MenuSuggestionIngredient struct {
	IngredientID string  `json:"ingredientId"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
}

// EstimatedNutrition は推定栄養素を表す構造体
type EstimatedNutrition struct {
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Fat           float64 `json:"fat"`
	Carbohydrates float64 `json:"carbohydrates"`
}

// MenuSuggestion はメニューサジェストを表す構造体
type MenuSuggestion struct {
	ID                 string                     `json:"id"`
	Title              string                     `json:"title"`
	Description        string                     `json:"description"`
	Reason             string                     `json:"reason"`
	IngredientsUsed    []MenuSuggestionIngredient `json:"ingredientsUsed"`
	Recipe             string                     `json:"recipe,omitempty"`
	EstimatedNutrition EstimatedNutrition         `json:"estimatedNutrition"`
	MealType           string                     `json:"mealType"`
	Status             string                     `json:"status"`
	CreatedAt          time.Time                  `json:"createdAt"`
	UpdatedAt          time.Time                  `json:"updatedAt"`
}

// CreateMenuSuggestionInput はメニューサジェスト作成の入力
type CreateMenuSuggestionInput struct {
	Title              string
	Description        string
	Reason             string
	IngredientsUsed    []MenuSuggestionIngredient
	EstimatedNutrition EstimatedNutrition
	MealType           string
}

// DeductedIngredient は控除された食材情報を表す構造体
type DeductedIngredient struct {
	IngredientID string  `json:"ingredientId"`
	Name         string  `json:"name"`
	Deducted     float64 `json:"deducted"`
	Remaining    float64 `json:"remaining"`
}

// AcceptMenuSuggestionResult はサジェスト採用の結果
type AcceptMenuSuggestionResult struct {
	AnalysisRequestID   string               `json:"analysisRequestId"`
	DeductedIngredients []DeductedIngredient `json:"deductedIngredients"`
}

// MenuSuggestionRepository はメニューサジェストの永続化を担当するインターフェース
type MenuSuggestionRepository interface {
	// Create は新しいメニューサジェストを作成します
	Create(ctx context.Context, userID string, input CreateMenuSuggestionInput) (*MenuSuggestion, error)

	// List はユーザーのメニューサジェスト一覧を取得します。statusが空の場合は全件取得します。
	List(ctx context.Context, userID string, status string, limit int) ([]MenuSuggestion, error)

	// GetByID は指定されたIDのメニューサジェストを取得します
	GetByID(ctx context.Context, userID string, id string) (*MenuSuggestion, error)

	// UpdateRecipe はレシピを更新します（遅延生成）
	UpdateRecipe(ctx context.Context, userID string, id string, recipe string) error

	// Accept はサジェストを採用し、食事記録と食材控除をトランザクションで実行します
	Accept(ctx context.Context, userID string, id string) (*AcceptMenuSuggestionResult, error)

	// Dismiss はサジェストを却下します
	Dismiss(ctx context.Context, userID string, id string) error
}
