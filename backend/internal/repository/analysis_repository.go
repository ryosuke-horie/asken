package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

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
	InputTypeImage InputType = "image"
	InputTypeText  InputType = "text"
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
	ID                 uuid.UUID `json:"id"`
	InputType          InputType `json:"input_type"`
	ImagePath          string    `json:"image_path"`
	InputText          string    `json:"input_text"`
	CreatedAt          time.Time `json:"created_at"`
	MealType           string    `json:"meal_type"`
	MealDate           time.Time `json:"meal_date"`
	TotalCalories      float64   `json:"total_calories"`
	TotalProtein       float64   `json:"total_protein"`
	TotalFat           float64   `json:"total_fat"`
	TotalCarbohydrates float64   `json:"total_carbohydrates"`
}

// HistoryDetail は履歴詳細を表す構造体
type HistoryDetail struct {
	HistoryItem
	Foods []gemini.NutritionInfo `json:"foods"`
}

// DailyTotal は1日の合計栄養素を表す構造体
type DailyTotal struct {
	TotalCalories      float64 `json:"total_calories"`
	TotalProtein       float64 `json:"total_protein"`
	TotalFat           float64 `json:"total_fat"`
	TotalCarbohydrates float64 `json:"total_carbohydrates"`
}

// AnalysisRepository は分析リクエストと結果の永続化を担当するインターフェース
type AnalysisRepository interface {
	// CreateRequest は新しい画像分析リクエストを作成します
	CreateRequest(ctx context.Context, imagePath string, mealType string, mealDate string, userID *uuid.UUID) (uuid.UUID, error)

	// CreateRequestWithText は新しいテキスト分析リクエストを作成します
	CreateRequestWithText(ctx context.Context, inputText string, mealType string, mealDate string, userID *uuid.UUID) (uuid.UUID, error)

	// GetRequest は指定されたIDの分析リクエストを取得します
	GetRequest(ctx context.Context, id uuid.UUID) (*AnalysisRequest, error)

	// UpdateStatus はリクエストのステータスを更新します
	UpdateStatus(ctx context.Context, id uuid.UUID, status AnalysisStatus, errorMessage string) error

	// SaveResult は分析結果を保存し、ステータスをcompletedに更新します（トランザクション）
	SaveResult(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error

	// GetResult は指定されたリクエストIDの分析結果を取得します
	GetResult(ctx context.Context, requestID uuid.UUID) (*service.AnalysisResult, error)

	// GetPendingRequests はpending状態のリクエストを取得します（limit: 取得件数上限）
	GetPendingRequests(ctx context.Context, limit int) ([]AnalysisRequest, error)

	// GetHistoryList は履歴一覧を取得します（ページネーション対応）
	GetHistoryList(ctx context.Context, page, limit int) ([]HistoryItem, int, error)

	// GetHistoryDetail は履歴詳細を取得します
	GetHistoryDetail(ctx context.Context, id uuid.UUID) (*HistoryDetail, error)

	// DeleteHistory は履歴を削除します（関連する画像も含む）
	DeleteHistory(ctx context.Context, id uuid.UUID) error

	// GetDailyMeals は指定された日付の食事データを取得します
	GetDailyMeals(ctx context.Context, date string) (map[string][]HistoryDetail, DailyTotal, error)
}

// postgresAnalysisRepository はPostgreSQLを使用したAnalysisRepositoryの実装
type postgresAnalysisRepository struct {
	db *sql.DB
}

// NewAnalysisRepository は新しいAnalysisRepositoryを作成します
func NewAnalysisRepository(db *sql.DB) AnalysisRepository {
	return &postgresAnalysisRepository{db: db}
}

// CreateRequest は新しい画像分析リクエストを作成します
func (r *postgresAnalysisRepository) CreateRequest(ctx context.Context, imagePath string, mealType string, mealDate string, userID *uuid.UUID) (uuid.UUID, error) {
	query := `
		INSERT INTO analysis_requests (status, input_type, image_path, meal_type, meal_date, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, StatusPending, InputTypeImage, imagePath, mealType, mealDate, userID).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("分析リクエストの作成に失敗: %w", err)
	}

	return id, nil
}

// CreateRequestWithText は新しいテキスト分析リクエストを作成します
func (r *postgresAnalysisRepository) CreateRequestWithText(ctx context.Context, inputText string, mealType string, mealDate string, userID *uuid.UUID) (uuid.UUID, error) {
	query := `
		INSERT INTO analysis_requests (status, input_type, input_text, meal_type, meal_date, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, StatusPending, InputTypeText, inputText, mealType, mealDate, userID).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("分析リクエストの作成に失敗: %w", err)
	}

	return id, nil
}

// GetRequest は指定されたIDの分析リクエストを取得します
func (r *postgresAnalysisRepository) GetRequest(ctx context.Context, id uuid.UUID) (*AnalysisRequest, error) {
	query := `
		SELECT id, status, input_type, image_path, input_text, error_message, created_at, updated_at
		FROM analysis_requests
		WHERE id = $1
	`

	var req AnalysisRequest
	var imagePath sql.NullString
	var inputText sql.NullString
	var errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.Status,
		&req.InputType,
		&imagePath,
		&inputText,
		&errorMessage,
		&req.CreatedAt,
		&req.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("リクエストが見つかりません: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("リクエストの取得に失敗: %w", err)
	}

	if imagePath.Valid {
		req.ImagePath = imagePath.String
	}
	if inputText.Valid {
		req.InputText = inputText.String
	}
	if errorMessage.Valid {
		req.ErrorMessage = errorMessage.String
	}

	return &req, nil
}

// UpdateStatus はリクエストのステータスを更新します
func (r *postgresAnalysisRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status AnalysisStatus, errorMessage string) error {
	query := `
		UPDATE analysis_requests
		SET status = $1, error_message = $2
		WHERE id = $3
	`

	var errorMsg sql.NullString
	if errorMessage != "" {
		errorMsg = sql.NullString{String: errorMessage, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, query, status, errorMsg, id)
	if err != nil {
		return fmt.Errorf("ステータスの更新に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("リクエストが見つかりません: %s", id)
	}

	return nil
}

// SaveResult は分析結果を保存し、ステータスをcompletedに更新します
func (r *postgresAnalysisRepository) SaveResult(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error {
	// トランザクション開始
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer tx.Rollback()

	// foods をJSONBに変換
	foodsJSON, err := json.Marshal(result.Foods)
	if err != nil {
		return fmt.Errorf("foods のJSON変換に失敗: %w", err)
	}

	// analysis_results に結果を保存
	insertQuery := `
		INSERT INTO analysis_results (
			analysis_request_id,
			foods,
			total_calories,
			total_protein,
			total_fat,
			total_carbohydrates
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, insertQuery,
		requestID,
		foodsJSON,
		result.TotalCalories,
		result.TotalProtein,
		result.TotalFat,
		result.TotalCarbohydrates,
	)
	if err != nil {
		return fmt.Errorf("分析結果の保存に失敗: %w", err)
	}

	// analysis_requests のステータスをcompletedに更新
	updateQuery := `
		UPDATE analysis_requests
		SET status = $1
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, updateQuery, StatusCompleted, requestID)
	if err != nil {
		return fmt.Errorf("ステータスの更新に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return nil
}

// GetResult は指定されたリクエストIDの分析結果を取得します
func (r *postgresAnalysisRepository) GetResult(ctx context.Context, requestID uuid.UUID) (*service.AnalysisResult, error) {
	query := `
		SELECT foods, total_calories, total_protein, total_fat, total_carbohydrates
		FROM analysis_results
		WHERE analysis_request_id = $1
	`

	var foodsJSON []byte
	var result service.AnalysisResult

	err := r.db.QueryRowContext(ctx, query, requestID).Scan(
		&foodsJSON,
		&result.TotalCalories,
		&result.TotalProtein,
		&result.TotalFat,
		&result.TotalCarbohydrates,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("結果が見つかりません: %s", requestID)
	}
	if err != nil {
		return nil, fmt.Errorf("結果の取得に失敗: %w", err)
	}

	// JSONB から foods をデシリアライズ
	if err := json.Unmarshal(foodsJSON, &result.Foods); err != nil {
		return nil, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
	}

	return &result, nil
}

// GetPendingRequests はpending状態のリクエストを取得します
func (r *postgresAnalysisRepository) GetPendingRequests(ctx context.Context, limit int) ([]AnalysisRequest, error) {
	query := `
		SELECT id, status, input_type, image_path, input_text, error_message, created_at, updated_at
		FROM analysis_requests
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("pending リクエストの取得に失敗: %w", err)
	}
	defer rows.Close()

	var requests []AnalysisRequest
	for rows.Next() {
		var req AnalysisRequest
		var imagePath sql.NullString
		var inputText sql.NullString
		var errorMessage sql.NullString

		err := rows.Scan(
			&req.ID,
			&req.Status,
			&req.InputType,
			&imagePath,
			&inputText,
			&errorMessage,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("行のスキャンに失敗: %w", err)
		}

		if imagePath.Valid {
			req.ImagePath = imagePath.String
		}
		if inputText.Valid {
			req.InputText = inputText.String
		}
		if errorMessage.Valid {
			req.ErrorMessage = errorMessage.String
		}

		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理に失敗: %w", err)
	}

	return requests, nil
}

// GetHistoryList は履歴一覧を取得します（ページネーション対応）
func (r *postgresAnalysisRepository) GetHistoryList(ctx context.Context, page, limit int) ([]HistoryItem, int, error) {
	// バリデーション
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 総件数を取得
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM analysis_requests ar
		INNER JOIN analysis_results res ON ar.id = res.analysis_request_id
		WHERE ar.status = $1
	`
	if err := r.db.QueryRowContext(ctx, countQuery, StatusCompleted).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("総件数の取得に失敗: %w", err)
	}

	// オフセットを計算
	offset := (page - 1) * limit

	// 履歴一覧を取得
	query := `
		SELECT
			ar.id,
			ar.input_type,
			ar.image_path,
			ar.input_text,
			ar.created_at,
			ar.meal_type,
			ar.meal_date,
			res.total_calories,
			res.total_protein,
			res.total_fat,
			res.total_carbohydrates
		FROM analysis_requests ar
		INNER JOIN analysis_results res ON ar.id = res.analysis_request_id
		WHERE ar.status = $1
		ORDER BY ar.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, StatusCompleted, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("履歴一覧の取得に失敗: %w", err)
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var item HistoryItem
		var imagePath sql.NullString
		var inputText sql.NullString
		var mealType sql.NullString
		var mealDate sql.NullTime
		err := rows.Scan(
			&item.ID,
			&item.InputType,
			&imagePath,
			&inputText,
			&item.CreatedAt,
			&mealType,
			&mealDate,
			&item.TotalCalories,
			&item.TotalProtein,
			&item.TotalFat,
			&item.TotalCarbohydrates,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("行のスキャンに失敗: %w", err)
		}
		if imagePath.Valid {
			item.ImagePath = imagePath.String
		}
		if inputText.Valid {
			item.InputText = inputText.String
		}
		if mealType.Valid {
			item.MealType = mealType.String
		}
		if mealDate.Valid {
			item.MealDate = mealDate.Time
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("行の反復処理に失敗: %w", err)
	}

	return items, total, nil
}

// GetHistoryDetail は履歴詳細を取得します
func (r *postgresAnalysisRepository) GetHistoryDetail(ctx context.Context, id uuid.UUID) (*HistoryDetail, error) {
	query := `
		SELECT
			ar.id,
			ar.input_type,
			ar.image_path,
			ar.input_text,
			ar.created_at,
			ar.meal_type,
			ar.meal_date,
			res.foods,
			res.total_calories,
			res.total_protein,
			res.total_fat,
			res.total_carbohydrates
		FROM analysis_requests ar
		INNER JOIN analysis_results res ON ar.id = res.analysis_request_id
		WHERE ar.id = $1 AND ar.status = $2
	`

	var detail HistoryDetail
	var foodsJSON []byte
	var imagePath sql.NullString
	var inputText sql.NullString
	var mealType sql.NullString
	var mealDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, StatusCompleted).Scan(
		&detail.ID,
		&detail.InputType,
		&imagePath,
		&inputText,
		&detail.CreatedAt,
		&mealType,
		&mealDate,
		&foodsJSON,
		&detail.TotalCalories,
		&detail.TotalProtein,
		&detail.TotalFat,
		&detail.TotalCarbohydrates,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("履歴が見つかりません: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("履歴詳細の取得に失敗: %w", err)
	}

	if imagePath.Valid {
		detail.ImagePath = imagePath.String
	}
	if inputText.Valid {
		detail.InputText = inputText.String
	}
	if mealType.Valid {
		detail.MealType = mealType.String
	}
	if mealDate.Valid {
		detail.MealDate = mealDate.Time
	}

	// JSONB から foods をデシリアライズ
	if err := json.Unmarshal(foodsJSON, &detail.Foods); err != nil {
		return nil, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
	}

	return &detail, nil
}

// DeleteHistory は履歴を削除します（関連する画像も含む）
func (r *postgresAnalysisRepository) DeleteHistory(ctx context.Context, id uuid.UUID) error {
	// トランザクション開始
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer tx.Rollback()

	// 入力タイプと画像パスを取得
	var inputType string
	var imagePath sql.NullString
	getInfoQuery := `SELECT input_type, image_path FROM analysis_requests WHERE id = $1`
	if err := tx.QueryRowContext(ctx, getInfoQuery, id).Scan(&inputType, &imagePath); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("履歴が見つかりません: %s", id)
		}
		return fmt.Errorf("リクエスト情報の取得に失敗: %w", err)
	}

	// analysis_results を削除（外部キー制約があるため先に削除）
	deleteResultsQuery := `DELETE FROM analysis_results WHERE analysis_request_id = $1`
	if _, err := tx.ExecContext(ctx, deleteResultsQuery, id); err != nil {
		return fmt.Errorf("分析結果の削除に失敗: %w", err)
	}

	// analysis_requests を削除
	deleteRequestQuery := `DELETE FROM analysis_requests WHERE id = $1`
	result, err := tx.ExecContext(ctx, deleteRequestQuery, id)
	if err != nil {
		return fmt.Errorf("分析リクエストの削除に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("履歴が見つかりません: %s", id)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	// 画像ファイルを削除（画像入力の場合のみ、データベース削除成功後）
	if inputType == string(InputTypeImage) && imagePath.Valid && imagePath.String != "" {
		if err := os.Remove(imagePath.String); err != nil {
			// 画像削除失敗は警告ログのみ（データベース削除は成功しているため）
			log.Printf("Warning: Failed to remove image file %s: %v", imagePath.String, err)
		} else {
			log.Printf("Image file removed: %s", imagePath.String)
		}
	}

	return nil
}

// GetDailyMeals は指定された日付の食事データを取得します
func (r *postgresAnalysisRepository) GetDailyMeals(ctx context.Context, date string) (map[string][]HistoryDetail, DailyTotal, error) {
	query := `
		SELECT
			ar.id,
			ar.input_type,
			ar.image_path,
			ar.input_text,
			ar.created_at,
			ar.meal_type,
			ar.meal_date,
			res.foods,
			res.total_calories,
			res.total_protein,
			res.total_fat,
			res.total_carbohydrates
		FROM analysis_requests ar
		INNER JOIN analysis_results res ON ar.id = res.analysis_request_id
		WHERE ar.meal_date = $1 AND ar.status = $2
		ORDER BY ar.meal_type, ar.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, date, StatusCompleted)
	if err != nil {
		return nil, DailyTotal{}, fmt.Errorf("日次食事の取得に失敗: %w", err)
	}
	defer rows.Close()

	meals := map[string][]HistoryDetail{
		"breakfast": {},
		"lunch":     {},
		"dinner":    {},
		"snack":     {},
	}

	var dailyTotal DailyTotal

	for rows.Next() {
		var detail HistoryDetail
		var imagePath sql.NullString
		var inputText sql.NullString
		var mealType sql.NullString
		var mealDate sql.NullTime
		var foodsJSON []byte

		err := rows.Scan(
			&detail.ID,
			&detail.InputType,
			&imagePath,
			&inputText,
			&detail.CreatedAt,
			&mealType,
			&mealDate,
			&foodsJSON,
			&detail.TotalCalories,
			&detail.TotalProtein,
			&detail.TotalFat,
			&detail.TotalCarbohydrates,
		)
		if err != nil {
			return nil, DailyTotal{}, fmt.Errorf("行のスキャンに失敗: %w", err)
		}

		if imagePath.Valid {
			detail.ImagePath = imagePath.String
		}
		if inputText.Valid {
			detail.InputText = inputText.String
		}
		if mealType.Valid {
			detail.MealType = mealType.String
		}
		if mealDate.Valid {
			detail.MealDate = mealDate.Time
		}

		if err := json.Unmarshal(foodsJSON, &detail.Foods); err != nil {
			return nil, DailyTotal{}, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
		}

		// 食事タイプが有効な場合のみ追加
		if mealType.Valid && mealType.String != "" {
			meals[mealType.String] = append(meals[mealType.String], detail)
		}

		// 合計を加算
		dailyTotal.TotalCalories += detail.TotalCalories
		dailyTotal.TotalProtein += detail.TotalProtein
		dailyTotal.TotalFat += detail.TotalFat
		dailyTotal.TotalCarbohydrates += detail.TotalCarbohydrates
	}

	if err := rows.Err(); err != nil {
		return nil, DailyTotal{}, fmt.Errorf("行の反復処理に失敗: %w", err)
	}

	return meals, dailyTotal, nil
}
