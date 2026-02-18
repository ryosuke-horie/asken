package handler

import (
	"context"
	"encoding/json"
	"errors"
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

// --- テストヘルパー ---

// makeHistoryDetail はテスト用のHistoryDetailを生成する。
// foodNameとamountで1食材の詳細を構築する。栄養素値はcals/prot/fat/carbで指定する。
func makeHistoryDetail(id uuid.UUID, createdAt time.Time, mealDate time.Time, foodName, amount string, cals, prot, fat, carb float64) *repository.HistoryDetail {
	return &repository.HistoryDetail{
		HistoryItem: repository.HistoryItem{
			ID: id, ImagePath: "/uploads/test.jpg", CreatedAt: createdAt,
			MealType: "lunch", MealDate: mealDate,
			TotalCalories: cals, TotalProtein: prot, TotalFat: fat, TotalCarbohydrates: carb,
		},
		Foods: []gemini.NutritionInfo{
			{Name: foodName, EstimatedAmount: amount, Calories: cals, Protein: prot, Fat: fat, Carbohydrates: carb},
		},
	}
}

// whiteRiceDetail は「白米 150g」のHistoryDetailを返す（旧データとして頻出）
func whiteRiceDetail(id uuid.UUID, createdAt time.Time, mealDate time.Time) *repository.HistoryDetail {
	return makeHistoryDetail(id, createdAt, mealDate, "白米", "150g", 252.0, 3.8, 0.5, 55.7)
}

// brownRiceDetail は「玄米 150g」のHistoryDetailを返す（更新後データとして頻出）
func brownRiceDetail(id uuid.UUID, createdAt time.Time, mealDate time.Time) *repository.HistoryDetail {
	return makeHistoryDetail(id, createdAt, mealDate, "玄米", "150g", 252.0, 3.8, 0.5, 55.7)
}

// brownRiceUpdateBody は「玄米 150g」への更新リクエストJSON
const brownRiceUpdateBody = `{"foods":[{"name":"玄米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`

// whiteRiceUpdateBody は「白米 150g」の更新リクエストJSON（名前変更なしのケース）
const whiteRiceUpdateBody = `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`

// recalculatedBrownRice はGemini再計算結果の「玄米」データ
func recalculatedBrownRice() []gemini.NutritionInfo {
	return []gemini.NutritionInfo{
		{Name: "玄米", EstimatedAmount: "150g", Calories: 228, Protein: 4.2, Fat: 1.8, Carbohydrates: 47.8},
	}
}

// newUpdateRequest はPUT /api/history/:id のHTTPリクエストを構築する
func newUpdateRequest(historyID uuid.UUID, userID string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	req = req.WithContext(ctx)
	return req, httptest.NewRecorder()
}

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
			return nil, fmt.Errorf("document not found: %w", repository.ErrNotFound)
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDetail(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryHandler_HandleDetail_InternalError(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, errors.New("internal error")
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
			return nil, fmt.Errorf("履歴が見つかりません: %s: %w", id, repository.ErrNotFound)
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

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryHandler_HandleUpdate_RepositoryError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return assert.AnError
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)
	req, w := newUpdateRequest(historyID, testUserID, whiteRiceUpdateBody)

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

	handler := NewHistoryHandler(mockRepo, nil)

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
	handler := NewHistoryHandler(mockRepo, nil)

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

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			assert.Len(t, foods, 1)
			assert.Equal(t, "玄米", foods[0].Name)
			return []gemini.NutritionInfo{
				{Name: "玄米", EstimatedAmount: "150g", Calories: 228, Protein: 4.2, Fat: 1.8, Carbohydrates: 47.8},
			}, nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, historyID, id)
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				// 1回目: 更新前の旧データ取得
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
			}
			// 2回目以降: 更新後データ（レスポンス用、鮮度チェック用）
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
					{Name: "玄米", EstimatedAmount: "150g", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
				},
			}, nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, historyID, id)
			callNum := updateCallCount.Add(1)
			if callNum == 2 {
				// 2回目: 非同期再計算後の保存。再計算済み栄養素値を検証
				assert.Len(t, foods, 1)
				assert.Equal(t, "玄米", foods[0].Name)
				assert.Equal(t, 228.0, foods[0].Calories)
				assert.Equal(t, 4.2, foods[0].Protein)
				assert.Equal(t, 1.8, foods[0].Fat)
				assert.Equal(t, 47.8, foods[0].Carbohydrates)
				defer close(asyncDone)
			}
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

	// 非同期再計算の完了を待つ（UpdateResult 2回目完了をシグナル）
	select {
	case <-asyncDone:
		assert.True(t, mockRecalculator.called.Load())
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

func TestHistoryHandler_HandleUpdate_RecalculateAsyncGeminiError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	var recalcCallCount atomic.Int32
	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			callNum := recalcCallCount.Add(1)
			if callNum == 2 {
				defer close(asyncDone)
			}
			return nil, errors.New("Gemini API error")
		},
	}

	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			updateCallCount.Add(1)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(1), updateCallCount.Load())
		assert.Equal(t, int32(2), recalcCallCount.Load())
	case <-time.After(10 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_GetHistoryDetailInternalError(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return nil, errors.New("internal error")
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)
	req, w := newUpdateRequest(historyID, testUserID, whiteRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleUpdate_StaleDataSkipped(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			if callNum == 2 {
				return brownRiceDetail(historyID, createdAt, mealDate), nil
			}
			// 3回目: 鮮度チェック時にデータが変わっている（ユーザーが再保存した状態を模倣）
			defer close(asyncDone)
			return makeHistoryDetail(historyID, createdAt, mealDate, "玄米おにぎり", "200g", 300.0, 5.0, 1.0, 60.0), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			updateCallCount.Add(1)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(1), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "空の名前",
			body: `{"foods":[{"name":"","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`,
		},
		{
			name: "スペースのみの名前",
			body: `{"foods":[{"name":"   ","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`,
		},
		{
			name: "空のEstimatedAmount",
			body: `{"foods":[{"name":"白米","estimated_amount":"","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`,
		},
		{
			name: "空のFoodsリスト",
			body: `{"foods":[]}`,
		},
		{
			name: "負のカロリー",
			body: `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":-100,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`,
		},
		{
			name: "負のたんぱく質",
			body: `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":-3.8,"fat_g":0.5,"carbohydrates_g":55.7}]}`,
		},
		{
			name: "負の脂質",
			body: `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":-0.5,"carbohydrates_g":55.7}]}`,
		},
		{
			name: "負の炭水化物",
			body: `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":-55.7}]}`,
		},
		{
			name: "複数食材の2番目が無効",
			body: `{"foods":[{"name":"白米","estimated_amount":"150g","calories_kcal":252,"protein_g":3.8,"fat_g":0.5,"carbohydrates_g":55.7},{"name":"","estimated_amount":"200ml","calories_kcal":30,"protein_g":2.0,"fat_g":0.5,"carbohydrates_g":4.0}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			historyID := uuid.New()
			testUserID := "test-user-123"

			mockRepo := &MockAnalysisRepository{}
			handler := NewHistoryHandler(mockRepo, nil)

			req := httptest.NewRequest(http.MethodPut, "/api/history/"+historyID.String(), strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleUpdate(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Validation error:")
		})
	}
}

func TestHistoryHandler_HandleUpdate_StalenessCheckGetDetailError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			if callNum == 2 {
				return brownRiceDetail(historyID, createdAt, mealDate), nil
			}
			// 3回目: 鮮度チェック時にエラー
			defer close(asyncDone)
			return nil, errors.New("firestore connection error")
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			updateCallCount.Add(1)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(1), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_AsyncUpdateResultError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			return brownRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 1 {
				return nil
			}
			if callNum == 3 {
				defer close(asyncDone)
			}
			return errors.New("firestore write error")
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(3), updateCallCount.Load())
	case <-time.After(10 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_UpdateResultNotFound(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return fmt.Errorf("document deleted: %w", repository.ErrNotFound)
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)
	req, w := newUpdateRequest(historyID, testUserID, whiteRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryHandler_HandleUpdate_UpdateResultInternalError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return errors.New("firestore internal error")
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)
	req, w := newUpdateRequest(historyID, testUserID, whiteRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryHandler_HandleUpdate_GeminiRetrySuccess(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	var recalcCallCount atomic.Int32
	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			callNum := recalcCallCount.Add(1)
			if callNum == 1 {
				return nil, errors.New("temporary Gemini API error")
			}
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			return brownRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 2 {
				assert.Len(t, foods, 1)
				assert.Equal(t, "玄米", foods[0].Name)
				assert.Equal(t, 228.0, foods[0].Calories)
				defer close(asyncDone)
			}
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(2), recalcCallCount.Load())
		assert.Equal(t, int32(2), updateCallCount.Load())
	case <-time.After(10 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestFoodsMatch(t *testing.T) {
	tests := []struct {
		name     string
		a        []gemini.NutritionInfo
		b        []gemini.NutritionInfo
		expected bool
	}{
		{
			name: "一致",
			a: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "150g"},
				{Name: "味噌汁", EstimatedAmount: "200ml"},
			},
			b: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "150g"},
				{Name: "味噌汁", EstimatedAmount: "200ml"},
			},
			expected: true,
		},
		{
			name: "名前不一致",
			a: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "150g"},
			},
			b: []gemini.NutritionInfo{
				{Name: "玄米", EstimatedAmount: "150g"},
			},
			expected: false,
		},
		{
			name: "量不一致",
			a: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "150g"},
			},
			b: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "200g"},
			},
			expected: false,
		},
		{
			name:     "長さ不一致",
			a:        []gemini.NutritionInfo{{Name: "白米", EstimatedAmount: "150g"}},
			b:        []gemini.NutritionInfo{{Name: "白米", EstimatedAmount: "150g"}, {Name: "味噌汁", EstimatedAmount: "200ml"}},
			expected: false,
		},
		{
			name:     "両方空",
			a:        []gemini.NutritionInfo{},
			b:        []gemini.NutritionInfo{},
			expected: true,
		},
		{
			name:     "両方nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name: "栄養素値のみ異なる場合は一致とみなす",
			a: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "150g", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
			},
			b: []gemini.NutritionInfo{
				{Name: "白米", EstimatedAmount: "150g", Calories: 300, Protein: 5.0, Fat: 1.0, Carbohydrates: 60.0},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := foodsMatch(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHistoryHandler_HandleUpdate_FirestoreRetryStaleData(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			if callNum <= 3 {
				return brownRiceDetail(historyID, createdAt, mealDate), nil
			}
			// 4回目: Firestoreリトライ前の鮮度再チェック -> データが変わっている
			defer close(asyncDone)
			return makeHistoryDetail(historyID, createdAt, mealDate, "玄米おにぎり", "200g", 400.0, 8.0, 2.0, 70.0), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 1 {
				return nil
			}
			return errors.New("temporary firestore error")
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(2), updateCallCount.Load())
	case <-time.After(10 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_FirestoreRetryRecheckError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			if callNum <= 3 {
				return brownRiceDetail(historyID, createdAt, mealDate), nil
			}
			// 4回目: Firestoreリトライ前の鮮度再チェック -> エラー
			defer close(asyncDone)
			return nil, errors.New("firestore connection error during recheck")
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 1 {
				return nil
			}
			return errors.New("temporary firestore error")
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(2), updateCallCount.Load())
	case <-time.After(10 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_PostUpdateGetDetailError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	var getDetailCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			return nil, errors.New("firestore read error after update")
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, nil)
	req, w := newUpdateRequest(historyID, testUserID, whiteRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get updated history")
}

func TestHistoryHandler_HandleUpdate_NonRetryableFirestoreError(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			return brownRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 1 {
				return nil
			}
			defer close(asyncDone)
			return fmt.Errorf("document deleted: %w", repository.ErrNotFound)
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		assert.Equal(t, int32(2), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestUpdateFoodItem_Validate_ZeroNutritionValues(t *testing.T) {
	// ゼロ値の栄養素はバリデーションを通過する（設計上の選択）
	item := UpdateFoodItem{
		Name:            "水",
		EstimatedAmount: "500ml",
		Calories:        0,
		Protein:         0,
		Fat:             0,
		Carbohydrates:   0,
	}
	assert.NoError(t, item.Validate())
}

func TestHistoryHandler_HandleUpdate_NonRetryableGeminiContextCanceled(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			defer close(asyncDone)
			return nil, fmt.Errorf("request canceled: %w", context.Canceled)
		},
	}

	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			updateCallCount.Add(1)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		// context.Canceledは非リトライのため即終了。UpdateResultは同期保存の1回のみ
		assert.Equal(t, int32(1), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_NonRetryableGeminiDeadlineExceeded(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			defer close(asyncDone)
			return nil, fmt.Errorf("timeout: %w", context.DeadlineExceeded)
		},
	}

	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			updateCallCount.Add(1)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		// context.DeadlineExceededは非リトライのため即終了
		assert.Equal(t, int32(1), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_NonRetryableFirestoreContextCanceled(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			return brownRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 1 {
				return nil
			}
			// 非同期保存でcontext.Canceledエラー（リトライ不要）
			defer close(asyncDone)
			return fmt.Errorf("context canceled: %w", context.Canceled)
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		// context.Canceledは非リトライのため即終了
		assert.Equal(t, int32(2), updateCallCount.Load())
	case <-time.After(5 * time.Second):
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_FirestoreRetrySuccess(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	asyncDone := make(chan struct{})

	mockRecalculator := &MockNutritionRecalculator{
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return recalculatedBrownRice(), nil
		},
	}

	var getDetailCallCount atomic.Int32
	var updateCallCount atomic.Int32
	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			callNum := getDetailCallCount.Add(1)
			if callNum == 1 {
				return whiteRiceDetail(historyID, createdAt, mealDate), nil
			}
			return brownRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			callNum := updateCallCount.Add(1)
			if callNum == 1 {
				return nil
			}
			if callNum == 2 {
				// 2回目: 非同期保存で一時エラー → リトライトリガー
				return errors.New("temporary firestore error")
			}
			// 3回目: リトライ成功。再計算済み栄養素値を検証
			assert.Len(t, foods, 1)
			assert.Equal(t, "玄米", foods[0].Name)
			assert.Equal(t, 228.0, foods[0].Calories)
			assert.Equal(t, 4.2, foods[0].Protein)
			defer close(asyncDone)
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)
	req, w := newUpdateRequest(historyID, testUserID, brownRiceUpdateBody)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	select {
	case <-asyncDone:
		// 同期保存(1) + 非同期失敗(2) + リトライ成功(3) = 3回
		assert.Equal(t, int32(3), updateCallCount.Load())
	case <-time.After(10 * time.Second):
		// リトライ間にfirestoreRetryDelay(1s)の待機があるため、余裕を持った10秒タイムアウト
		t.Fatal("非同期再計算がタイムアウト")
	}
}

func TestHistoryHandler_HandleUpdate_TooManyFoods(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryHandler(mockRepo, nil)

	// 51個の食材を持つリクエストを作成
	foods := make([]map[string]interface{}, 51)
	for i := 0; i < 51; i++ {
		foods[i] = map[string]interface{}{
			"name":             fmt.Sprintf("食材%d", i+1),
			"estimated_amount": "100g",
			"calories_kcal":    100.0,
			"protein_g":        10.0,
			"fat_g":            5.0,
			"carbohydrates_g":  20.0,
		}
	}
	reqBody := map[string]interface{}{"foods": foods}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, w := newUpdateRequest(historyID, testUserID, string(body))

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "too many food items")
}

func TestUpdateHistoryRequest_Validate_ExactlyMaxFoods(t *testing.T) {
	// ちょうど50個（上限値）の食材はバリデーションを通過する
	foods := make([]UpdateFoodItem, 50)
	for i := 0; i < 50; i++ {
		foods[i] = UpdateFoodItem{
			Name:            fmt.Sprintf("食材%d", i+1),
			EstimatedAmount: "100g",
			Calories:        100.0,
			Protein:         10.0,
			Fat:             5.0,
			Carbohydrates:   20.0,
		}
	}
	req := UpdateHistoryRequest{Foods: foods}
	assert.NoError(t, req.Validate())
}

func TestHistoryHandler_HandleUpdate_FoodCountChangeSkipsRecalculation(t *testing.T) {
	historyID := uuid.New()
	createdAt := time.Now()
	mealDate := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	testUserID := "test-user-123"

	mockRecalculator := &MockNutritionRecalculator{}

	mockRepo := &MockAnalysisRepository{
		GetHistoryDetailFunc: func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
			// 旧データは1食材。リクエストは2食材なので食材数変更 → 再計算スキップ
			return whiteRiceDetail(historyID, createdAt, mealDate), nil
		},
		UpdateResultFunc: func(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
			return nil
		},
	}

	handler := NewHistoryHandler(mockRepo, mockRecalculator)

	// 2食材のリクエスト（旧データは1食材）
	body := `{"foods":[{"name":"玄米","estimated_amount":"150g","calories_kcal":228,"protein_g":4.2,"fat_g":1.8,"carbohydrates_g":47.8},{"name":"味噌汁","estimated_amount":"200ml","calories_kcal":30,"protein_g":2.0,"fat_g":0.5,"carbohydrates_g":4.0}]}`
	req, w := newUpdateRequest(historyID, testUserID, body)

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 少し待って、非同期再計算が呼ばれないことを確認
	time.Sleep(100 * time.Millisecond)
	assert.False(t, mockRecalculator.called.Load())
}
