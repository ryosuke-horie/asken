package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupProfileMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock の作成に失敗: %v", err)
	}
	return db, mock
}

func TestProfileRepository_GetByUserID_Success(t *testing.T) {
	db, mock := setupProfileMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	profileID := uuid.New()
	sportType := "柔術"
	trainingGoals := pq.StringArray{"減量", "スタミナ強化"}
	weightClass := 65
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "sport_type", "training_goals", "weight_class", "created_at", "updated_at"}).
		AddRow(profileID, userID, sportType, trainingGoals, weightClass, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT id, user_id, sport_type, training_goals, weight_class, created_at, updated_at FROM user_profiles`).
		WithArgs(userID).
		WillReturnRows(rows)

	profile, err := repo.GetByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, profileID, profile.ID)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, sportType, *profile.SportType)
	assert.Equal(t, []string{"減量", "スタミナ強化"}, profile.TrainingGoals)
	assert.Equal(t, weightClass, *profile.WeightClass)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepository_GetByUserID_NotFound(t *testing.T) {
	db, mock := setupProfileMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, user_id, sport_type, training_goals, weight_class, created_at, updated_at FROM user_profiles`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	profile, err := repo.GetByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Nil(t, profile)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepository_GetByUserID_DBError(t *testing.T) {
	db, mock := setupProfileMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, user_id, sport_type, training_goals, weight_class, created_at, updated_at FROM user_profiles`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	profile, err := repo.GetByUserID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "プロフィールの取得に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepository_CreateOrUpdate_Success(t *testing.T) {
	db, mock := setupProfileMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	profileID := uuid.New()
	sportType := "柔術"
	trainingGoals := []string{"減量", "スタミナ強化"}
	weightClass := 65
	createdAt := time.Now()
	updatedAt := time.Now()

	input := &UserProfile{
		UserID:        userID,
		SportType:     &sportType,
		TrainingGoals: trainingGoals,
		WeightClass:   &weightClass,
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "sport_type", "training_goals", "weight_class", "created_at", "updated_at"}).
		AddRow(profileID, userID, sportType, pq.StringArray(trainingGoals), weightClass, createdAt, updatedAt)

	mock.ExpectQuery(`INSERT INTO user_profiles`).
		WithArgs(userID, &sportType, pq.Array(trainingGoals), &weightClass).
		WillReturnRows(rows)

	profile, err := repo.CreateOrUpdate(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, profileID, profile.ID)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, sportType, *profile.SportType)
	assert.Equal(t, trainingGoals, profile.TrainingGoals)
	assert.Equal(t, weightClass, *profile.WeightClass)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepository_CreateOrUpdate_DBError(t *testing.T) {
	db, mock := setupProfileMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	sportType := "柔術"
	trainingGoals := []string{"減量"}
	weightClass := 65

	input := &UserProfile{
		UserID:        userID,
		SportType:     &sportType,
		TrainingGoals: trainingGoals,
		WeightClass:   &weightClass,
	}

	mock.ExpectQuery(`INSERT INTO user_profiles`).
		WithArgs(userID, &sportType, pq.Array(trainingGoals), &weightClass).
		WillReturnError(sql.ErrConnDone)

	profile, err := repo.CreateOrUpdate(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "プロフィールの保存に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepository_CreateOrUpdate_WithNilValues(t *testing.T) {
	db, mock := setupProfileMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	profileID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	input := &UserProfile{
		UserID:        userID,
		SportType:     nil,
		TrainingGoals: nil,
		WeightClass:   nil,
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "sport_type", "training_goals", "weight_class", "created_at", "updated_at"}).
		AddRow(profileID, userID, nil, pq.StringArray{}, nil, createdAt, updatedAt)

	mock.ExpectQuery(`INSERT INTO user_profiles`).
		WithArgs(userID, nil, pq.Array([]string(nil)), nil).
		WillReturnRows(rows)

	profile, err := repo.CreateOrUpdate(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, profileID, profile.ID)
	assert.Nil(t, profile.SportType)
	assert.Empty(t, profile.TrainingGoals)
	assert.Nil(t, profile.WeightClass)
	assert.NoError(t, mock.ExpectationsWereMet())
}
