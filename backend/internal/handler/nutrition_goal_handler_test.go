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

// MockNutritionGoalRepository はNutritionGoalRepository用テストモック
type MockNutritionGoalRepository struct {
	GetGoalFunc func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error)
	SetGoalFunc func(ctx context.Context, userID string, targetCalories float64) (*repository.NutritionGoal, error)
}

func (m *MockNutritionGoalRepository) GetGoal(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
	if m.GetGoalFunc != nil {
		return m.GetGoalFunc(ctx, userID, currentWeightKg, targetWeightKg)
	}
	return nil, nil
}

func (m *MockNutritionGoalRepository) SetGoal(ctx context.Context, userID string, targetCalories float64) (*repository.NutritionGoal, error) {
	if m.SetGoalFunc != nil {
		return m.SetGoalFunc(ctx, userID, targetCalories)
	}
	return nil, nil
}

func newTestNutritionGoalHandler(nutritionRepo *MockNutritionGoalRepository, weightGoalRepo *MockWeightGoalRepository) *NutritionGoalHandler {
	return NewNutritionGoalHandler(nutritionRepo, weightGoalRepo)
}

func TestNutritionGoalHandler_HandleGet_WithGoal(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			assert.Equal(t, testUserID, userID)
			return &repository.NutritionGoal{
				TargetCalories:      2000.0,
				TargetProtein:       100.0,
				TargetFat:           55.6,
				TargetCarbohydrates: 275.0,
				Phase:               repository.NutritionPhaseMaintenance,
				UpdatedAt:           now,
			}, nil
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response NutritionGoalNullableResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.NotNil(t, response.Goal)
	assert.Equal(t, 2000.0, response.Goal.TargetCalories)
	assert.Equal(t, 100.0, response.Goal.TargetProtein)
	assert.Equal(t, "maintenance", response.Goal.Phase)
}

func TestNutritionGoalHandler_HandleGet_NoGoal(t *testing.T) {
	testUserID := "test-user-123"

	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			return nil, nil
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal", nil)
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

func TestNutritionGoalHandler_HandleGet_Unauthorized(t *testing.T) {
	nutritionRepo := &MockNutritionGoalRepository{}
	weightGoalRepo := &MockWeightGoalRepository{}
	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNutritionGoalHandler_HandleGet_MethodNotAllowed(t *testing.T) {
	handler := newTestNutritionGoalHandler(&MockNutritionGoalRepository{}, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestNutritionGoalHandler_HandleGet_WithCurrentWeight(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			require.NotNil(t, currentWeightKg, "currentWeightKgはnilでないべき")
			assert.Equal(t, 75.0, *currentWeightKg)
			require.NotNil(t, targetWeightKg, "targetWeightKgはnilでないべき")
			assert.Equal(t, 65.0, *targetWeightKg)
			return &repository.NutritionGoal{
				TargetCalories:      2000.0,
				TargetProtein:       150.0,
				TargetFat:           44.4,
				TargetCarbohydrates: 250.0,
				Phase:               repository.NutritionPhaseWeightLoss,
				UpdatedAt:           now,
			}, nil
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal?current_weight=75.0&target_weight=65.0", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response NutritionGoalNullableResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.NotNil(t, response.Goal)
	assert.Equal(t, "weight_loss", response.Goal.Phase)
}

func TestNutritionGoalHandler_HandleGet_WithoutCurrentWeight(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			assert.Nil(t, currentWeightKg, "currentWeightKgはnilであるべき")
			return &repository.NutritionGoal{
				TargetCalories:      2000.0,
				TargetProtein:       100.0,
				TargetFat:           55.6,
				TargetCarbohydrates: 275.0,
				Phase:               repository.NutritionPhaseMaintenance,
				UpdatedAt:           now,
			}, nil
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNutritionGoalHandler_HandleGet_InvalidCurrentWeight(t *testing.T) {
	testUserID := "test-user-123"
	nutritionRepo := &MockNutritionGoalRepository{}
	weightGoalRepo := &MockWeightGoalRepository{}
	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal?current_weight=invalid", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNutritionGoalHandler_HandleGet_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			return nil, fmt.Errorf("database error")
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNutritionGoalHandler_HandleGet_WeightGoalFallback(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			require.NotNil(t, targetWeightKg, "targetWeightKgはnilでないべき")
			assert.Equal(t, 65.0, *targetWeightKg, "体重目標リポジトリから取得した値が使用されるべき")
			return &repository.NutritionGoal{
				TargetCalories:      2000.0,
				TargetProtein:       100.0,
				TargetFat:           55.6,
				TargetCarbohydrates: 275.0,
				Phase:               repository.NutritionPhaseMaintenance,
				UpdatedAt:           now,
			}, nil
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return &repository.WeightGoal{
				TargetWeightKg: 65.0,
				UpdatedAt:      now,
			}, nil
		},
	}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	// target_weightを指定しない→リポジトリから取得
	req := httptest.NewRequest(http.MethodGet, "/api/nutrition/goal", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNutritionGoalHandler_HandleSet_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	nutritionRepo := &MockNutritionGoalRepository{
		SetGoalFunc: func(ctx context.Context, userID string, targetCalories float64) (*repository.NutritionGoal, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, 2000.0, targetCalories)
			return &repository.NutritionGoal{
				TargetCalories:      2000.0,
				TargetProtein:       100.0,
				TargetFat:           55.6,
				TargetCarbohydrates: 275.0,
				Phase:               repository.NutritionPhaseMaintenance,
				UpdatedAt:           now,
			}, nil
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	body := SetNutritionGoalRequest{TargetCalories: 2000.0}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/nutrition/goal", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response NutritionGoalResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 2000.0, response.TargetCalories)
	assert.Equal(t, 100.0, response.TargetProtein)
}

func TestNutritionGoalHandler_HandleSet_Unauthorized(t *testing.T) {
	nutritionRepo := &MockNutritionGoalRepository{}
	weightGoalRepo := &MockWeightGoalRepository{}
	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodPut, "/api/nutrition/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNutritionGoalHandler_HandleSet_MethodNotAllowed(t *testing.T) {
	handler := newTestNutritionGoalHandler(&MockNutritionGoalRepository{}, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/goal", nil)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestNutritionGoalHandler_HandleSet_InvalidJSON(t *testing.T) {
	testUserID := "test-user-123"
	nutritionRepo := &MockNutritionGoalRepository{}
	weightGoalRepo := &MockWeightGoalRepository{}
	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	req := httptest.NewRequest(http.MethodPut, "/api/nutrition/goal", bytes.NewReader([]byte("{invalid json}")))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNutritionGoalHandler_HandleSet_ValidationError(t *testing.T) {
	testUserID := "test-user-123"
	nutritionRepo := &MockNutritionGoalRepository{}
	weightGoalRepo := &MockWeightGoalRepository{}
	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	tests := []struct {
		name     string
		calories float64
	}{
		{"カロリーが低すぎ", 500.0},
		{"カロリーが高すぎ", 6000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := SetNutritionGoalRequest{TargetCalories: tt.calories}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPut, "/api/nutrition/goal", bytes.NewReader(bodyBytes))
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleSet(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestNutritionGoalHandler_HandleSet_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	nutritionRepo := &MockNutritionGoalRepository{
		SetGoalFunc: func(ctx context.Context, userID string, targetCalories float64) (*repository.NutritionGoal, error) {
			return nil, fmt.Errorf("database error")
		},
	}
	weightGoalRepo := &MockWeightGoalRepository{}

	handler := newTestNutritionGoalHandler(nutritionRepo, weightGoalRepo)

	body := SetNutritionGoalRequest{TargetCalories: 2000.0}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/nutrition/goal", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleSet(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
