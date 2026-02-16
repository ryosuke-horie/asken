package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeightGoalHandler_HandleGet_WithGoal(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			assert.Equal(t, testUserID, userID)
			return &repository.WeightGoal{
				TargetWeightKg: 63.0,
				UpdatedAt:      now,
			}, nil
		},
	}

	handler := NewWeightGoalHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WeightGoalNullableResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.NotNil(t, response.Goal)
	assert.Equal(t, 63.0, response.Goal.TargetWeightKg)
}

func TestWeightGoalHandler_HandleGet_NoGoal(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return nil, nil
		},
	}

	handler := NewWeightGoalHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Nil(t, response["goal"])
}

func TestWeightGoalHandler_HandleGet_Unauthorized(t *testing.T) {
	mockRepo := &MockWeightGoalRepository{}
	handler := NewWeightGoalHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeightGoalHandler_HandleGet_MethodNotAllowed(t *testing.T) {
	handler := NewWeightGoalHandler(&MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/weight/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWeightGoalHandler_HandleSet_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightGoalRepository{
		SetGoalFunc: func(ctx context.Context, userID string, targetWeightKg float64) (*repository.WeightGoal, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, 63.0, targetWeightKg)
			return &repository.WeightGoal{
				TargetWeightKg: 63.0,
				UpdatedAt:      now,
			}, nil
		},
	}

	handler := NewWeightGoalHandler(mockRepo)

	body := SetWeightGoalRequest{TargetWeightKg: 63.0}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/goal", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WeightGoalResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 63.0, response.TargetWeightKg)
}

func TestWeightGoalHandler_HandleSet_ValidationError(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockWeightGoalRepository{}
	handler := NewWeightGoalHandler(mockRepo)

	tests := []struct {
		name     string
		body     SetWeightGoalRequest
		wantCode int
	}{
		{
			name:     "体重が低すぎ",
			body:     SetWeightGoalRequest{TargetWeightKg: 10.0},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "体重が高すぎ",
			body:     SetWeightGoalRequest{TargetWeightKg: 500.0},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/weight/goal", bytes.NewReader(bodyBytes))
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleSet(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestWeightGoalHandler_HandleSet_InvalidJSON(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockWeightGoalRepository{}
	handler := NewWeightGoalHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/goal", bytes.NewReader([]byte("{invalid json}")))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWeightGoalHandler_HandleSet_Unauthorized(t *testing.T) {
	mockRepo := &MockWeightGoalRepository{}
	handler := NewWeightGoalHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeightGoalHandler_HandleSet_MethodNotAllowed(t *testing.T) {
	handler := NewWeightGoalHandler(&MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/weight/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWeightGoalHandler_HandleGet_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightGoalHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeightGoalHandler_HandleSet_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightGoalRepository{
		SetGoalFunc: func(ctx context.Context, userID string, targetWeightKg float64) (*repository.WeightGoal, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightGoalHandler(mockRepo)

	body := SetWeightGoalRequest{TargetWeightKg: 63.0}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/goal", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
