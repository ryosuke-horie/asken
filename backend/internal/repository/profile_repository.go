package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserProfile struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	SportType     *string   `json:"sport_type"`
	TrainingGoals []string  `json:"training_goals"`
	WeightClass   *int      `json:"weight_class"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	CreateOrUpdate(ctx context.Context, profile *UserProfile) (*UserProfile, error)
}

type postgresProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &postgresProfileRepository{db: db}
}

func (r *postgresProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	query := `
		SELECT id, user_id, sport_type, training_goals, weight_class, created_at, updated_at
		FROM user_profiles
		WHERE user_id = $1
	`

	var profile UserProfile
	var trainingGoals pq.StringArray

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.SportType,
		&trainingGoals,
		&profile.WeightClass,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("プロフィールの取得に失敗: %w", err)
	}

	profile.TrainingGoals = trainingGoals
	return &profile, nil
}

func (r *postgresProfileRepository) CreateOrUpdate(ctx context.Context, profile *UserProfile) (*UserProfile, error) {
	query := `
		INSERT INTO user_profiles (user_id, sport_type, training_goals, weight_class)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET
			sport_type = EXCLUDED.sport_type,
			training_goals = EXCLUDED.training_goals,
			weight_class = EXCLUDED.weight_class,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, sport_type, training_goals, weight_class, created_at, updated_at
	`

	var result UserProfile
	var trainingGoals pq.StringArray

	err := r.db.QueryRowContext(ctx, query,
		profile.UserID,
		profile.SportType,
		pq.Array(profile.TrainingGoals),
		profile.WeightClass,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.SportType,
		&trainingGoals,
		&result.WeightClass,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("プロフィールの保存に失敗: %w", err)
	}

	result.TrainingGoals = trainingGoals
	return &result, nil
}
