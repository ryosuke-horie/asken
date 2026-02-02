package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupConditionMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock の作成に失敗: %v", err)
	}
	return db, mock
}

func TestConditionRepository_CreateOrUpdateRecord_Success(t *testing.T) {
	db, mock := setupConditionMockDB(t)
	defer db.Close()

	repo := NewConditionRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	recordID := uuid.New()
	condition := 3
	fatigue := 2
	recordedAt := "2024-01-15"
	recordedAtTime, _ := time.Parse("2006-01-02", recordedAt)
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "condition", "fatigue", "recorded_at", "created_at"}).
		AddRow(recordID, userID, condition, fatigue, recordedAtTime, createdAt)

	mock.ExpectQuery(`INSERT INTO condition_records`).
		WithArgs(userID, condition, fatigue, recordedAt).
		WillReturnRows(rows)

	record, err := repo.CreateOrUpdateRecord(ctx, userID, condition, fatigue, recordedAt)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, recordID, record.ID)
	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, condition, record.Condition)
	assert.Equal(t, fatigue, record.Fatigue)
	assert.Equal(t, recordedAt, record.RecordedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConditionRepository_CreateOrUpdateRecord_DBError(t *testing.T) {
	db, mock := setupConditionMockDB(t)
	defer db.Close()

	repo := NewConditionRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	condition := 3
	fatigue := 2
	recordedAt := "2024-01-15"

	mock.ExpectQuery(`INSERT INTO condition_records`).
		WithArgs(userID, condition, fatigue, recordedAt).
		WillReturnError(sql.ErrConnDone)

	record, err := repo.CreateOrUpdateRecord(ctx, userID, condition, fatigue, recordedAt)

	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "体調記録の保存に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConditionRepository_GetRecordByDate_Success(t *testing.T) {
	db, mock := setupConditionMockDB(t)
	defer db.Close()

	repo := NewConditionRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	recordID := uuid.New()
	date := "2024-01-15"
	recordedAtTime, _ := time.Parse("2006-01-02", date)
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "condition", "fatigue", "recorded_at", "created_at"}).
		AddRow(recordID, userID, 3, 2, recordedAtTime, createdAt)

	mock.ExpectQuery(`SELECT id, user_id, condition, fatigue, recorded_at, created_at FROM condition_records`).
		WithArgs(userID, date).
		WillReturnRows(rows)

	record, err := repo.GetRecordByDate(ctx, userID, date)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, recordID, record.ID)
	assert.Equal(t, 3, record.Condition)
	assert.Equal(t, 2, record.Fatigue)
	assert.Equal(t, date, record.RecordedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConditionRepository_GetRecordByDate_NotFound(t *testing.T) {
	db, mock := setupConditionMockDB(t)
	defer db.Close()

	repo := NewConditionRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	date := "2024-01-15"

	mock.ExpectQuery(`SELECT id, user_id, condition, fatigue, recorded_at, created_at FROM condition_records`).
		WithArgs(userID, date).
		WillReturnError(sql.ErrNoRows)

	record, err := repo.GetRecordByDate(ctx, userID, date)

	assert.NoError(t, err)
	assert.Nil(t, record)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConditionRepository_GetRecordByDate_DBError(t *testing.T) {
	db, mock := setupConditionMockDB(t)
	defer db.Close()

	repo := NewConditionRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	date := "2024-01-15"

	mock.ExpectQuery(`SELECT id, user_id, condition, fatigue, recorded_at, created_at FROM condition_records`).
		WithArgs(userID, date).
		WillReturnError(sql.ErrConnDone)

	record, err := repo.GetRecordByDate(ctx, userID, date)

	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "体調記録の取得に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}
