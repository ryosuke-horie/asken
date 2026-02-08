package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkipMealHandler_Handle_Success(t *testing.T) {
	expectedID := uuid.New()
	userID := "test-firebase-uid"

	mockRepo := &MockAnalysisRepository{
		CreateSkippedMealFunc: func(ctx context.Context, mealType string, mealDate string, uid *string) (uuid.UUID, error) {
			assert.Equal(t, "lunch", mealType)
			assert.Equal(t, "2026-01-24", mealDate)
			assert.Equal(t, userID, *uid)
			return expectedID, nil
		},
	}

	handler := NewSkipMealHandler(mockRepo)

	reqBody := SkipMealRequest{
		MealType: "lunch",
		MealDate: "2026-01-24",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response SkipMealResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedID.String(), response.ID)
}

func TestSkipMealHandler_Handle_MissingMealType(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewSkipMealHandler(mockRepo)

	reqBody := SkipMealRequest{
		MealDate: "2026-01-24",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	userID := "test-firebase-uid"
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSkipMealHandler_Handle_InvalidMealType(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewSkipMealHandler(mockRepo)

	reqBody := SkipMealRequest{
		MealType: "invalid",
		MealDate: "2026-01-24",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	userID := "test-firebase-uid"
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSkipMealHandler_Handle_MissingMealDate(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewSkipMealHandler(mockRepo)

	reqBody := SkipMealRequest{
		MealType: "lunch",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	userID := "test-firebase-uid"
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSkipMealHandler_Handle_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewSkipMealHandler(mockRepo)

	reqBody := SkipMealRequest{
		MealType: "lunch",
		MealDate: "2026-01-24",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSkipMealHandler_Handle_RepositoryError(t *testing.T) {
	userID := "test-firebase-uid"

	mockRepo := &MockAnalysisRepository{
		CreateSkippedMealFunc: func(ctx context.Context, mealType string, mealDate string, uid *string) (uuid.UUID, error) {
			return uuid.Nil, assert.AnError
		},
	}

	handler := NewSkipMealHandler(mockRepo)

	reqBody := SkipMealRequest{
		MealType: "lunch",
		MealDate: "2026-01-24",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSkipMealHandler_Handle_OversizedBody(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewSkipMealHandler(mockRepo)

	// 1KBを超えるJSONボディを作成
	largeBody := map[string]string{
		"meal_type": "lunch",
		"meal_date": "2026-01-24",
		"padding":   string(make([]byte, 2000)),
	}
	body, _ := json.Marshal(largeBody)

	req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	userID := "test-firebase-uid"
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSkipMealHandler_Handle_AllMealTypes(t *testing.T) {
	mealTypes := []string{"breakfast", "lunch", "dinner", "snack"}
	userID := "test-firebase-uid"

	for _, mealType := range mealTypes {
		t.Run(mealType, func(t *testing.T) {
			expectedID := uuid.New()

			mockRepo := &MockAnalysisRepository{
				CreateSkippedMealFunc: func(ctx context.Context, mt string, mealDate string, uid *string) (uuid.UUID, error) {
					assert.Equal(t, mealType, mt)
					return expectedID, nil
				},
			}

			handler := NewSkipMealHandler(mockRepo)

			reqBody := SkipMealRequest{
				MealType: mealType,
				MealDate: "2026-01-24",
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/meals/skip", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.Handle(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}
