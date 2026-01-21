package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryHandler_HandleList_Success(t *testing.T) {
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	id1 := uuid.New()
	id2 := uuid.New()

	mockRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, page, limit int) ([]repository.HistoryItem, int, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, limit)
			return []repository.HistoryItem{
				{
					ID:                 id1,
					ImagePath:          "/uploads/test1.jpg",
					CreatedAt:          createdAt,
					MealType:           "lunch",
					MealDate:           mealDate,
					TotalCalories:      500.0,
					TotalProtein:       20.0,
					TotalFat:           15.0,
					TotalCarbohydrates: 60.0,
				},
				{
					ID:                 id2,
					ImagePath:          "/uploads/test2.jpg",
					CreatedAt:          createdAt,
					MealType:           "dinner",
					MealDate:           mealDate,
					TotalCalories:      300.0,
					TotalProtein:       10.0,
					TotalFat:           8.0,
					TotalCarbohydrates: 40.0,
				},
			}, 2, nil
		},
	}

	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(20), response["limit"])

	items := response["items"].([]interface{})
	assert.Len(t, items, 2)
}

func TestHistoryHandler_HandleList_WithPagination(t *testing.T) {
	mockRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, page, limit int) ([]repository.HistoryItem, int, error) {
			assert.Equal(t, 2, page)
			assert.Equal(t, 10, limit)
			return []repository.HistoryItem{}, 50, nil
		},
	}

	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history?page=2&limit=10", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(50), response["total"])
	assert.Equal(t, float64(2), response["page"])
	assert.Equal(t, float64(10), response["limit"])
}

func TestHistoryHandler_HandleList_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/history", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHistoryHandler_HandleList_RepositoryError(t *testing.T) {
	mockRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, page, limit int) ([]repository.HistoryItem, int, error) {
			return nil, 0, assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleDetail_Success(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, id uuid.UUID) (*repository.HistoryDetail, error) {
			assert.Equal(t, historyID, id)
			return &repository.HistoryDetail{
				HistoryItem: repository.HistoryItem{
					ID:                 historyID,
					ImagePath:          "/uploads/test.jpg",
					CreatedAt:          createdAt,
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
			}, nil
		},
	}

	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+historyID.String(), nil)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response repository.HistoryDetail
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, historyID, response.ID)
	assert.Equal(t, 500.0, response.TotalCalories)
	assert.Len(t, response.Foods, 1)
}

func TestHistoryHandler_HandleDetail_NotFound(t *testing.T) {
	historyID := uuid.New()

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+historyID.String(), nil)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleDetail_InvalidUUID(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history/invalid-uuid", nil)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHistoryHandler_HandleDetail_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/history/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
