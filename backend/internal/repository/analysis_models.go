package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// ErrNotFound はリソースが見つからない場合のエラー
var ErrNotFound = errors.New("resource not found")

// AnalysisStatus は分析リクエストのステータスを表す型
type AnalysisStatus string

const (
	StatusPending    AnalysisStatus = "pending"
	StatusProcessing AnalysisStatus = "processing"
	StatusCompleted  AnalysisStatus = "completed"
	StatusFailed     AnalysisStatus = "failed"
)

// InputType は入力タイプを表す型
type InputType string

const (
	InputTypeImage      InputType = "image"
	InputTypeText       InputType = "text"
	InputTypeMylist     InputType = "mylist"
	InputTypeSkipped    InputType = "skipped"
	InputTypeSuggestion InputType = "suggestion"
)

// AnalysisRequest は分析リクエストを表す構造体
type AnalysisRequest struct {
	ID           uuid.UUID      `json:"id"`
	Status       AnalysisStatus `json:"status"`
	InputType    InputType      `json:"input_type"`
	ImagePath    string         `json:"image_path"`
	InputText    string         `json:"input_text"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// HistoryItem は履歴一覧の各項目を表す構造体
type HistoryItem struct {
	ID                  uuid.UUID          `json:"id"`
	InputType           InputType          `json:"input_type"`
	ImagePath           string             `json:"image_path"`
	InputText           string             `json:"input_text"`
	CreatedAt           time.Time          `json:"created_at"`
	MealType            string             `json:"meal_type"`
	MealDate            time.Time          `json:"meal_date"`
	TotalCalories       float64            `json:"total_calories"`
	TotalProtein        float64            `json:"total_protein"`
	TotalFat            float64            `json:"total_fat"`
	TotalCarbohydrates  float64            `json:"total_carbohydrates"`
	TotalMicronutrients map[string]float64 `json:"total_micronutrients,omitempty"`
}

// HistoryDetail は履歴詳細を表す構造体
type HistoryDetail struct {
	HistoryItem
	Foods []gemini.NutritionInfo `json:"foods"`
}

// DailyTotal は1日の合計栄養素を表す構造体
type DailyTotal struct {
	TotalCalories       float64            `json:"total_calories"`
	TotalProtein        float64            `json:"total_protein"`
	TotalFat            float64            `json:"total_fat"`
	TotalCarbohydrates  float64            `json:"total_carbohydrates"`
	TotalMicronutrients map[string]float64 `json:"total_micronutrients,omitempty"`
}

// AnalysisRepository は分析リクエストと結果の永続化を担当するインターフェース
type AnalysisRepository interface {
	// CreateRequest は新しい画像分析リクエストを作成します
	CreateRequest(ctx context.Context, imagePath string, mealType string, mealDate string, userID *string) (uuid.UUID, error)

	// CreateRequestWithText は新しいテキスト分析リクエストを作成します
	CreateRequestWithText(ctx context.Context, inputText string, mealType string, mealDate string, userID *string) (uuid.UUID, error)

	// GetRequest は指定されたIDの分析リクエストを取得します（userIDでスコープ）
	GetRequest(ctx context.Context, userID string, id uuid.UUID) (*AnalysisRequest, error)

	// UpdateStatus はリクエストのステータスを更新します（ワーカー用: 全ユーザー横断）
	UpdateStatus(ctx context.Context, id uuid.UUID, status AnalysisStatus, errorMessage string) error

	// SaveResult は分析結果を保存し、ステータスをcompletedに更新します（ワーカー用: 全ユーザー横断）
	SaveResult(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error

	// GetResult は指定されたリクエストIDの分析結果を取得します（userIDでスコープ）
	GetResult(ctx context.Context, userID string, requestID uuid.UUID) (*service.AnalysisResult, error)

	// GetPendingRequests はpending状態のリクエストを取得します（ワーカー用: 全ユーザー横断）
	GetPendingRequests(ctx context.Context, limit int) ([]AnalysisRequest, error)

	// GetHistoryList は履歴一覧を取得します（userIDでスコープ、ページネーション対応）
	GetHistoryList(ctx context.Context, userID string, page, limit int) ([]HistoryItem, int, error)

	// GetHistoryDetail は履歴詳細を取得します（userIDでスコープ）
	GetHistoryDetail(ctx context.Context, userID string, id uuid.UUID) (*HistoryDetail, error)

	// DeleteHistory は履歴を削除します（userIDでスコープ、関連する画像も含む）
	DeleteHistory(ctx context.Context, userID string, id uuid.UUID) error

	// GetDailyMeals は指定された日付の食事データを取得します（userIDでスコープ）
	// tz: IANAタイムゾーン名（例: "Asia/Tokyo"）。空文字の場合はUTCとして処理
	GetDailyMeals(ctx context.Context, userID string, date string, tz string) (map[string][]HistoryDetail, DailyTotal, error)

	// CreateRequestFromMylist はマイリストからの食事記録を作成します
	CreateRequestFromMylist(ctx context.Context, inputText string, mealType string, mealDate string, userID *string, result *service.AnalysisResult) (uuid.UUID, error)

	// CreateSkippedMeal は「食べなかった」記録を作成します
	CreateSkippedMeal(ctx context.Context, mealType string, mealDate string, userID *string) (uuid.UUID, error)

	// UpdateResult は分析結果を更新します（userIDでスコープ、foods配列と合計値を再計算）
	UpdateResult(ctx context.Context, userID string, historyID uuid.UUID, foods []gemini.NutritionInfo) error
}
