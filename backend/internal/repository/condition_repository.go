package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConditionRecord struct {
	ID         uuid.UUID `json:"id"`
	UserID     string    `json:"user_id"`
	Condition  int       `json:"condition"`
	Fatigue    int       `json:"fatigue"`
	RecordedAt string    `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConditionRepository interface {
	CreateOrUpdateRecord(ctx context.Context, userID string, condition, fatigue int, recordedAt string) (*ConditionRecord, error)
	GetRecordByDate(ctx context.Context, userID string, date string) (*ConditionRecord, error)
}

type postgresConditionRepository struct {
	db *sql.DB
}

func NewConditionRepository(db *sql.DB) ConditionRepository {
	return &postgresConditionRepository{db: db}
}

func (r *postgresConditionRepository) CreateOrUpdateRecord(ctx context.Context, userID string, condition, fatigue int, recordedAt string) (*ConditionRecord, error) {
	query := `
		INSERT INTO condition_records (user_id, condition, fatigue, recorded_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, recorded_at)
		DO UPDATE SET condition = EXCLUDED.condition, fatigue = EXCLUDED.fatigue
		RETURNING id, user_id, condition, fatigue, recorded_at, created_at
	`

	var record ConditionRecord
	var recordedAtTime time.Time

	err := r.db.QueryRowContext(ctx, query, userID, condition, fatigue, recordedAt).Scan(
		&record.ID,
		&record.UserID,
		&record.Condition,
		&record.Fatigue,
		&recordedAtTime,
		&record.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("体調記録の保存に失敗: %w", err)
	}

	record.RecordedAt = recordedAtTime.Format("2006-01-02")
	return &record, nil
}

func (r *postgresConditionRepository) GetRecordByDate(ctx context.Context, userID string, date string) (*ConditionRecord, error) {
	query := `
		SELECT id, user_id, condition, fatigue, recorded_at, created_at
		FROM condition_records
		WHERE user_id = $1 AND recorded_at = $2
	`

	var record ConditionRecord
	var recordedAtTime time.Time

	err := r.db.QueryRowContext(ctx, query, userID, date).Scan(
		&record.ID,
		&record.UserID,
		&record.Condition,
		&record.Fatigue,
		&recordedAtTime,
		&record.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("体調記録の取得に失敗: %w", err)
	}

	record.RecordedAt = recordedAtTime.Format("2006-01-02")
	return &record, nil
}
