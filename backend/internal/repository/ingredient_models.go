package repository

import (
	"context"
	"time"
)

// IngredientCategory は食材カテゴリを表す型
type IngredientCategory string

const (
	CategoryMeat      IngredientCategory = "meat"
	CategoryFish      IngredientCategory = "fish"
	CategoryVegetable IngredientCategory = "vegetable"
	CategoryFruit     IngredientCategory = "fruit"
	CategoryDairy     IngredientCategory = "dairy"
	CategoryGrain     IngredientCategory = "grain"
	CategorySeasoning IngredientCategory = "seasoning"
	CategoryBeverage  IngredientCategory = "beverage"
	CategoryOther     IngredientCategory = "other"
)

// IngredientSource は食材の入力元を表す型
type IngredientSource string

const (
	SourceReceipt IngredientSource = "receipt"
	SourceManual  IngredientSource = "manual"
)

// Ingredient は食材を表す構造体
type Ingredient struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Category     string     `json:"category"`
	Quantity     float64    `json:"quantity"`
	Unit         string     `json:"unit"`
	PurchaseDate *time.Time `json:"purchaseDate,omitempty"`
	ExpiryDate   *time.Time `json:"expiryDate,omitempty"`
	Source       string     `json:"source"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// CreateIngredientInput は食材作成の入力
type CreateIngredientInput struct {
	Name         string
	Category     string
	Quantity     float64
	Unit         string
	PurchaseDate *time.Time
	ExpiryDate   *time.Time
	Source       string
}

// UpdateIngredientInput は食材更新の入力
type UpdateIngredientInput struct {
	Name         string
	Category     string
	Quantity     float64
	Unit         string
	PurchaseDate *time.Time
	ExpiryDate   *time.Time
}

// IngredientRepository は食材の永続化を担当するインターフェース
type IngredientRepository interface {
	// Create は新しい食材を作成します
	Create(ctx context.Context, userID string, input CreateIngredientInput) (*Ingredient, error)

	// List はユーザーの食材一覧を取得します。categoryが空の場合は全件取得します。
	List(ctx context.Context, userID string, category string) ([]Ingredient, error)

	// GetByID は指定されたIDの食材を取得します
	GetByID(ctx context.Context, userID string, ingredientID string) (*Ingredient, error)

	// Update は食材を更新します
	Update(ctx context.Context, userID string, ingredientID string, input UpdateIngredientInput) (*Ingredient, error)

	// Delete は食材を削除します
	Delete(ctx context.Context, userID string, ingredientID string) error
}
