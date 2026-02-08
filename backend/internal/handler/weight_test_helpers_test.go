package handler

import (
	"context"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

const testRecordUUID = "550e8400-e29b-41d4-a716-446655440000"

// MockWeightRecordRepository はWeightRecordRepository用テストモック
type MockWeightRecordRepository struct {
	CreateRecordFunc           func(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error)
	GetRecordFunc              func(ctx context.Context, userID string, recordID string) (*repository.WeightRecord, error)
	UpdateRecordFunc           func(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error)
	DeleteRecordFunc           func(ctx context.Context, userID string, recordID string) error
	ListRecordsByDateRangeFunc func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error)
}

func (m *MockWeightRecordRepository) CreateRecord(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error) {
	if m.CreateRecordFunc != nil {
		return m.CreateRecordFunc(ctx, userID, weightKg, recordedAt, note)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) GetRecord(ctx context.Context, userID string, recordID string) (*repository.WeightRecord, error) {
	if m.GetRecordFunc != nil {
		return m.GetRecordFunc(ctx, userID, recordID)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) UpdateRecord(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error) {
	if m.UpdateRecordFunc != nil {
		return m.UpdateRecordFunc(ctx, userID, recordID, weightKg, note)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) DeleteRecord(ctx context.Context, userID string, recordID string) error {
	if m.DeleteRecordFunc != nil {
		return m.DeleteRecordFunc(ctx, userID, recordID)
	}
	return nil
}

func (m *MockWeightRecordRepository) ListRecordsByDateRange(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
	if m.ListRecordsByDateRangeFunc != nil {
		return m.ListRecordsByDateRangeFunc(ctx, userID, from, to)
	}
	return nil, nil
}

// MockWeightGoalRepository はWeightGoalRepository用テストモック
type MockWeightGoalRepository struct {
	GetGoalFunc func(ctx context.Context, userID string) (*repository.WeightGoal, error)
	SetGoalFunc func(ctx context.Context, userID string, targetWeightKg float64) (*repository.WeightGoal, error)
}

func (m *MockWeightGoalRepository) GetGoal(ctx context.Context, userID string) (*repository.WeightGoal, error) {
	if m.GetGoalFunc != nil {
		return m.GetGoalFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockWeightGoalRepository) SetGoal(ctx context.Context, userID string, targetWeightKg float64) (*repository.WeightGoal, error) {
	if m.SetGoalFunc != nil {
		return m.SetGoalFunc(ctx, userID, targetWeightKg)
	}
	return nil, nil
}
