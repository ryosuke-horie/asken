package repository

import (
	"context"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// MyMenuRepository はマイメニューの管理を担当するインターフェース
type MyMenuRepository interface {
	// Create は新しいマイメニューを作成します
	Create(ctx context.Context, userID string, name string, foods []gemini.NutritionInfo) (*MyMenuItem, error)

	// List はユーザーのマイメニュー一覧を取得します
	List(ctx context.Context, userID string) ([]MyMenuItem, error)

	// Get は指定されたIDのマイメニューを取得します
	Get(ctx context.Context, userID string, menuID string) (*MyMenuItem, error)

	// Update はマイメニューを更新します
	Update(ctx context.Context, userID string, menuID string, name string, foods []gemini.NutritionInfo) (*MyMenuItem, error)

	// Delete はマイメニューを削除します
	Delete(ctx context.Context, userID string, menuID string) error
}

// MyMenuItem はマイメニュー項目を表す構造体
type MyMenuItem struct {
	ID                 string
	Name               string
	Foods              []gemini.NutritionInfo
	TotalCalories      float64
	TotalProtein       float64
	TotalFat           float64
	TotalCarbohydrates float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
