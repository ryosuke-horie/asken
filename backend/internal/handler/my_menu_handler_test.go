package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

type MockMyMenuRepository struct {
	CreateFunc func(ctx context.Context, userID string, name string, foods []gemini.NutritionInfo) (*repository.MyMenuItem, error)
	ListFunc   func(ctx context.Context, userID string) ([]repository.MyMenuItem, error)
	GetFunc    func(ctx context.Context, userID string, menuID string) (*repository.MyMenuItem, error)
	UpdateFunc func(ctx context.Context, userID string, menuID string, name string, foods []gemini.NutritionInfo) (*repository.MyMenuItem, error)
	DeleteFunc func(ctx context.Context, userID string, menuID string) error
}

func (m *MockMyMenuRepository) Create(ctx context.Context, userID string, name string, foods []gemini.NutritionInfo) (*repository.MyMenuItem, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, name, foods)
	}
	return nil, nil
}

func (m *MockMyMenuRepository) List(ctx context.Context, userID string) ([]repository.MyMenuItem, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockMyMenuRepository) Get(ctx context.Context, userID string, menuID string) (*repository.MyMenuItem, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, userID, menuID)
	}
	return nil, nil
}

func (m *MockMyMenuRepository) Update(ctx context.Context, userID string, menuID string, name string, foods []gemini.NutritionInfo) (*repository.MyMenuItem, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, userID, menuID, name, foods)
	}
	return nil, nil
}

func (m *MockMyMenuRepository) Delete(ctx context.Context, userID string, menuID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, userID, menuID)
	}
	return nil
}

func TestMyMenuHandler_HandleCreate_TotalMicronutrientsInResponse(t *testing.T) {
	micronutrients := map[string]float64{
		"iron_mg":    10.5,
		"calcium_mg": 200.0,
	}

	mockRepo := &MockMyMenuRepository{
		CreateFunc: func(_ context.Context, _ string, name string, foods []gemini.NutritionInfo) (*repository.MyMenuItem, error) {
			return &repository.MyMenuItem{
				ID:                  "test-id",
				Name:                name,
				Foods:               foods,
				TotalCalories:       500,
				TotalProtein:        30,
				TotalFat:            20,
				TotalCarbohydrates:  50,
				TotalMicronutrients: micronutrients,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}, nil
		},
	}

	handler := NewMyMenuHandler(mockRepo, &MockAnalysisRepository{})

	reqBody := `{
		"name": "テストメニュー",
		"foods": [
			{
				"name": "鶏むね肉",
				"estimated_amount": "100g",
				"calories_kcal": 500,
				"protein_g": 30,
				"fat_g": 20,
				"carbohydrates_g": 50,
				"micronutrients": {"iron_mg": 10.5, "calcium_mg": 200.0}
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/my-menu", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp MyMenuResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, micronutrients, resp.TotalMicronutrients)
}

func TestMyMenuHandler_HandleList_TotalMicronutrientsInResponse(t *testing.T) {
	micronutrients := map[string]float64{
		"vitamin_c_mg": 50.0,
		"zinc_mg":      3.5,
	}

	mockRepo := &MockMyMenuRepository{
		ListFunc: func(_ context.Context, _ string) ([]repository.MyMenuItem, error) {
			return []repository.MyMenuItem{
				{
					ID:                  "test-id",
					Name:                "テストメニュー",
					Foods:               []gemini.NutritionInfo{{Name: "りんご", Calories: 100}},
					TotalCalories:       100,
					TotalMicronutrients: micronutrients,
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				},
			}, nil
		},
	}

	handler := NewMyMenuHandler(mockRepo, &MockAnalysisRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/my-menu", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []MyMenuResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
	assert.Equal(t, micronutrients, resp[0].TotalMicronutrients)
}

func TestMyMenuHandler_HandleRecord_TotalMicronutrientsPassedToMealRecord(t *testing.T) {
	menuID := "550e8400-e29b-41d4-a716-446655440000"
	micronutrients := map[string]float64{
		"iron_mg": 5.0,
	}

	var capturedResult *service.AnalysisResult

	mockMenuRepo := &MockMyMenuRepository{
		GetFunc: func(_ context.Context, _, _ string) (*repository.MyMenuItem, error) {
			return &repository.MyMenuItem{
				ID:                  menuID,
				Name:                "テストメニュー",
				Foods:               []gemini.NutritionInfo{{Name: "牛肉", Calories: 300}},
				TotalCalories:       300,
				TotalMicronutrients: micronutrients,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}, nil
		},
	}

	mockAnalysisRepo := &MockAnalysisRepository{
		CreateRequestFromMylistFunc: func(_ context.Context, _ string, _, _ string, _ *string, result *service.AnalysisResult) (uuid.UUID, error) {
			capturedResult = result
			return uuid.New(), nil
		},
	}

	handler := NewMyMenuHandler(mockMenuRepo, mockAnalysisRepo)

	reqBody := `{"meal_type": "lunch", "meal_date": "2024-01-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/my-menu/"+menuID+"/record", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.HandleRecord(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, capturedResult)
	assert.Equal(t, micronutrients, capturedResult.TotalMicronutrients)
}

func TestMyMenuHandler_MethodNotAllowed(t *testing.T) {
	handler := NewMyMenuHandler(&MockMyMenuRepository{}, &MockAnalysisRepository{})
	menuID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name   string
		method string
		url    string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{
			name:   "HandleList",
			method: http.MethodPost,
			url:    "/api/my-menu",
			handle: handler.HandleList,
		},
		{
			name:   "HandleCreate",
			method: http.MethodGet,
			url:    "/api/my-menu",
			handle: handler.HandleCreate,
		},
		{
			name:   "HandleGet",
			method: http.MethodPost,
			url:    "/api/my-menu/" + menuID,
			handle: handler.HandleGet,
		},
		{
			name:   "HandleUpdate",
			method: http.MethodGet,
			url:    "/api/my-menu/" + menuID,
			handle: handler.HandleUpdate,
		},
		{
			name:   "HandleDelete",
			method: http.MethodGet,
			url:    "/api/my-menu/" + menuID,
			handle: handler.HandleDelete,
		},
		{
			name:   "HandleRecord",
			method: http.MethodGet,
			url:    "/api/my-menu/" + menuID + "/record",
			handle: handler.HandleRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			tt.handle(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}
