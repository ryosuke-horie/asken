package repository

import (
	"context"
	"time"
)

// WeightRecord は体重記録を表す構造体
type WeightRecord struct {
	ID         string    `json:"id"`
	WeightKg   float64   `json:"weight_kg"`
	RecordedAt time.Time `json:"recorded_at"`
	Note       string    `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WeightGoal は目標体重を表す構造体
type WeightGoal struct {
	TargetWeightKg float64   `json:"target_weight_kg"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WeightRecordRepository は体重記録の永続化を担当するインターフェース
type WeightRecordRepository interface {
	// CreateRecord は新しい体重記録を作成します
	CreateRecord(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*WeightRecord, error)

	// GetRecord は指定されたIDの体重記録を取得します
	GetRecord(ctx context.Context, userID string, recordID string) (*WeightRecord, error)

	// UpdateRecord は体重記録を更新します
	UpdateRecord(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*WeightRecord, error)

	// DeleteRecord は体重記録を削除します
	DeleteRecord(ctx context.Context, userID string, recordID string) error

	// ListRecordsByDateRange は期間指定で体重記録を取得します
	ListRecordsByDateRange(ctx context.Context, userID string, from time.Time, to time.Time) ([]WeightRecord, error)

	// GetGoal は目標体重を取得します
	GetGoal(ctx context.Context, userID string) (*WeightGoal, error)

	// SetGoal は目標体重を設定・更新します
	SetGoal(ctx context.Context, userID string, targetWeightKg float64) (*WeightGoal, error)
}
