package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WeightRecord struct {
	ID         uuid.UUID `json:"id"`
	UserID     string    `json:"user_id"`
	Weight     float64   `json:"weight"`
	RecordedAt string    `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type WeightGoal struct {
	ID           uuid.UUID `json:"id"`
	UserID       string    `json:"user_id"`
	TargetWeight float64   `json:"target_weight"`
	TargetDate   string    `json:"target_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WeightStats struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
}

type WeightRepository interface {
	CreateOrUpdateRecord(ctx context.Context, userID string, weight float64, recordedAt string) (*WeightRecord, error)
	GetRecordsByPeriod(ctx context.Context, userID string, startDate, endDate string) ([]*WeightRecord, error)
	GetLatestRecord(ctx context.Context, userID string) (*WeightRecord, error)
	GetStatsByPeriod(ctx context.Context, userID string, startDate, endDate string) (*WeightStats, error)
	GetGoal(ctx context.Context, userID string) (*WeightGoal, error)
	CreateOrUpdateGoal(ctx context.Context, userID string, targetWeight float64, targetDate string) (*WeightGoal, error)
}

type postgresWeightRepository struct {
	db *sql.DB
}

func NewWeightRepository(db *sql.DB) WeightRepository {
	return &postgresWeightRepository{db: db}
}

func (r *postgresWeightRepository) CreateOrUpdateRecord(ctx context.Context, userID string, weight float64, recordedAt string) (*WeightRecord, error) {
	query := `
		INSERT INTO weight_records (user_id, weight, recorded_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, recorded_at)
		DO UPDATE SET weight = EXCLUDED.weight
		RETURNING id, user_id, weight, recorded_at, created_at
	`

	var record WeightRecord
	var recordedAtTime time.Time

	err := r.db.QueryRowContext(ctx, query, userID, weight, recordedAt).Scan(
		&record.ID,
		&record.UserID,
		&record.Weight,
		&recordedAtTime,
		&record.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("体重記録の保存に失敗: %w", err)
	}

	record.RecordedAt = recordedAtTime.Format("2006-01-02")
	return &record, nil
}

func (r *postgresWeightRepository) GetRecordsByPeriod(ctx context.Context, userID string, startDate, endDate string) ([]*WeightRecord, error) {
	query := `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_records
		WHERE user_id = $1 AND recorded_at >= $2 AND recorded_at <= $3
		ORDER BY recorded_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("体重記録の取得に失敗: %w", err)
	}
	defer rows.Close()

	var records []*WeightRecord
	for rows.Next() {
		var record WeightRecord
		var recordedAtTime time.Time

		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Weight,
			&recordedAtTime,
			&record.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("体重記録のスキャンに失敗: %w", err)
		}

		record.RecordedAt = recordedAtTime.Format("2006-01-02")
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("体重記録の読み取りに失敗: %w", err)
	}

	return records, nil
}

func (r *postgresWeightRepository) GetLatestRecord(ctx context.Context, userID string) (*WeightRecord, error) {
	query := `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_records
		WHERE user_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`

	var record WeightRecord
	var recordedAtTime time.Time

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&record.ID,
		&record.UserID,
		&record.Weight,
		&recordedAtTime,
		&record.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("最新の体重記録の取得に失敗: %w", err)
	}

	record.RecordedAt = recordedAtTime.Format("2006-01-02")
	return &record, nil
}

func (r *postgresWeightRepository) GetStatsByPeriod(ctx context.Context, userID string, startDate, endDate string) (*WeightStats, error) {
	query := `
		SELECT
			COALESCE(MIN(weight), 0) as min,
			COALESCE(MAX(weight), 0) as max,
			COALESCE(AVG(weight), 0) as average
		FROM weight_records
		WHERE user_id = $1 AND recorded_at >= $2 AND recorded_at <= $3
	`

	var stats WeightStats
	err := r.db.QueryRowContext(ctx, query, userID, startDate, endDate).Scan(
		&stats.Min,
		&stats.Max,
		&stats.Average,
	)

	if err != nil {
		return nil, fmt.Errorf("体重統計の取得に失敗: %w", err)
	}

	return &stats, nil
}

func (r *postgresWeightRepository) GetGoal(ctx context.Context, userID string) (*WeightGoal, error) {
	query := `
		SELECT id, user_id, target_weight, target_date, created_at, updated_at
		FROM weight_goals
		WHERE user_id = $1
	`

	var goal WeightGoal
	var targetDateTime time.Time

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.TargetWeight,
		&targetDateTime,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("目標体重の取得に失敗: %w", err)
	}

	goal.TargetDate = targetDateTime.Format("2006-01-02")
	return &goal, nil
}

func (r *postgresWeightRepository) CreateOrUpdateGoal(ctx context.Context, userID string, targetWeight float64, targetDate string) (*WeightGoal, error) {
	query := `
		INSERT INTO weight_goals (user_id, target_weight, target_date)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET target_weight = EXCLUDED.target_weight, target_date = EXCLUDED.target_date, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, target_weight, target_date, created_at, updated_at
	`

	var goal WeightGoal
	var targetDateTimeParsed time.Time

	err := r.db.QueryRowContext(ctx, query, userID, targetWeight, targetDate).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.TargetWeight,
		&targetDateTimeParsed,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("目標体重の保存に失敗: %w", err)
	}

	goal.TargetDate = targetDateTimeParsed.Format("2006-01-02")
	return &goal, nil
}
