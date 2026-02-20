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
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMenuSuggestionRepository はMenuSuggestionRepository用テストモック
type MockMenuSuggestionRepository struct {
	CreateFunc       func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error)
	ListFunc         func(ctx context.Context, userID string, status string, limit int) ([]repository.MenuSuggestion, error)
	GetByIDFunc      func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error)
	UpdateRecipeFunc func(ctx context.Context, userID string, id string, recipe string) error
	AcceptFunc       func(ctx context.Context, userID string, id string) (*repository.AcceptMenuSuggestionResult, error)
	DismissFunc      func(ctx context.Context, userID string, id string) error
}

func (m *MockMenuSuggestionRepository) Create(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, input)
	}
	return nil, nil
}

func (m *MockMenuSuggestionRepository) List(ctx context.Context, userID string, status string, limit int) ([]repository.MenuSuggestion, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID, status, limit)
	}
	return []repository.MenuSuggestion{}, nil
}

func (m *MockMenuSuggestionRepository) GetByID(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockMenuSuggestionRepository) UpdateRecipe(ctx context.Context, userID string, id string, recipe string) error {
	if m.UpdateRecipeFunc != nil {
		return m.UpdateRecipeFunc(ctx, userID, id, recipe)
	}
	return nil
}

func (m *MockMenuSuggestionRepository) Accept(ctx context.Context, userID string, id string) (*repository.AcceptMenuSuggestionResult, error) {
	if m.AcceptFunc != nil {
		return m.AcceptFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockMenuSuggestionRepository) Dismiss(ctx context.Context, userID string, id string) error {
	if m.DismissFunc != nil {
		return m.DismissFunc(ctx, userID, id)
	}
	return nil
}

// MockMenuSuggester はMenuSuggester用テストモック（内部処理に依存せずHandlerをテスト可能にする）
// HandleSuggestではgemini.MenuSuggesterを直接使用するため、インターフェース化できない。
// テスト用にNewMenuSuggesterWithHTTPClientを使ってモックHTTPClientで注入する。
func newTestMenuSuggestionHandler(
	menuRepo *MockMenuSuggestionRepository,
	ingredientRepo *MockIngredientRepository,
	nutritionRepo *MockNutritionGoalRepository,
	weightRepo *MockWeightRecordRepository,
	analysisRepo *MockAnalysisRepository,
	mockHTTPClient gemini.GeminiHTTPClient,
) *MenuSuggestionHandler {
	suggester := gemini.NewMenuSuggesterWithHTTPClient(mockHTTPClient)
	return NewMenuSuggestionHandler(menuRepo, ingredientRepo, nutritionRepo, weightRepo, analysisRepo, suggester)
}

// sampleMenuSuggestion はテスト用のMenuSuggestionを生成するヘルパー
func sampleMenuSuggestion(id string) *repository.MenuSuggestion {
	now := time.Now()
	return &repository.MenuSuggestion{
		ID:          id,
		Title:       "テストメニュー",
		Description: "テストの説明",
		Reason:      "テストの理由",
		IngredientsUsed: []repository.MenuSuggestionIngredient{
			{IngredientID: "ing-1", Name: "鶏むね肉", Quantity: 200, Unit: "g"},
		},
		EstimatedNutrition: repository.EstimatedNutrition{
			Calories: 350, Protein: 40, Fat: 8, Carbohydrates: 15,
		},
		MealType:  "lunch",
		Status:    "suggested",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newMenuRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, url, &buf)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user-id")
	return req.WithContext(ctx)
}

// --- HandleList ---

func TestHandleList_Success(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		ListFunc: func(ctx context.Context, userID string, status string, limit int) ([]repository.MenuSuggestion, error) {
			assert.Equal(t, "test-user-id", userID)
			assert.Equal(t, "suggested", status)
			return []repository.MenuSuggestion{*sampleMenuSuggestion("id-1")}, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp suggestionsListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Suggestions, 1)
	assert.Equal(t, "id-1", resp.Suggestions[0].ID)
	assert.Equal(t, "テストメニュー", resp.Suggestions[0].Title)
}

func TestHandleList_WithLimitQuery(t *testing.T) {
	capturedLimit := 0
	menuRepo := &MockMenuSuggestionRepository{
		ListFunc: func(ctx context.Context, userID string, status string, limit int) ([]repository.MenuSuggestion, error) {
			capturedLimit = limit
			return []repository.MenuSuggestion{}, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions?limit=20", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 20, capturedLimit)
}

func TestHandleList_InvalidStatus(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions?status=invalid", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleList_AllStatus(t *testing.T) {
	capturedStatus := "sentinel"
	menuRepo := &MockMenuSuggestionRepository{
		ListFunc: func(ctx context.Context, userID string, status string, limit int) ([]repository.MenuSuggestion, error) {
			capturedStatus = status
			return []repository.MenuSuggestion{}, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions?status=all", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", capturedStatus) // "all" は "" に変換される
}

func TestHandleList_Unauthorized(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := httptest.NewRequest(http.MethodGet, "/api/menu/suggestions", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleList_MethodNotAllowed(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// --- HandleGet ---

func TestHandleGet_Success_WithExistingRecipe(t *testing.T) {
	suggestion := sampleMenuSuggestion("suggestion-1")
	suggestion.Recipe = "既存のレシピ内容"
	menuRepo := &MockMenuSuggestionRepository{
		GetByIDFunc: func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
			assert.Equal(t, "suggestion-1", id)
			return suggestion, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/suggestion-1", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp menuSuggestionResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "suggestion-1", resp.ID)
	assert.Equal(t, "既存のレシピ内容", resp.Recipe)
}

func TestHandleGet_Success_RecipeLazyGeneration(t *testing.T) {
	suggestion := sampleMenuSuggestion("suggestion-2")
	// Recipe は空なので遅延生成される
	suggestion.Recipe = ""

	menuRepo := &MockMenuSuggestionRepository{
		GetByIDFunc: func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
			return suggestion, nil
		},
		UpdateRecipeFunc: func(ctx context.Context, userID string, id string, recipe string) error {
			assert.Equal(t, "suggestion-2", id)
			return nil
		},
	}
	recipeJSON := `{"recipe": "生成されたレシピ内容"}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: recipeJSON}, nil
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/suggestion-2", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp menuSuggestionResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "生成されたレシピ内容", resp.Recipe)
}

func TestHandleGet_NotFound(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		GetByIDFunc: func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
			return nil, fmt.Errorf("not found: %w", repository.ErrNotFound)
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- HandleAccept ---

func TestHandleAccept_Success(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		AcceptFunc: func(ctx context.Context, userID string, id string) (*repository.AcceptMenuSuggestionResult, error) {
			assert.Equal(t, "suggestion-1", id)
			return &repository.AcceptMenuSuggestionResult{
				AnalysisRequestID: "analysis-123",
				DeductedIngredients: []repository.DeductedIngredient{
					{IngredientID: "ing-1", Name: "鶏むね肉", Deducted: 200, Remaining: 100},
				},
			}, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1/accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp acceptResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "analysis-123", resp.AnalysisRequestID)
	require.Len(t, resp.DeductedIngredients, 1)
	assert.Equal(t, "鶏むね肉", resp.DeductedIngredients[0].Name)
	assert.Equal(t, float64(200), resp.DeductedIngredients[0].Deducted)
	assert.Equal(t, float64(100), resp.DeductedIngredients[0].Remaining)
}

func TestHandleAccept_NotFound(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		AcceptFunc: func(ctx context.Context, userID string, id string) (*repository.AcceptMenuSuggestionResult, error) {
			return nil, fmt.Errorf("not found: %w", repository.ErrNotFound)
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/nonexistent/accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleAccept_AlreadyProcessed(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		AcceptFunc: func(ctx context.Context, userID string, id string) (*repository.AcceptMenuSuggestionResult, error) {
			return nil, fmt.Errorf("採用トランザクションに失敗: %w", repository.ErrAlreadyProcessed)
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1/accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandleAccept_MethodNotAllowed(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/suggestion-1/accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// --- HandleDismiss ---

func TestHandleDismiss_Success(t *testing.T) {
	dismissed := false
	menuRepo := &MockMenuSuggestionRepository{
		DismissFunc: func(ctx context.Context, userID string, id string) error {
			assert.Equal(t, "suggestion-1", id)
			dismissed = true
			return nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1/dismiss", nil)
	w := httptest.NewRecorder()
	handler.HandleDismiss(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, dismissed)
}

func TestHandleDismiss_NotFound(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		DismissFunc: func(ctx context.Context, userID string, id string) error {
			return fmt.Errorf("not found: %w", repository.ErrNotFound)
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/nonexistent/dismiss", nil)
	w := httptest.NewRecorder()
	handler.HandleDismiss(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDismiss_AlreadyProcessed(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		DismissFunc: func(ctx context.Context, userID string, id string) error {
			return fmt.Errorf("却下トランザクションに失敗: %w", repository.ErrAlreadyProcessed)
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1/dismiss", nil)
	w := httptest.NewRecorder()
	handler.HandleDismiss(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandleDismiss_Unauthorized(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := httptest.NewRequest(http.MethodPost, "/api/menu/suggestions/suggestion-1/dismiss", nil)
	w := httptest.NewRecorder()
	handler.HandleDismiss(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- HandleSuggest ---

func TestHandleSuggest_InvalidMealType(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "invalid", "count": 3})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSuggest_Unauthorized(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := httptest.NewRequest(http.MethodPost, "/api/menu/suggest", bytes.NewBufferString(`{"mealType":"lunch","count":2}`))
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleSuggest_MethodNotAllowed(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggest", nil)
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleSuggest_Success(t *testing.T) {
	geminiRespJSON := `{
		"suggestions": [
			{
				"title": "チキン炒め",
				"description": "簡単チキン炒め",
				"reason": "タンパク質補給",
				"ingredients": [{"name": "鶏むね肉", "quantity": 200.0, "unit": "g"}],
				"estimatedNutrition": {"calories": 300.0, "protein": 35.0, "fat": 7.0, "carbohydrates": 10.0}
			}
		]
	}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}

	saved := sampleMenuSuggestion("new-id")
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			assert.Equal(t, "lunch", input.MealType)
			assert.Equal(t, "チキン炒め", input.Title)
			saved.Title = input.Title
			return saved, nil
		},
	}
	ingredientRepo := &MockIngredientRepository{
		ListFunc: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
			return []repository.Ingredient{
				{ID: "ing-1", Name: "鶏むね肉", Quantity: 300, Unit: "g"},
			}, nil
		},
	}

	handler := newTestMenuSuggestionHandler(menuRepo, ingredientRepo, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	body := map[string]interface{}{"mealType": "lunch", "count": 1}
	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", body)
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp suggestionsListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Suggestions, 1)
	assert.Equal(t, "チキン炒め", resp.Suggestions[0].Title)
}

// --- ユーティリティ関数 ---

func TestExtractMenuSuggestionID(t *testing.T) {
	tests := []struct {
		path    string
		want    string
		wantErr bool
	}{
		{"/api/menu/suggestions/abc-123", "abc-123", false},
		{"/api/menu/suggestions/", "", true},
		{"/api/menu/suggestions", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := extractMenuSuggestionID(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExtractMenuSuggestionIDFromSubPath(t *testing.T) {
	tests := []struct {
		path    string
		sub     string
		want    string
		wantErr bool
	}{
		{"/api/menu/suggestions/abc-123/accept", "accept", "abc-123", false},
		{"/api/menu/suggestions/abc-123/dismiss", "dismiss", "abc-123", false},
		{"/api/menu/suggestions/abc-123/other", "accept", "", true},
		{"/api/menu/suggestions//accept", "accept", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.sub, func(t *testing.T) {
			got, err := extractMenuSuggestionIDFromSubPath(tt.path, tt.sub)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestValidateMenuSuggestionMealType(t *testing.T) {
	valid := []string{"breakfast", "lunch", "dinner", "snack"}
	for _, v := range valid {
		assert.NoError(t, validateMenuSuggestionMealType(v))
	}
	assert.Error(t, validateMenuSuggestionMealType("invalid"))
	assert.Error(t, validateMenuSuggestionMealType(""))
}

func TestValidateSuggestionStatus(t *testing.T) {
	valid := []string{"suggested", "accepted", "dismissed"}
	for _, v := range valid {
		assert.NoError(t, validateSuggestionStatus(v))
	}
	assert.Error(t, validateSuggestionStatus("invalid"))
	assert.Error(t, validateSuggestionStatus("all"))
}
