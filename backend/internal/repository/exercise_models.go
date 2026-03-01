package repository

import (
	"context"
	"fmt"
	"time"
)

const RecordedDateLayout = "2006-01-02"

// EstimationMethod は消費カロリーの推定方法を表す型
type EstimationMethod string

const (
	EstimationMethodMET    EstimationMethod = "met"
	EstimationMethodGemini EstimationMethod = "gemini"
)

// ExerciseRecord は運動記録を表す構造体
type ExerciseRecord struct {
	ID                 string           `json:"id"`
	ExerciseName       string           `json:"exercise_name"`
	DurationMinutes    int              `json:"duration_minutes"`
	BurnedCaloriesKcal float64          `json:"burned_calories_kcal"`
	EstimationMethod   EstimationMethod `json:"estimation_method"`
	RecordedDate       string           `json:"recorded_date"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// CreateExerciseInput は運動記録作成の入力
type CreateExerciseInput struct {
	ExerciseName       string
	DurationMinutes    int
	BurnedCaloriesKcal float64
	EstimationMethod   EstimationMethod
	RecordedDate       string
}

// ExerciseDailyResult は日次運動記録の取得結果
type ExerciseDailyResult struct {
	Records                 []ExerciseRecord
	TotalBurnedCaloriesKcal float64
}

// ExerciseRepository は運動記録の永続化を担当するインターフェース
type ExerciseRepository interface {
	// Create は新しい運動記録を作成します
	Create(ctx context.Context, userID string, input CreateExerciseInput) (*ExerciseRecord, error)

	// ListByDate は指定日の運動記録一覧を取得します（createdAt昇順）
	ListByDate(ctx context.Context, userID string, recordedDate string) ([]ExerciseRecord, error)

	// Delete は運動記録を削除します（userIDスコープ検証）
	Delete(ctx context.Context, userID string, recordID string) error
}

// ValidateExerciseName は種目名のバリデーションを行います
func ValidateExerciseName(name string) error {
	runes := []rune(name)
	if len(runes) == 0 {
		return fmt.Errorf("exercise_nameは必須です")
	}
	if len(runes) > 100 {
		return fmt.Errorf("exercise_nameは100文字以内で指定してください")
	}
	return nil
}

// ValidateDurationMinutes は実施時間のバリデーションを行います
func ValidateDurationMinutes(minutes int) error {
	if minutes < 5 || minutes > 600 {
		return fmt.Errorf("duration_minutesは5〜600の範囲で指定してください")
	}
	return nil
}

// ValidateRecordedDate は日付文字列のフォーマットを検証します（YYYY-MM-DD）
func ValidateRecordedDate(date string) error {
	if _, err := time.Parse(RecordedDateLayout, date); err != nil {
		return fmt.Errorf("recorded_dateはYYYY-MM-DD形式で指定してください")
	}
	return nil
}
