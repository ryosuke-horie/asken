package service

import (
	"context"
	"fmt"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// ExerciseCalorieEstimator はGeminiを使った消費カロリー推定のインターフェース
type ExerciseCalorieEstimator interface {
	EstimateCalories(ctx context.Context, exerciseName string, durationMinutes int) (float64, error)
}

// ValidationError はバリデーションエラーを表す型
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// defaultBodyWeightKg はMET計算に使用するデフォルト体重（ユーザー体重未設定時）
const defaultBodyWeightKg = 70.0

// metValues はプリセット種目のMET値テーブル（体重70kg想定の計算に使用）
// MET値出典: Ainsworth et al., "Compendium of Physical Activities" 2011年版
var metValues = map[string]float64{
	"柔術":       10.0,
	"グラップリング":  10.0,
	"レスリング":    7.0,
	"キックボクシング": 10.3,
	"ボクシング":    12.8,
	"空手":       10.0,
	"柔道":       10.0,
	"MMA":      10.0,
	"ランニング":    9.8,
	"自転車":      4.5,
	"サイクリング":   8.0,
	"水泳":       8.3,
	"縄跳び":      11.8,
	"ウォーキング":   3.5,
	"筋力トレーニング": 5.0,
	"筋トレ":      5.0,
	"体幹トレーニング": 3.5,
}

// calculateByMET はMET値から消費カロリーを計算する
// 計算式: MET × 体重(kg) × 時間(h) × 1.05
func calculateByMET(met float64, durationMinutes int) float64 {
	hours := float64(durationMinutes) / 60.0
	return met * defaultBodyWeightKg * hours * 1.05
}

// CreateExerciseInput は運動記録作成の入力（サービス層）
type CreateExerciseInput struct {
	ExerciseName    string
	DurationMinutes int
}

// ExerciseService は運動記録サービス
type ExerciseService struct {
	repo      repository.ExerciseRepository
	estimator ExerciseCalorieEstimator
}

// NewExerciseService は新しいExerciseServiceを作成
func NewExerciseService(repo repository.ExerciseRepository, estimator ExerciseCalorieEstimator) *ExerciseService {
	if repo == nil {
		panic("exercise service: repository must not be nil")
	}
	if estimator == nil {
		panic("exercise service: estimator must not be nil")
	}
	return &ExerciseService{
		repo:      repo,
		estimator: estimator,
	}
}

// CreateExerciseRecord は運動記録を作成する
// プリセット種目はMET値で計算し、プリセット外はGeminiで推定する
func (s *ExerciseService) CreateExerciseRecord(ctx context.Context, userID string, input CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error) {
	if err := repository.ValidateExerciseName(input.ExerciseName); err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}
	if err := repository.ValidateDurationMinutes(input.DurationMinutes); err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}
	if err := repository.ValidateRecordedDate(recordedDate); err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}

	var burnedCaloriesKcal float64
	var estimationMethod repository.EstimationMethod

	if met, ok := metValues[input.ExerciseName]; ok {
		burnedCaloriesKcal = calculateByMET(met, input.DurationMinutes)
		estimationMethod = repository.EstimationMethodMET
	} else {
		kcal, err := s.estimator.EstimateCalories(ctx, input.ExerciseName, input.DurationMinutes)
		if err != nil {
			return nil, fmt.Errorf("消費カロリーの推定に失敗: %w", err)
		}
		burnedCaloriesKcal = kcal
		estimationMethod = repository.EstimationMethodGemini
	}

	repoInput := repository.CreateExerciseInput{
		ExerciseName:       input.ExerciseName,
		DurationMinutes:    input.DurationMinutes,
		BurnedCaloriesKcal: burnedCaloriesKcal,
		EstimationMethod:   estimationMethod,
		RecordedDate:       recordedDate,
	}

	return s.repo.Create(ctx, userID, repoInput)
}

// GetDailyExercise は指定日の運動記録と合計消費カロリーを取得する
func (s *ExerciseService) GetDailyExercise(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error) {
	records, err := s.repo.ListByDate(ctx, userID, recordedDate)
	if err != nil {
		return nil, fmt.Errorf("運動記録の取得に失敗: %w", err)
	}

	if records == nil {
		records = []repository.ExerciseRecord{}
	}

	var total float64
	for _, r := range records {
		total += r.BurnedCaloriesKcal
	}

	return &repository.ExerciseDailyResult{
		Records:                 records,
		TotalBurnedCaloriesKcal: total,
	}, nil
}

// DeleteExerciseRecord は運動記録を削除する
func (s *ExerciseService) DeleteExerciseRecord(ctx context.Context, userID string, recordID string) error {
	return s.repo.Delete(ctx, userID, recordID)
}
