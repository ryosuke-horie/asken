package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExerciseRepo はテスト用のExerciseRepositoryモック
type MockExerciseRepo struct {
	CreateFunc     func(ctx context.Context, userID string, input repository.CreateExerciseInput) (*repository.ExerciseRecord, error)
	ListByDateFunc func(ctx context.Context, userID string, recordedDate string) ([]repository.ExerciseRecord, error)
	DeleteFunc     func(ctx context.Context, userID string, recordID string) error
}

func (m *MockExerciseRepo) Create(ctx context.Context, userID string, input repository.CreateExerciseInput) (*repository.ExerciseRecord, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, input)
	}
	return nil, nil
}

func (m *MockExerciseRepo) ListByDate(ctx context.Context, userID string, recordedDate string) ([]repository.ExerciseRecord, error) {
	if m.ListByDateFunc != nil {
		return m.ListByDateFunc(ctx, userID, recordedDate)
	}
	return []repository.ExerciseRecord{}, nil
}

func (m *MockExerciseRepo) Delete(ctx context.Context, userID string, recordID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, userID, recordID)
	}
	return nil
}

// MockEstimator はテスト用のExerciseCalorieEstimatorモック
type MockEstimator struct {
	EstimateFunc func(ctx context.Context, exerciseName string, durationMinutes int) (float64, error)
}

func (m *MockEstimator) EstimateCalories(ctx context.Context, exerciseName string, durationMinutes int) (float64, error) {
	if m.EstimateFunc != nil {
		return m.EstimateFunc(ctx, exerciseName, durationMinutes)
	}
	return 0, nil
}

func newTestExerciseService(repo repository.ExerciseRepository, estimator ExerciseCalorieEstimator) *ExerciseService {
	return NewExerciseService(repo, estimator)
}

func TestCalculateByMET(t *testing.T) {
	tests := []struct {
		name     string
		met      float64
		minutes  int
		expected float64
	}{
		{"柔術90分", 10.0, 90, 1102.5},     // 10 * 70 * 1.5 * 1.05
		{"ランニング60分", 9.8, 60, 720.3},    // 9.8 * 70 * 1.0 * 1.05
		{"ウォーキング30分", 3.5, 30, 128.625}, // 3.5 * 70 * 0.5 * 1.05
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateByMET(tt.met, tt.minutes)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestCreateExerciseRecord_PresetMET(t *testing.T) {
	ctx := context.Background()
	var capturedInput repository.CreateExerciseInput

	repo := &MockExerciseRepo{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateExerciseInput) (*repository.ExerciseRecord, error) {
			capturedInput = input
			return &repository.ExerciseRecord{
				ID:                 "test-id",
				ExerciseName:       input.ExerciseName,
				DurationMinutes:    input.DurationMinutes,
				BurnedCaloriesKcal: input.BurnedCaloriesKcal,
				EstimationMethod:   input.EstimationMethod,
				RecordedDate:       input.RecordedDate,
			}, nil
		},
	}
	estimator := &MockEstimator{}
	svc := newTestExerciseService(repo, estimator)

	record, err := svc.CreateExerciseRecord(ctx, "test-user", CreateExerciseInput{
		ExerciseName:    "柔術",
		DurationMinutes: 90,
	}, "2026-02-28")

	require.NoError(t, err)
	assert.Equal(t, "柔術", record.ExerciseName)
	assert.Equal(t, repository.EstimationMethodMET, capturedInput.EstimationMethod)
	assert.Greater(t, capturedInput.BurnedCaloriesKcal, 0.0)
}

func TestCreateExerciseRecord_GeminiEstimation(t *testing.T) {
	ctx := context.Background()
	var capturedInput repository.CreateExerciseInput

	repo := &MockExerciseRepo{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateExerciseInput) (*repository.ExerciseRecord, error) {
			capturedInput = input
			return &repository.ExerciseRecord{
				ID:                 "test-id",
				ExerciseName:       input.ExerciseName,
				DurationMinutes:    input.DurationMinutes,
				BurnedCaloriesKcal: input.BurnedCaloriesKcal,
				EstimationMethod:   input.EstimationMethod,
			}, nil
		},
	}
	estimator := &MockEstimator{
		EstimateFunc: func(ctx context.Context, exerciseName string, durationMinutes int) (float64, error) {
			return 300.0, nil
		},
	}
	svc := newTestExerciseService(repo, estimator)

	record, err := svc.CreateExerciseRecord(ctx, "test-user", CreateExerciseInput{
		ExerciseName:    "フリークライミング",
		DurationMinutes: 60,
	}, "2026-02-28")

	require.NoError(t, err)
	assert.Equal(t, repository.EstimationMethodGemini, capturedInput.EstimationMethod)
	assert.Equal(t, 300.0, record.BurnedCaloriesKcal)
}

func TestCreateExerciseRecord_ValidationError(t *testing.T) {
	ctx := context.Background()
	repo := &MockExerciseRepo{}
	estimator := &MockEstimator{}
	svc := newTestExerciseService(repo, estimator)

	t.Run("空の種目名", func(t *testing.T) {
		_, err := svc.CreateExerciseRecord(ctx, "test-user", CreateExerciseInput{
			ExerciseName:    "",
			DurationMinutes: 60,
		}, "2026-02-28")
		assert.Error(t, err)
	})

	t.Run("時間が短すぎる", func(t *testing.T) {
		_, err := svc.CreateExerciseRecord(ctx, "test-user", CreateExerciseInput{
			ExerciseName:    "柔術",
			DurationMinutes: 4,
		}, "2026-02-28")
		assert.Error(t, err)
	})
}

func TestCreateExerciseRecord_GeminiError(t *testing.T) {
	ctx := context.Background()
	repo := &MockExerciseRepo{}
	estimator := &MockEstimator{
		EstimateFunc: func(ctx context.Context, exerciseName string, durationMinutes int) (float64, error) {
			return 0, errors.New("Gemini API error")
		},
	}
	svc := newTestExerciseService(repo, estimator)

	_, err := svc.CreateExerciseRecord(ctx, "test-user", CreateExerciseInput{
		ExerciseName:    "フリークライミング",
		DurationMinutes: 60,
	}, "2026-02-28")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "消費カロリーの推定に失敗")
}

func TestGetDailyExercise(t *testing.T) {
	ctx := context.Background()

	repo := &MockExerciseRepo{
		ListByDateFunc: func(ctx context.Context, userID string, recordedDate string) ([]repository.ExerciseRecord, error) {
			return []repository.ExerciseRecord{
				{ID: "1", ExerciseName: "柔術", BurnedCaloriesKcal: 486.0},
				{ID: "2", ExerciseName: "ランニング", BurnedCaloriesKcal: 324.0},
			}, nil
		},
	}
	svc := newTestExerciseService(repo, &MockEstimator{})

	result, err := svc.GetDailyExercise(ctx, "test-user", "2026-02-28")

	require.NoError(t, err)
	assert.Len(t, result.Records, 2)
	assert.InDelta(t, 810.0, result.TotalBurnedCaloriesKcal, 0.001)
}

func TestGetDailyExercise_Empty(t *testing.T) {
	ctx := context.Background()
	repo := &MockExerciseRepo{}
	svc := newTestExerciseService(repo, &MockEstimator{})

	result, err := svc.GetDailyExercise(ctx, "test-user", "2099-01-01")

	require.NoError(t, err)
	assert.Empty(t, result.Records)
	assert.Equal(t, 0.0, result.TotalBurnedCaloriesKcal)
}

func TestDeleteExerciseRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("正常削除", func(t *testing.T) {
		repo := &MockExerciseRepo{
			DeleteFunc: func(ctx context.Context, userID string, recordID string) error {
				return nil
			},
		}
		svc := newTestExerciseService(repo, &MockEstimator{})
		err := svc.DeleteExerciseRecord(ctx, "test-user", "record-id")
		assert.NoError(t, err)
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		repo := &MockExerciseRepo{
			DeleteFunc: func(ctx context.Context, userID string, recordID string) error {
				return repository.ErrNotFound
			},
		}
		svc := newTestExerciseService(repo, &MockEstimator{})
		err := svc.DeleteExerciseRecord(ctx, "test-user", "non-existent")
		assert.Error(t, err)
	})
}
