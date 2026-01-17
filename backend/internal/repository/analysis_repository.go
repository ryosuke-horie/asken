package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/service"
)

// AnalysisStatus は分析リクエストのステータスを表す型
type AnalysisStatus string

const (
	StatusPending    AnalysisStatus = "pending"
	StatusProcessing AnalysisStatus = "processing"
	StatusCompleted  AnalysisStatus = "completed"
	StatusFailed     AnalysisStatus = "failed"
)

// AnalysisRequest は分析リクエストを表す構造体
type AnalysisRequest struct {
	ID           uuid.UUID      `json:"id"`
	Status       AnalysisStatus `json:"status"`
	ImagePath    string         `json:"image_path"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// AnalysisRepository は分析リクエストと結果の永続化を担当するインターフェース
type AnalysisRepository interface {
	// CreateRequest は新しい分析リクエストを作成します
	CreateRequest(ctx context.Context, imagePath string) (uuid.UUID, error)

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
}

// postgresAnalysisRepository はPostgreSQLを使用したAnalysisRepositoryの実装
type postgresAnalysisRepository struct {
	db *sql.DB
}

// NewAnalysisRepository は新しいAnalysisRepositoryを作成します
func NewAnalysisRepository(db *sql.DB) AnalysisRepository {
	return &postgresAnalysisRepository{db: db}
}

// CreateRequest は新しい分析リクエストを作成します
func (r *postgresAnalysisRepository) CreateRequest(ctx context.Context, imagePath string) (uuid.UUID, error) {
	query := `
		INSERT INTO analysis_requests (status, image_path)
		VALUES ($1, $2)
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, StatusPending, imagePath).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("分析リクエストの作成に失敗: %w", err)
	}

	return id, nil
}

// GetRequest は指定されたIDの分析リクエストを取得します
func (r *postgresAnalysisRepository) GetRequest(ctx context.Context, id uuid.UUID) (*AnalysisRequest, error) {
	query := `
		SELECT id, status, image_path, error_message, created_at, updated_at
		FROM analysis_requests
		WHERE id = $1
	`

	var req AnalysisRequest
	var errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.Status,
		&req.ImagePath,
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
		SELECT id, status, image_path, error_message, created_at, updated_at
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
		var errorMessage sql.NullString

		err := rows.Scan(
			&req.ID,
			&req.Status,
			&req.ImagePath,
			&errorMessage,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("行のスキャンに失敗: %w", err)
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
