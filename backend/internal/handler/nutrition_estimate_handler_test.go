package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNutritionEstimator はNutritionEstimatorのモック実装
type mockNutritionEstimator struct {
	result *gemini.NutritionInfo
	err    error
}

func (m *mockNutritionEstimator) EstimateSingleFood(_ context.Context, _ string, _ int) (*gemini.NutritionInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestNutritionEstimateHandler_Success(t *testing.T) {
	mock := &mockNutritionEstimator{
		result: &gemini.NutritionInfo{
			Name:            "味噌ラーメン",
			EstimatedAmount: "1杯",
			Calories:        500,
			Protein:         20,
			Fat:             15,
			Carbohydrates:   60,
		},
	}
	handler := NewNutritionEstimateHandler(mock)

	body := `{"food_name": "味噌ラーメン", "quantity": 1}`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response NutritionEstimateResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "味噌ラーメン", response.Name)
	assert.Equal(t, "1杯", response.EstimatedAmount)
	assert.Equal(t, 500.0, response.CaloriesKcal)
	assert.Equal(t, 20.0, response.ProteinG)
	assert.Equal(t, 15.0, response.FatG)
	assert.Equal(t, 60.0, response.CarbohydratesG)
}

func TestNutritionEstimateHandler_MethodNotAllowed(t *testing.T) {
	mock := &mockNutritionEstimator{}
	handler := NewNutritionEstimateHandler(mock)
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/nutrition/estimate", nil)
			w := httptest.NewRecorder()

			handler.Handle(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestNutritionEstimateHandler_EmptyFoodName(t *testing.T) {
	mock := &mockNutritionEstimator{}
	handler := NewNutritionEstimateHandler(mock)

	body := `{"food_name": "", "quantity": 1}`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNutritionEstimateHandler_TooLongFoodName(t *testing.T) {
	mock := &mockNutritionEstimator{}
	handler := NewNutritionEstimateHandler(mock)

	// 201文字の食品名
	longName := make([]byte, 201)
	for i := range longName {
		longName[i] = 'a'
	}
	body := `{"food_name": "` + string(longName) + `", "quantity": 1}`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNutritionEstimateHandler_TooLargeQuantity(t *testing.T) {
	mock := &mockNutritionEstimator{}
	handler := NewNutritionEstimateHandler(mock)

	body := `{"food_name": "ラーメン", "quantity": 101}`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNutritionEstimateHandler_DefaultQuantity(t *testing.T) {
	mock := &mockNutritionEstimator{
		result: &gemini.NutritionInfo{
			Name:            "ラーメン",
			EstimatedAmount: "1杯",
			Calories:        500,
			Protein:         20,
			Fat:             15,
			Carbohydrates:   60,
		},
	}
	handler := NewNutritionEstimateHandler(mock)

	// quantityを指定しない場合、デフォルト値1が使用される
	body := `{"food_name": "ラーメン"}`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNutritionEstimateHandler_InvalidJSON(t *testing.T) {
	mock := &mockNutritionEstimator{}
	handler := NewNutritionEstimateHandler(mock)

	body := `invalid json`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNutritionEstimateHandler_EstimatorError(t *testing.T) {
	mock := &mockNutritionEstimator{
		err: errors.New("API error"),
	}
	handler := NewNutritionEstimateHandler(mock)

	body := `{"food_name": "ラーメン", "quantity": 1}`
	req := httptest.NewRequest(http.MethodPost, "/api/nutrition/estimate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
