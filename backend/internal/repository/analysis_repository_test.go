package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/service"
	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock の作成に失敗: %v", err)
	}
	return db, mock
}

func TestCreateRequest(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	imagePath := "/uploads/test.jpg"
	expectedID := uuid.New()

	// モック設定
	mock.ExpectQuery(`INSERT INTO analysis_requests`).
		WithArgs(StatusPending, imagePath).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	// 実行
	id, err := repo.CreateRequest(ctx, imagePath)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateRequest_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	imagePath := "/uploads/test.jpg"

	// モック設定 - エラーを返す
	mock.ExpectQuery(`INSERT INTO analysis_requests`).
		WithArgs(StatusPending, imagePath).
		WillReturnError(sql.ErrConnDone)

	// 実行
	id, err := repo.CreateRequest(ctx, imagePath)

	// 検証
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "分析リクエストの作成に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRequest(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// モック設定
	rows := sqlmock.NewRows([]string{"id", "status", "image_path", "error_message", "created_at", "updated_at"}).
		AddRow(requestID, StatusPending, "/uploads/test.jpg", nil, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT id, status, image_path, error_message, created_at, updated_at FROM analysis_requests`).
		WithArgs(requestID).
		WillReturnRows(rows)

	// 実行
	req, err := repo.GetRequest(ctx, requestID)

	// 検証
	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, requestID, req.ID)
	assert.Equal(t, StatusPending, req.Status)
	assert.Equal(t, "/uploads/test.jpg", req.ImagePath)
	assert.Empty(t, req.ErrorMessage)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRequest_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// モック設定 - 結果なし
	mock.ExpectQuery(`SELECT id, status, image_path, error_message, created_at, updated_at FROM analysis_requests`).
		WithArgs(requestID).
		WillReturnError(sql.ErrNoRows)

	// 実行
	req, err := repo.GetRequest(ctx, requestID)

	// 検証
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "リクエストが見つかりません")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatus(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// モック設定
	mock.ExpectExec(`UPDATE analysis_requests SET status`).
		WithArgs(StatusProcessing, sql.NullString{}, requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 実行
	err := repo.UpdateStatus(ctx, requestID, StatusProcessing, "")

	// 検証
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatus_WithErrorMessage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	errorMsg := "Gemini API タイムアウト"

	// モック設定
	mock.ExpectExec(`UPDATE analysis_requests SET status`).
		WithArgs(StatusFailed, sql.NullString{String: errorMsg, Valid: true}, requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 実行
	err := repo.UpdateStatus(ctx, requestID, StatusFailed, errorMsg)

	// 検証
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatus_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// モック設定 - 影響を受けた行が0
	mock.ExpectExec(`UPDATE analysis_requests SET status`).
		WithArgs(StatusProcessing, sql.NullString{}, requestID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// 実行
	err := repo.UpdateStatus(ctx, requestID, StatusProcessing, "")

	// 検証
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "リクエストが見つかりません")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveResult(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	result := &service.AnalysisResult{
		Foods: []gemini.NutritionInfo{
			{
				Name:            "白米",
				EstimatedAmount: "150g",
				Calories:        252,
				Protein:         3.8,
				Fat:             0.5,
				Carbohydrates:   55.7,
			},
		},
		TotalCalories:      252,
		TotalProtein:       3.8,
		TotalFat:           0.5,
		TotalCarbohydrates: 55.7,
	}

	foodsJSON, _ := json.Marshal(result.Foods)

	// トランザクション開始
	mock.ExpectBegin()

	// analysis_results へのINSERT
	mock.ExpectExec(`INSERT INTO analysis_results`).
		WithArgs(requestID, foodsJSON, 252.0, 3.8, 0.5, 55.7).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// analysis_requests のUPDATE
	mock.ExpectExec(`UPDATE analysis_requests SET status`).
		WithArgs(StatusCompleted, requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// トランザクションコミット
	mock.ExpectCommit()

	// 実行
	err := repo.SaveResult(ctx, requestID, result)

	// 検証
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveResult_RollbackOnError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	result := &service.AnalysisResult{
		Foods:              []gemini.NutritionInfo{},
		TotalCalories:      252,
		TotalProtein:       3.8,
		TotalFat:           0.5,
		TotalCarbohydrates: 55.7,
	}

	foodsJSON, _ := json.Marshal(result.Foods)

	// トランザクション開始
	mock.ExpectBegin()

	// analysis_results へのINSERT - エラーを返す
	mock.ExpectExec(`INSERT INTO analysis_results`).
		WithArgs(requestID, foodsJSON, 252.0, 3.8, 0.5, 55.7).
		WillReturnError(sql.ErrConnDone)

	// トランザクションロールバック
	mock.ExpectRollback()

	// 実行
	err := repo.SaveResult(ctx, requestID, result)

	// 検証
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分析結果の保存に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetResult(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	foods := []gemini.NutritionInfo{
		{
			Name:            "白米",
			EstimatedAmount: "150g",
			Calories:        252,
			Protein:         3.8,
			Fat:             0.5,
			Carbohydrates:   55.7,
		},
	}
	foodsJSON, _ := json.Marshal(foods)

	// モック設定
	rows := sqlmock.NewRows([]string{"foods", "total_calories", "total_protein", "total_fat", "total_carbohydrates"}).
		AddRow(foodsJSON, 252.0, 3.8, 0.5, 55.7)

	mock.ExpectQuery(`SELECT foods, total_calories, total_protein, total_fat, total_carbohydrates FROM analysis_results`).
		WithArgs(requestID).
		WillReturnRows(rows)

	// 実行
	result, err := repo.GetResult(ctx, requestID)

	// 検証
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 252.0, result.TotalCalories)
	assert.Equal(t, 3.8, result.TotalProtein)
	assert.Equal(t, 0.5, result.TotalFat)
	assert.Equal(t, 55.7, result.TotalCarbohydrates)
	assert.Len(t, result.Foods, 1)
	assert.Equal(t, "白米", result.Foods[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetResult_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// モック設定 - 結果なし
	mock.ExpectQuery(`SELECT foods, total_calories, total_protein, total_fat, total_carbohydrates FROM analysis_results`).
		WithArgs(requestID).
		WillReturnError(sql.ErrNoRows)

	// 実行
	result, err := repo.GetResult(ctx, requestID)

	// 検証
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "結果が見つかりません")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPendingRequests(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// モック設定
	rows := sqlmock.NewRows([]string{"id", "status", "image_path", "error_message", "created_at", "updated_at"}).
		AddRow(id1, StatusPending, "/uploads/test1.jpg", nil, createdAt, updatedAt).
		AddRow(id2, StatusPending, "/uploads/test2.jpg", nil, createdAt.Add(1*time.Minute), updatedAt)

	mock.ExpectQuery(`SELECT id, status, image_path, error_message, created_at, updated_at FROM analysis_requests`).
		WithArgs(StatusPending, 10).
		WillReturnRows(rows)

	// 実行
	requests, err := repo.GetPendingRequests(ctx, 10)

	// 検証
	assert.NoError(t, err)
	assert.Len(t, requests, 2)
	assert.Equal(t, id1, requests[0].ID)
	assert.Equal(t, id2, requests[1].ID)
	assert.Equal(t, StatusPending, requests[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPendingRequests_Empty(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	// モック設定 - 空の結果
	rows := sqlmock.NewRows([]string{"id", "status", "image_path", "error_message", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, status, image_path, error_message, created_at, updated_at FROM analysis_requests`).
		WithArgs(StatusPending, 10).
		WillReturnRows(rows)

	// 実行
	requests, err := repo.GetPendingRequests(ctx, 10)

	// 検証
	assert.NoError(t, err)
	assert.Empty(t, requests)
	assert.NoError(t, mock.ExpectationsWereMet())
}
