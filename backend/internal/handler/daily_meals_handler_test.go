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

	handler := NewDailyMealsHandler(mockRepo)

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

	handler := NewDailyMealsHandler(mockRepo)

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

	handler := NewDailyMealsHandler(mockRepo)

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

	handler := NewDailyMealsHandler(mockRepo)

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

	handler := NewDailyMealsHandler(mockRepo)

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

	handler := NewDailyMealsHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?tz=Asia/Tokyo", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDailyMealsHandler_Handle_InvalidTimezone(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetDailyMealsFunc: func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "Invalid/Zone", tz)
			expectedDate := time.Now().UTC().Format("2006-01-02")
			assert.Equal(t, expectedDate, date)
			return map[string][]repository.HistoryDetail{
				"breakfast": {},
				"lunch":     {},
				"dinner":    {},
				"snack":     {},
			}, repository.DailyTotal{}, nil
		},
	}

	handler := NewDailyMealsHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?tz=Invalid/Zone", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDailyMealsHandler_Handle_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewDailyMealsHandler(mockRepo)

	// コンテキストにユーザーIDを設定しない
	req := httptest.NewRequest(http.MethodGet, "/api/meals/daily?date=2026-01-21", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
