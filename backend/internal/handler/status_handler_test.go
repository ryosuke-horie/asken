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
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusHandler_Pending(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, requestID, id)
			return &repository.AnalysisRequest{
				ID:        requestID,
				Status:    repository.StatusPending,
				ImagePath: "/uploads/test.jpg",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "pending", response["status"])
	assert.Equal(t, "分析リクエストを受け付けました", response["message"])
}

func TestStatusHandler_HandlerGuard_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestStatusHandler_Processing(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			assert.Equal(t, testUserID, userID)
			return &repository.AnalysisRequest{
				ID:        requestID,
				Status:    repository.StatusProcessing,
				ImagePath: "/uploads/test.jpg",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "processing", response["status"])
	assert.Equal(t, "分析処理中です", response["message"])
}

func TestStatusHandler_Completed(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			assert.Equal(t, testUserID, userID)
			return &repository.AnalysisRequest{
				ID:        requestID,
				Status:    repository.StatusCompleted,
				ImagePath: "/uploads/test.jpg",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
		GetResultFunc: func(ctx context.Context, userID string, requestID uuid.UUID) (*service.AnalysisResult, error) {
			assert.Equal(t, testUserID, userID)
			return &service.AnalysisResult{
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
			}, nil
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "completed", response["status"])
	assert.NotNil(t, response["result"])

	result := response["result"].(map[string]interface{})
	assert.Equal(t, 252.0, result["total_calories"])
}

func TestStatusHandler_Failed(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			assert.Equal(t, testUserID, userID)
			return &repository.AnalysisRequest{
				ID:           requestID,
				Status:       repository.StatusFailed,
				ImagePath:    "/uploads/test.jpg",
				ErrorMessage: "Gemini API タイムアウト",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "failed", response["status"])
	assert.Equal(t, "Gemini API タイムアウト", response["error"])
}

func TestStatusHandler_Completed_GetResultError(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			return &repository.AnalysisRequest{
				ID:        requestID,
				Status:    repository.StatusCompleted,
				ImagePath: "/uploads/test.jpg",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
		GetResultFunc: func(ctx context.Context, userID string, requestID uuid.UUID) (*service.AnalysisResult, error) {
			return nil, assert.AnError
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatusHandler_UnknownStatus(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			return &repository.AnalysisRequest{
				ID:        requestID,
				Status:    repository.AnalysisStatus("unknown"),
				ImagePath: "/uploads/test.jpg",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatusHandler_InvalidURLPath(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}
	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatusHandler_NotFound(t *testing.T) {
	requestID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetRequestFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
			return nil, assert.AnError
		},
	}

	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+requestID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStatusHandler_InvalidUUID(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}
	handler := NewStatusHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/analyze/invalid-uuid", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatusHandler_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewStatusHandler(mockRepo)

	// コンテキストにユーザーIDを設定しない
	req := httptest.NewRequest(http.MethodGet, "/api/analyze/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
