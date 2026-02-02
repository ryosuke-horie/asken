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

func setupWeightMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock の作成に失敗: %v", err)
	}
	return db, mock
}

func TestCreateOrUpdateRecord_Success(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	recordID := uuid.New()
	weight := 65.5
	recordedAt := "2024-01-15"
	recordedAtTime, _ := time.Parse("2006-01-02", recordedAt)
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "weight", "recorded_at", "created_at"}).
		AddRow(recordID, userID, weight, recordedAtTime, createdAt)

	mock.ExpectQuery(`INSERT INTO weight_records`).
		WithArgs(userID, weight, recordedAt).
		WillReturnRows(rows)

	record, err := repo.CreateOrUpdateRecord(ctx, userID, weight, recordedAt)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, recordID, record.ID)
	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, weight, record.Weight)
	assert.Equal(t, recordedAt, record.RecordedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrUpdateRecord_DBError(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	weight := 65.5
	recordedAt := "2024-01-15"

	mock.ExpectQuery(`INSERT INTO weight_records`).
		WithArgs(userID, weight, recordedAt).
		WillReturnError(sql.ErrConnDone)

	record, err := repo.CreateOrUpdateRecord(ctx, userID, weight, recordedAt)

	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "体重記録の保存に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecordsByPeriod_Success(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	startDate := "2024-01-01"
	endDate := "2024-01-15"
	record1Time, _ := time.Parse("2006-01-02", "2024-01-10")
	record2Time, _ := time.Parse("2006-01-02", "2024-01-15")

	rows := sqlmock.NewRows([]string{"id", "user_id", "weight", "recorded_at", "created_at"}).
		AddRow(uuid.New(), userID, 65.0, record1Time, time.Now()).
		AddRow(uuid.New(), userID, 65.5, record2Time, time.Now())

	mock.ExpectQuery(`SELECT id, user_id, weight, recorded_at, created_at FROM weight_records`).
		WithArgs(userID, startDate, endDate).
		WillReturnRows(rows)

	records, err := repo.GetRecordsByPeriod(ctx, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, 65.0, records[0].Weight)
	assert.Equal(t, 65.5, records[1].Weight)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecordsByPeriod_Empty(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	startDate := "2024-01-01"
	endDate := "2024-01-15"

	rows := sqlmock.NewRows([]string{"id", "user_id", "weight", "recorded_at", "created_at"})

	mock.ExpectQuery(`SELECT id, user_id, weight, recorded_at, created_at FROM weight_records`).
		WithArgs(userID, startDate, endDate).
		WillReturnRows(rows)

	records, err := repo.GetRecordsByPeriod(ctx, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.Empty(t, records)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecordsByPeriod_DBError(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	startDate := "2024-01-01"
	endDate := "2024-01-15"

	mock.ExpectQuery(`SELECT id, user_id, weight, recorded_at, created_at FROM weight_records`).
		WithArgs(userID, startDate, endDate).
		WillReturnError(sql.ErrConnDone)

	records, err := repo.GetRecordsByPeriod(ctx, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "体重記録の取得に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLatestRecord_Success(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	recordID := uuid.New()
	recordedAtTime, _ := time.Parse("2006-01-02", "2024-01-15")

	rows := sqlmock.NewRows([]string{"id", "user_id", "weight", "recorded_at", "created_at"}).
		AddRow(recordID, userID, 65.5, recordedAtTime, time.Now())

	mock.ExpectQuery(`SELECT id, user_id, weight, recorded_at, created_at FROM weight_records`).
		WithArgs(userID).
		WillReturnRows(rows)

	record, err := repo.GetLatestRecord(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, recordID, record.ID)
	assert.Equal(t, 65.5, record.Weight)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLatestRecord_NotFound(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"

	mock.ExpectQuery(`SELECT id, user_id, weight, recorded_at, created_at FROM weight_records`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	record, err := repo.GetLatestRecord(ctx, userID)

	assert.NoError(t, err)
	assert.Nil(t, record)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLatestRecord_DBError(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"

	mock.ExpectQuery(`SELECT id, user_id, weight, recorded_at, created_at FROM weight_records`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	record, err := repo.GetLatestRecord(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "最新の体重記録の取得に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStatsByPeriod_Success(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	startDate := "2024-01-01"
	endDate := "2024-01-15"

	rows := sqlmock.NewRows([]string{"min", "max", "average"}).
		AddRow(65.0, 67.5, 66.25)

	mock.ExpectQuery(`SELECT`).
		WithArgs(userID, startDate, endDate).
		WillReturnRows(rows)

	stats, err := repo.GetStatsByPeriod(ctx, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 65.0, stats.Min)
	assert.Equal(t, 67.5, stats.Max)
	assert.Equal(t, 66.25, stats.Average)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStatsByPeriod_DBError(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	startDate := "2024-01-01"
	endDate := "2024-01-15"

	mock.ExpectQuery(`SELECT`).
		WithArgs(userID, startDate, endDate).
		WillReturnError(sql.ErrConnDone)

	stats, err := repo.GetStatsByPeriod(ctx, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "体重統計の取得に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGoal_Success(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	goalID := uuid.New()
	targetDate, _ := time.Parse("2006-01-02", "2024-06-01")
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "target_weight", "target_date", "created_at", "updated_at"}).
		AddRow(goalID, userID, 60.0, targetDate, createdAt, createdAt)

	mock.ExpectQuery(`SELECT id, user_id, target_weight, target_date, created_at, updated_at FROM weight_goals`).
		WithArgs(userID).
		WillReturnRows(rows)

	goal, err := repo.GetGoal(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	assert.Equal(t, goalID, goal.ID)
	assert.Equal(t, 60.0, goal.TargetWeight)
	assert.Equal(t, "2024-06-01", goal.TargetDate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGoal_NotFound(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"

	mock.ExpectQuery(`SELECT id, user_id, target_weight, target_date, created_at, updated_at FROM weight_goals`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	goal, err := repo.GetGoal(ctx, userID)

	assert.NoError(t, err)
	assert.Nil(t, goal)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrUpdateGoal_Success(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	goalID := uuid.New()
	targetWeight := 60.0
	targetDate := "2024-06-01"
	targetDateTime, _ := time.Parse("2006-01-02", targetDate)
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "target_weight", "target_date", "created_at", "updated_at"}).
		AddRow(goalID, userID, targetWeight, targetDateTime, createdAt, createdAt)

	mock.ExpectQuery(`INSERT INTO weight_goals`).
		WithArgs(userID, targetWeight, targetDate).
		WillReturnRows(rows)

	goal, err := repo.CreateOrUpdateGoal(ctx, userID, targetWeight, targetDate)

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	assert.Equal(t, goalID, goal.ID)
	assert.Equal(t, targetWeight, goal.TargetWeight)
	assert.Equal(t, targetDate, goal.TargetDate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrUpdateGoal_DBError(t *testing.T) {
	db, mock := setupWeightMockDB(t)
	defer db.Close()

	repo := NewWeightRepository(db)
	ctx := context.Background()

	userID := "test-firebase-uid"
	targetWeight := 60.0
	targetDate := "2024-06-01"

	mock.ExpectQuery(`INSERT INTO weight_goals`).
		WithArgs(userID, targetWeight, targetDate).
		WillReturnError(sql.ErrConnDone)

	goal, err := repo.CreateOrUpdateGoal(ctx, userID, targetWeight, targetDate)

	assert.Error(t, err)
	assert.Nil(t, goal)
	assert.Contains(t, err.Error(), "目標体重の保存に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}
