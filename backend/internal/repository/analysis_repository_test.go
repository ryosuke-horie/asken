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
	mealType := "lunch"
	mealDate := "2026-01-21"
	expectedID := uuid.New()

	// モック設定（input_typeを追加、userIDはnil）
	mock.ExpectQuery(`INSERT INTO analysis_requests`).
		WithArgs(StatusPending, InputTypeImage, imagePath, mealType, mealDate, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	// 実行
	id, err := repo.CreateRequest(ctx, imagePath, mealType, mealDate, nil)

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
	mealType := "lunch"
	mealDate := "2026-01-21"

	// モック設定 - エラーを返す
	mock.ExpectQuery(`INSERT INTO analysis_requests`).
		WithArgs(StatusPending, InputTypeImage, imagePath, mealType, mealDate, nil).
		WillReturnError(sql.ErrConnDone)

	// 実行
	id, err := repo.CreateRequest(ctx, imagePath, mealType, mealDate, nil)

	// 検証
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "分析リクエストの作成に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateRequestWithText(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	inputText := "ご飯二杯, 焼肉"
	mealType := "lunch"
	mealDate := "2026-01-21"
	expectedID := uuid.New()

	// モック設定
	mock.ExpectQuery(`INSERT INTO analysis_requests`).
		WithArgs(StatusPending, InputTypeText, inputText, mealType, mealDate, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	// 実行
	id, err := repo.CreateRequestWithText(ctx, inputText, mealType, mealDate, nil)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
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

	// モック設定（input_type, input_textを追加）
	rows := sqlmock.NewRows([]string{"id", "status", "input_type", "image_path", "input_text", "error_message", "created_at", "updated_at"}).
		AddRow(requestID, StatusPending, InputTypeImage, "/uploads/test.jpg", nil, nil, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT id, status, input_type, image_path, input_text, error_message, created_at, updated_at FROM analysis_requests`).
		WithArgs(requestID).
		WillReturnRows(rows)

	// 実行
	req, err := repo.GetRequest(ctx, requestID)

	// 検証
	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, requestID, req.ID)
	assert.Equal(t, StatusPending, req.Status)
	assert.Equal(t, InputTypeImage, req.InputType)
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
	mock.ExpectQuery(`SELECT id, status, input_type, image_path, input_text, error_message, created_at, updated_at FROM analysis_requests`).
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

	// モック設定（input_type, input_textを追加）
	rows := sqlmock.NewRows([]string{"id", "status", "input_type", "image_path", "input_text", "error_message", "created_at", "updated_at"}).
		AddRow(id1, StatusPending, InputTypeImage, "/uploads/test1.jpg", nil, nil, createdAt, updatedAt).
		AddRow(id2, StatusPending, InputTypeImage, "/uploads/test2.jpg", nil, nil, createdAt.Add(1*time.Minute), updatedAt)

	mock.ExpectQuery(`SELECT id, status, input_type, image_path, input_text, error_message, created_at, updated_at FROM analysis_requests`).
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
	assert.Equal(t, InputTypeImage, requests[0].InputType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPendingRequests_Empty(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	// モック設定 - 空の結果
	rows := sqlmock.NewRows([]string{"id", "status", "input_type", "image_path", "input_text", "error_message", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, status, input_type, image_path, input_text, error_message, created_at, updated_at FROM analysis_requests`).
		WithArgs(StatusPending, 10).
		WillReturnRows(rows)

	// 実行
	requests, err := repo.GetPendingRequests(ctx, 10)

	// 検証
	assert.NoError(t, err)
	assert.Empty(t, requests)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHistoryList(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	createdAt1 := time.Now()
	createdAt2 := time.Now().Add(-1 * time.Hour)
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)

	// 総件数のクエリ
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM analysis_requests`).
		WithArgs(StatusCompleted).
		WillReturnRows(countRows)

	// 履歴一覧のクエリ（input_type, input_textを含む）
	rows := sqlmock.NewRows([]string{"id", "input_type", "image_path", "input_text", "created_at", "meal_type", "meal_date", "total_calories", "total_protein", "total_fat", "total_carbohydrates"}).
		AddRow(id1, InputTypeImage, "/uploads/test1.jpg", nil, createdAt1, "lunch", mealDate, 500.0, 20.0, 15.0, 60.0).
		AddRow(id2, InputTypeImage, "/uploads/test2.jpg", nil, createdAt2, "dinner", mealDate, 300.0, 10.0, 8.0, 40.0)

	mock.ExpectQuery(`SELECT ar.id, ar.input_type, ar.image_path, ar.input_text, ar.created_at, ar.meal_type`).
		WithArgs(StatusCompleted, 20, 0).
		WillReturnRows(rows)

	// 実行
	items, total, err := repo.GetHistoryList(ctx, 1, 20)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, id1, items[0].ID)
	assert.Equal(t, InputTypeImage, items[0].InputType)
	assert.Equal(t, 500.0, items[0].TotalCalories)
	assert.Equal(t, "lunch", items[0].MealType)
	assert.Equal(t, id2, items[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHistoryList_Pagination(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	// 総件数のクエリ
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(50)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM analysis_requests`).
		WithArgs(StatusCompleted).
		WillReturnRows(countRows)

	// 2ページ目を取得（offset = 20）
	rows := sqlmock.NewRows([]string{"id", "input_type", "image_path", "input_text", "created_at", "meal_type", "meal_date", "total_calories", "total_protein", "total_fat", "total_carbohydrates"})

	mock.ExpectQuery(`SELECT ar.id, ar.input_type, ar.image_path, ar.input_text, ar.created_at, ar.meal_type`).
		WithArgs(StatusCompleted, 20, 20).
		WillReturnRows(rows)

	// 実行
	items, total, err := repo.GetHistoryList(ctx, 2, 20)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, 50, total)
	assert.Empty(t, items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHistoryList_InvalidPage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	// 総件数のクエリ
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM analysis_requests`).
		WithArgs(StatusCompleted).
		WillReturnRows(countRows)

	// page < 1 は 1 にフォールバック
	rows := sqlmock.NewRows([]string{"id", "input_type", "image_path", "input_text", "created_at", "meal_type", "meal_date", "total_calories", "total_protein", "total_fat", "total_carbohydrates"})

	mock.ExpectQuery(`SELECT ar.id, ar.input_type, ar.image_path, ar.input_text, ar.created_at, ar.meal_type`).
		WithArgs(StatusCompleted, 20, 0).
		WillReturnRows(rows)

	// 実行（page = 0）
	_, total, err := repo.GetHistoryList(ctx, 0, 20)

	// 検証
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHistoryDetail(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)

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

	// モック設定（input_type, input_textを含む）
	rows := sqlmock.NewRows([]string{"id", "input_type", "image_path", "input_text", "created_at", "meal_type", "meal_date", "foods", "total_calories", "total_protein", "total_fat", "total_carbohydrates"}).
		AddRow(requestID, InputTypeImage, "/uploads/test.jpg", nil, createdAt, "lunch", mealDate, foodsJSON, 252.0, 3.8, 0.5, 55.7)

	mock.ExpectQuery(`SELECT ar.id, ar.input_type, ar.image_path, ar.input_text, ar.created_at, ar.meal_type`).
		WithArgs(requestID, StatusCompleted).
		WillReturnRows(rows)

	// 実行
	detail, err := repo.GetHistoryDetail(ctx, requestID)

	// 検証
	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, requestID, detail.ID)
	assert.Equal(t, InputTypeImage, detail.InputType)
	assert.Equal(t, "/uploads/test.jpg", detail.ImagePath)
	assert.Equal(t, "lunch", detail.MealType)
	assert.Equal(t, 252.0, detail.TotalCalories)
	assert.Len(t, detail.Foods, 1)
	assert.Equal(t, "白米", detail.Foods[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHistoryDetail_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// モック設定 - 結果なし
	mock.ExpectQuery(`SELECT ar.id, ar.input_type, ar.image_path, ar.input_text, ar.created_at, ar.meal_type`).
		WithArgs(requestID, StatusCompleted).
		WillReturnError(sql.ErrNoRows)

	// 実行
	detail, err := repo.GetHistoryDetail(ctx, requestID)

	// 検証
	assert.Error(t, err)
	assert.Nil(t, detail)
	assert.Contains(t, err.Error(), "履歴が見つかりません")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHistory(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	imagePath := "/uploads/test.jpg"

	// トランザクション開始
	mock.ExpectBegin()

	// input_typeと画像パスの取得
	imageRows := sqlmock.NewRows([]string{"input_type", "image_path"}).AddRow(InputTypeImage, imagePath)
	mock.ExpectQuery(`SELECT input_type, image_path FROM analysis_requests WHERE id`).
		WithArgs(requestID).
		WillReturnRows(imageRows)

	// analysis_results の削除
	mock.ExpectExec(`DELETE FROM analysis_results WHERE analysis_request_id`).
		WithArgs(requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// analysis_requests の削除
	mock.ExpectExec(`DELETE FROM analysis_requests WHERE id`).
		WithArgs(requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// トランザクションコミット
	mock.ExpectCommit()

	// 実行
	err := repo.DeleteHistory(ctx, requestID)

	// 検証
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHistory_TextInput(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// トランザクション開始
	mock.ExpectBegin()

	// input_typeの取得（テキスト入力）
	imageRows := sqlmock.NewRows([]string{"input_type", "image_path"}).AddRow(InputTypeText, nil)
	mock.ExpectQuery(`SELECT input_type, image_path FROM analysis_requests WHERE id`).
		WithArgs(requestID).
		WillReturnRows(imageRows)

	// analysis_results の削除
	mock.ExpectExec(`DELETE FROM analysis_results WHERE analysis_request_id`).
		WithArgs(requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// analysis_requests の削除
	mock.ExpectExec(`DELETE FROM analysis_requests WHERE id`).
		WithArgs(requestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// トランザクションコミット
	mock.ExpectCommit()

	// 実行
	err := repo.DeleteHistory(ctx, requestID)

	// 検証（テキスト入力の場合、ファイル削除は行われない）
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHistory_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()

	// トランザクション開始
	mock.ExpectBegin()

	// input_typeと画像パスの取得 - 結果なし
	mock.ExpectQuery(`SELECT input_type, image_path FROM analysis_requests WHERE id`).
		WithArgs(requestID).
		WillReturnError(sql.ErrNoRows)

	// トランザクションロールバック
	mock.ExpectRollback()

	// 実行
	err := repo.DeleteHistory(ctx, requestID)

	// 検証
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "履歴が見つかりません")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHistory_ResultsDeleteError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAnalysisRepository(db)
	ctx := context.Background()

	requestID := uuid.New()
	imagePath := "/uploads/test.jpg"

	// トランザクション開始
	mock.ExpectBegin()

	// input_typeと画像パスの取得
	imageRows := sqlmock.NewRows([]string{"input_type", "image_path"}).AddRow(InputTypeImage, imagePath)
	mock.ExpectQuery(`SELECT input_type, image_path FROM analysis_requests WHERE id`).
		WithArgs(requestID).
		WillReturnRows(imageRows)

	// analysis_results の削除 - エラー
	mock.ExpectExec(`DELETE FROM analysis_results WHERE analysis_request_id`).
		WithArgs(requestID).
		WillReturnError(sql.ErrConnDone)

	// トランザクションロールバック
	mock.ExpectRollback()

	// 実行
	err := repo.DeleteHistory(ctx, requestID)

	// 検証
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分析結果の削除に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}
