package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryHandler_HandleList_Success(t *testing.T) {
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	id1 := uuid.New()
	id2 := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, userID string, page, limit int) ([]repository.HistoryItem, int, error) {
			assert.Equal(t, testUserID, userID)
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

	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
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
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, userID string, page, limit int) ([]repository.HistoryItem, int, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, 2, page)
			assert.Equal(t, 10, limit)
			return []repository.HistoryItem{}, 50, nil
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history?page=2&limit=10", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
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
	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/history", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHistoryHandler_HandleList_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, userID string, page, limit int) ([]repository.HistoryItem, int, error) {
			return nil, 0, assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleList_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	// コンテキストにユーザーIDを設定しない
	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHistoryHandler_HandleDetail_Success(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			assert.Equal(t, testUserID, userID)
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

	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
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
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, fmt.Errorf("履歴が見つかりません: %s: %w", id, repository.ErrNotFound)
		},
	}

	handler := NewHistoryHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryHandler_HandleDetail_RepositoryError(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleDetail_InvalidUUID(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history/invalid-uuid", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHistoryHandler_HandleDetail_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/history/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHistoryHandler_HandleDetail_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	// コンテキストにユーザーIDを設定しない
	req := httptest.NewRequest(http.MethodGet, "/api/history/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHistoryHandler_HandleUpdate_Success(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, historyID, id)
			assert.Len(t, foods, 1)
			assert.Equal(t, "白米", foods[0].Name)
			return nil
		},
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			assert.Equal(t, testUserID, userID)
			return &repository.HistoryDetail{
				HistoryItem: repository.HistoryItem{
					ID:                 historyID,
					ImagePath:          "/uploads/test.jpg",
					CreatedAt:          createdAt,
					MealType:           "lunch",
					MealDate:           mealDate,
					TotalCalories:      252.0,
					TotalProtein:       3.8,
					TotalFat:           0.5,
					TotalCarbohydrates: 55.7,
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

	handler := NewHistoryHandler(mockRepo, nil)

	body := `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response repository.HistoryDetail
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, historyID, response.ID)
	assert.Len(t, response.Foods, 1)
}

func TestHistoryHandler_HandleUpdate_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	// コンテキストにユーザーIDを設定しない
	body := `{"foods":[]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHistoryHandler_HandleUpdate_InvalidUUID(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	body := `{"foods":[]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/invalid-uuid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHistoryHandler_HandleUpdate_NotFound(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, errors.New("履歴が見つかりません")
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)

	body := `{"foods":[]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryHandler_HandleUpdate_RepositoryError(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo)

	body := `{"foods":[]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleUpdate_GetDetailError(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return nil
		},
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo)

	body := `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleUpdate_InvalidBody(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHistoryHandler_HandleUpdate_OversizedBody(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo)

	// 1MBを超えるJSONボディを作成
	largeFoods := make([]map[string]interface{}, 0)
	for i := 0; i < 10000; i++ {
		largeFoods = append(largeFoods, map[string]interface{}{
			"name":             strings.Repeat("あ", 100),
			"estimated_amount": "100g",
			"calories_kcal":    100.0,
			"protein_g":        10.0,
			"fat_g":            5.0,
			"carbohydrates_g":  20.0,
		})
	}
	reqBody := map[string]interface{}{
		"foods": largeFoods,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)
	require.Greater(t, len(body), 1<<20, "テストボディが1MBを超えていること")

	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), "リクエストボディが大きすぎます")
}

func TestHistoryHandler_HandleUpdate_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/history/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDetectNameChanges(t *testing.T) {
	tests := []struct {
		name     string
		oldFoods []gemini.NutritionInfo
		newFoods []gemini.NutritionInfo
		expected []int
	}{
		{
			name: "名前変更なし",
			oldFoods: []gemini.NutritionInfo{
				{Name: "白米"},
				{Name: "味噌汁"},
			},
			newFoods: []gemini.NutritionInfo{
				{Name: "白米"},
				{Name: "味噌汁"},
			},
			expected: nil,
		},
		{
			name: "1つの名前が変更",
			oldFoods: []gemini.NutritionInfo{
				{Name: "白米"},
				{Name: "味噌汁"},
			},
			newFoods: []gemini.NutritionInfo{
				{Name: "玄米"},
				{Name: "味噌汁"},
			},
			expected: []int{0},
		},
		{
			name: "複数の名前が変更",
			oldFoods: []gemini.NutritionInfo{
				{Name: "白米"},
				{Name: "味噌汁"},
				{Name: "焼き魚"},
			},
			newFoods: []gemini.NutritionInfo{
				{Name: "玄米"},
				{Name: "味噌汁"},
				{Name: "刺身"},
			},
			expected: []int{0, 2},
		},
		{
			name:     "新しい食材が追加された場合（旧より多い）",
			oldFoods: []gemini.NutritionInfo{{Name: "白米"}},
			newFoods: []gemini.NutritionInfo{{Name: "白米"}, {Name: "味噌汁"}},
			expected: nil,
		},
		{
			name:     "食材が削除された場合（旧より少ない）",
			oldFoods: []gemini.NutritionInfo{{Name: "白米"}, {Name: "味噌汁"}},
			newFoods: []gemini.NutritionInfo{{Name: "白米"}},
			expected: nil,
		},
		{
			name:     "両方とも空",
			oldFoods: []gemini.NutritionInfo{},
			newFoods: []gemini.NutritionInfo{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectNameChanges(tt.oldFoods, tt.newFoods)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// MockNutritionRecalculator はテスト用のNutritionRecalculatorモック
type MockNutritionRecalculator struct {
	CalculateNutritionFunc func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error)
	called                 atomic.Bool
}

func (m *MockNutritionRecalculator) CalculateNutrition(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
	m.called.Store(true)
	if m.CalculateNutritionFunc != nil {
		return m.CalculateNutritionFunc(ctx, foods)
	}
	return nil, nil
}

func TestHistoryHandler_HandleUpdate_WithNameChange(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	recalcDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			defer close(recalcDone)
			assert.Len(t, foods, 1)
			assert.Equal(t, "玄米", foods[0].Name)
			return []gemini.NutritionInfo{
				{Name: "玄米", EstimatedAmount: "150g", Calories: 228, Protein: 4.2, Fat: 1.8, Carbohydrates: 47.8},
			}, nil
		},
	}

	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return &repository.HistoryDetail{
				HistoryItem: repository.HistoryItem{
					ID:                 historyID,
					ImagePath:          "/uploads/test.jpg",
					CreatedAt:          createdAt,
					MealType:           "lunch",
					MealDate:           mealDate,
					TotalCalories:      252.0,
					TotalProtein:       3.8,
					TotalFat:           0.5,
					TotalCarbohydrates: 55.7,
				},
				Foods: []gemini.NutritionInfo{
					{Name: "白米", EstimatedAmount: "150g", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
				},
			}, nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			updateCallCount.Add(1)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)

	body := `{"foods":[{"name":"玄米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 非同期再計算の完了を待つ
	select {
	case <-recalcDone:
		assert.True(t, mockRecalculator.called.Load())
		// UpdateResultは2回呼ばれる（同期保存 + 非同期再計算保存）
		assert.Equal(t, int32(2), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_NoNameChange(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRecalculator := &MockNutritionRecalculator{}

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return &repository.HistoryDetail{
				HistoryItem: repository.HistoryItem{
					ID:                 historyID,
					ImagePath:          "/uploads/test.jpg",
					CreatedAt:          createdAt,
					MealType:           "lunch",
					MealDate:           mealDate,
					TotalCalories:      252.0,
					TotalProtein:       3.8,
					TotalFat:           0.5,
					TotalCarbohydrates: 55.7,
				},
				Foods: []gemini.NutritionInfo{
					{Name: "白米", EstimatedAmount: "150g", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
				},
			}, nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)

	body := `{"foods":[{"name":"白米","estimated_amount":"200g","calories_kcal":336,"protein_g":5.1,"fat_g":0.7,"carbohydrates_g":74.3}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 少し待って、非同期再計算が呼ばれないことを確認
	time.Sleep(100 * time.Millisecond)
	assert.False(t, mockRecalculator.called.Load())
}
