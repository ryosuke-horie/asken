package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailyMealsHandler_Handle_Success(t *testing.T) {
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "2026-01-21", date)
			assert.Equal(t, "UTC", tz) // デフォルトはUTC
			meals := map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch": {
					{
						HistoryItem: repository.HistoryItem{
							ID:                 id,
							ImagePath:          "/uploads/test.jpg",
							CreatedAt:          time.Now(),
							MealType:           "lunch",
							MealDate:           mealDate,
							TotalCalories:      500.0,
							TotalProtein:       20.0,
							TotalFat:           15.0,
							TotalCarbohydrates: 60.0,
						},
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
					},
				},
				"dinner": {},
				"snack":  {},
			}
			total := repository.DailyTotal{
				TotalCalories:      500.0,
				TotalProtein:       20.0,
				TotalFat:           15.0,
				TotalCarbohydrates: 60.0,
			}
			return meals, total, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	// コンテキストにユーザーIDを設定
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "2026-01-21", response.Date)
	assert.Equal(t, 500.0, response.DailyTotal.TotalCalories)
	assert.Len(t, response.Meals["lunch"], 1)
	assert.Len(t, response.Meals["breakfast"], 0)
}

func TestDailyMealsHandler_Handle_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/meals/daily", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDailyMealsHandler_Handle_DefaultDate(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			assert.Equal(t, testUserID, userID)
			// 今日の日付が渡されるべき（UTCタイムゾーン）
			expectedDate := time.Now().UTC().Format("2006-01-02")
			assert.Equal(t, expectedDate, date)
			assert.Equal(t, "UTC", tz)
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	// dateパラメータなしでリクエスト
	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
	// コンテキストにユーザーIDを設定
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDailyMealsHandler_Handle_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return nil, repository.DailyTotal{}, assert.AnError
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	// コンテキストにユーザーIDを設定
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDailyMealsHandler_Handle_EmptyMeals(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	// コンテキストにユーザーIDを設定
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "2026-01-21", response.Date)
	assert.Equal(t, 0.0, response.DailyTotal.TotalCalories)
	assert.Len(t, response.Meals["breakfast"], 0)
	assert.Len(t, response.Meals["lunch"], 0)
	assert.Len(t, response.Meals["dinner"], 0)
	assert.Len(t, response.Meals["snack"], 0)
}

func TestDailyMealsHandler_Handle_WithTimezone(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "2026-01-21", date)
			assert.Equal(t, "Asia/Tokyo", tz)
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21&tz=Asia/Tokyo", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "2026-01-21", response.Date)
}

func TestDailyMealsHandler_Handle_DefaultDateWithTimezone(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "Asia/Tokyo", tz)
			loc, _ := time.LoadLocation("Asia/Tokyo")
			expectedDate := time.Now().In(loc).Format("2006-01-02")
			assert.Equal(t, expectedDate, date)
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?tz=Asia/Tokyo", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDailyMealsHandler_Handle_InvalidTimezone(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}

	handler := NewDailyMealsHandler(mockRepo, nil)

	// date未指定で無効なtzを指定した場合 → 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?tz=Invalid/Zone", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDailyMealsHandler_Handle_InvalidTimezone_WithDateSpecified(t *testing.T) {
	// 仕様: date指定あり + 無効なtz の場合はtzバリデーションをスキップして200を返す
	// (tzはdate未指定時の現在日付計算にのみ使用される)
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			assert.Equal(t, "2026-01-21", date)
			assert.Equal(t, "Invalid/Zone", tz)
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21&tz=Invalid/Zone", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDailyMealsHandler_Handle_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewDailyMealsHandler(mockRepo, nil)

	// コンテキストにユーザーIDを設定しない
	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDailyMealsHandler_Handle_WithExercise_TotalBurnedCaloriesIncluded(t *testing.T) {
	testUserID := "test-user-123"

	mockAnalysisRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{TotalCalories: 2000.0}, nil
		},
	}

	mockExerciseRepo := &MockExerciseRepository{
		ListByDateFunc: func(ctx context.Context, userID string, recordedDate string) ([]repository.ExerciseRecord, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "2026-02-28", recordedDate)
			return []repository.ExerciseRecord{
				{ID: "rec-1", BurnedCaloriesKcal: 1102.5},
				{ID: "rec-2", BurnedCaloriesKcal: 720.3},
			}, nil
		},
	}

	handler := NewDailyMealsHandler(mockAnalysisRepo, mockExerciseRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.InDelta(t, 1822.8, response.TotalBurnedCaloriesKcal, 0.01)
	assert.Equal(t, 2000.0, response.DailyTotal.TotalCalories)
}

func TestDailyMealsHandler_Handle_ExerciseRepoError_ZeroBurnedCalories(t *testing.T) {
	testUserID := "test-user-123"

	mockAnalysisRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return map[string][]repository.HistoryDetail{}, repository.DailyTotal{TotalCalories: 1500.0}, nil
		},
	}

	mockExerciseRepo := &MockExerciseRepository{
		ListByDateFunc: func(ctx context.Context, userID string, recordedDate string) ([]repository.ExerciseRecord, error) {
			return nil, assert.AnError
		},
	}

	handler := NewDailyMealsHandler(mockAnalysisRepo, mockExerciseRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// 運動記録取得エラーはサイレントに処理され、200 + 0kcal を返す
	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 0.0, response.TotalBurnedCaloriesKcal)
	assert.Equal(t, 1500.0, response.DailyTotal.TotalCalories)
}

func TestDailyMealsHandler_Handle_WithPendingAnalyses(t *testing.T) {
	testUserID := "test-user-123"
	pendingID := uuid.New().String()

	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
		GetPendingAnalysesForDateFunc: func(ctx context.Context, userID string, date string, tz string) ([]repository.PendingAnalysisEntry, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "2026-01-21", date)
			return []repository.PendingAnalysisEntry{
				{
					ID:        pendingID,
					MealType:  "lunch",
					Status:    repository.StatusProcessing,
					InputType: repository.InputTypeText,
					CreatedAt: time.Date(2026, 1, 21, 12, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.PendingAnalyses, 1)
	assert.Equal(t, pendingID, response.PendingAnalyses[0].ID)
	assert.Equal(t, "lunch", response.PendingAnalyses[0].MealType)
	assert.Equal(t, repository.StatusProcessing, response.PendingAnalyses[0].Status)
}

func TestDailyMealsHandler_Handle_PendingAnalysesError_FallbackToEmpty(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return map[string][]repository.HistoryDetail{}, repository.DailyTotal{TotalCalories: 1000.0}, nil
		},
		GetPendingAnalysesForDateFunc: func(ctx context.Context, userID string, date string, tz string) ([]repository.PendingAnalysisEntry, error) {
			return nil, assert.AnError
		},
	}

	handler := NewDailyMealsHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// pending分析取得エラーは非致命的: 200 + 空リストを返す
	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Meals)
	assert.Empty(t, response.PendingAnalyses)
	assert.Equal(t, 1000.0, response.DailyTotal.TotalCalories)
}

func TestDailyMealsHandler_Handle_ExerciseRepoNil_ZeroBurnedCalories(t *testing.T) {
	testUserID := "test-user-123"

	mockAnalysisRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			return map[string][]repository.HistoryDetail{}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockAnalysisRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyMealsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 0.0, response.TotalBurnedCaloriesKcal)
}
