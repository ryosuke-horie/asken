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

func TestHandleSuggest_CountNormalization(t *testing.T) {
	// count=0以下→3、count=6以上→5 に正規化されることを確認
	// Geminiプロンプトに "%d件提案" として正規化後の件数が含まれる
	tests := []struct {
		name          string
		count         int
		expectedCount string
	}{
		{"count=0はデフォルト3に正規化", 0, "3件"},
		{"count負値はデフォルト3に正規化", -1, "3件"},
		{"count=6は上限5に正規化", 6, "5件"},
		{"count=5は正常値のまま", 5, "5件"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPrompt := ""
			mockHTTPClient := &gemini.MockGeminiHTTPClient{
				ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
					capturedPrompt = prompt
					return &gemini.Response{Response: `{"suggestions": []}`}, nil
				},
			}
			handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

			body := map[string]interface{}{"mealType": "lunch", "count": tt.count}
			req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", body)
			w := httptest.NewRecorder()
			handler.HandleSuggest(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Contains(t, capturedPrompt, tt.expectedCount)
		})
	}
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

// --- NewMenuSuggestionHandler パニックテスト ---

func TestNewMenuSuggestionHandler_NilDependency_Panics(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	suggester := gemini.NewMenuSuggesterWithHTTPClient(mockHTTPClient)
	menuRepo := &MockMenuSuggestionRepository{}
	ingRepo := &MockIngredientRepository{}
	nutRepo := &MockNutritionGoalRepository{}
	weightRepo := &MockWeightRecordRepository{}
	anaRepo := &MockAnalysisRepository{}

	tests := []struct {
		name  string
		build func()
	}{
		{"nilMenuRepo", func() { NewMenuSuggestionHandler(nil, ingRepo, nutRepo, weightRepo, anaRepo, suggester) }},
		{"nilIngredientRepo", func() { NewMenuSuggestionHandler(menuRepo, nil, nutRepo, weightRepo, anaRepo, suggester) }},
		{"nilNutritionGoalRepo", func() { NewMenuSuggestionHandler(menuRepo, ingRepo, nil, weightRepo, anaRepo, suggester) }},
		{"nilWeightRecordRepo", func() { NewMenuSuggestionHandler(menuRepo, ingRepo, nutRepo, nil, anaRepo, suggester) }},
		{"nilAnalysisRepo", func() { NewMenuSuggestionHandler(menuRepo, ingRepo, nutRepo, weightRepo, nil, suggester) }},
		{"nilSuggester", func() { NewMenuSuggestionHandler(menuRepo, ingRepo, nutRepo, weightRepo, anaRepo, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, tt.build)
		})
	}
}

// --- HandleSuggest 追加テスト ---

func TestHandleSuggest_IngredientRepoError(t *testing.T) {
	ingredientRepo := &MockIngredientRepository{
		ListFunc: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
			return nil, fmt.Errorf("DB接続エラー")
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, ingredientRepo, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "lunch", "count": 2})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSuggest_GeminiError(t *testing.T) {
	ingredientRepo := &MockIngredientRepository{
		ListFunc: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
			return []repository.Ingredient{}, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return nil, fmt.Errorf("Gemini API タイムアウト")
		},
	}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, ingredientRepo, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "dinner", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSuggest_CreateError(t *testing.T) {
	geminiRespJSON := `{
		"suggestions": [
			{
				"title": "テスト料理",
				"description": "説明",
				"reason": "理由",
				"ingredients": [],
				"estimatedNutrition": {"calories": 300.0, "protein": 20.0, "fat": 5.0, "carbohydrates": 40.0}
			}
		]
	}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			return nil, fmt.Errorf("Firestore書き込みエラー")
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "breakfast", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSuggest_NutritionGoalError_StillSucceeds(t *testing.T) {
	// 栄養目標取得に失敗してもメニューサジェストは成功する（非致命的エラー）
	geminiRespJSON := `{"suggestions": [{"title": "テスト料理", "description": "説明", "reason": "理由", "ingredients": [], "estimatedNutrition": {"calories": 300.0, "protein": 20.0, "fat": 5.0, "carbohydrates": 40.0}}]}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}
	saved := sampleMenuSuggestion("new-id-1")
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			return saved, nil
		},
	}
	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			return nil, fmt.Errorf("DB接続エラー（ErrNotFound以外）")
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, nutritionRepo, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "lunch", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandleSuggest_AnalysisHistoryError_StillSucceeds(t *testing.T) {
	// 食事履歴取得に失敗してもメニューサジェストは成功する（非致命的エラー）
	geminiRespJSON := `{"suggestions": [{"title": "テスト料理2", "description": "説明", "reason": "理由", "ingredients": [], "estimatedNutrition": {"calories": 300.0, "protein": 20.0, "fat": 5.0, "carbohydrates": 40.0}}]}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}
	saved := sampleMenuSuggestion("new-id-2")
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			return saved, nil
		},
	}
	analysisRepo := &MockAnalysisRepository{
		GetHistoryListFunc: func(ctx context.Context, userID string, page, limit int) ([]repository.HistoryItem, int, error) {
			return nil, 0, fmt.Errorf("食事履歴取得エラー")
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, analysisRepo, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "dinner", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandleSuggest_NutritionGoalErrNotFound_StillSucceeds(t *testing.T) {
	// 栄養目標がErrNotFoundの場合も成功する（未設定状態）
	geminiRespJSON := `{"suggestions": [{"title": "テスト料理_notfound", "description": "説明", "reason": "理由", "ingredients": [], "estimatedNutrition": {"calories": 300.0, "protein": 20.0, "fat": 5.0, "carbohydrates": 40.0}}]}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}
	saved := sampleMenuSuggestion("new-id-notfound")
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			return saved, nil
		},
	}
	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			return nil, repository.ErrNotFound
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, nutritionRepo, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "lunch", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandleSuggest_WithNutritionGoal_IncludesGoalInContext(t *testing.T) {
	// 栄養目標が設定されている場合、コンテキストに含まれることを確認
	capturedPrompt := ""
	geminiRespJSON := `{"suggestions": [{"title": "栄養目標考慮メニュー", "description": "説明", "reason": "理由", "ingredients": [], "estimatedNutrition": {"calories": 400.0, "protein": 30.0, "fat": 10.0, "carbohydrates": 50.0}}]}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			capturedPrompt = prompt
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}
	saved := sampleMenuSuggestion("new-id-goal")
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			return saved, nil
		},
	}
	nutritionRepo := &MockNutritionGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*repository.NutritionGoal, error) {
			return &repository.NutritionGoal{
				TargetCalories:      2500,
				TargetProtein:       180,
				TargetFat:           70,
				TargetCarbohydrates: 300,
				Phase:               "bulk",
			}, nil
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, nutritionRepo, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "lunch", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, capturedPrompt, "2500") // 栄養目標がプロンプトに含まれる
}

func TestHandleSuggest_WeightRecordError_StillSucceeds(t *testing.T) {
	// 体重記録取得に失敗してもメニューサジェストは成功する（非致命的エラー）
	geminiRespJSON := `{"suggestions": [{"title": "テスト料理3", "description": "説明", "reason": "理由", "ingredients": [], "estimatedNutrition": {"calories": 300.0, "protein": 20.0, "fat": 5.0, "carbohydrates": 40.0}}]}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: geminiRespJSON}, nil
		},
	}
	saved := sampleMenuSuggestion("new-id-3")
	menuRepo := &MockMenuSuggestionRepository{
		CreateFunc: func(ctx context.Context, userID string, input repository.CreateMenuSuggestionInput) (*repository.MenuSuggestion, error) {
			return saved, nil
		},
	}
	weightRepo := &MockWeightRecordRepository{
		ListRecordsByDateRangeFunc: func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
			return nil, fmt.Errorf("体重記録取得エラー")
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, weightRepo, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggest", map[string]interface{}{"mealType": "breakfast", "count": 1})
	w := httptest.NewRecorder()
	handler.HandleSuggest(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// --- HandleGet 追加テスト ---

func TestHandleGet_Unauthorized(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := httptest.NewRequest(http.MethodGet, "/api/menu/suggestions/suggestion-1", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGet_MethodNotAllowed(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleGet_InvalidPath(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGet_InternalError(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		GetByIDFunc: func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
			return nil, fmt.Errorf("Firestore接続エラー")
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/suggestion-1", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGet_RecipeGenerationFails_ReturnsOK(t *testing.T) {
	// レシピ生成失敗はノンクリティカル: レシピなしで200 OKを返す
	suggestion := sampleMenuSuggestion("suggestion-3")
	suggestion.Recipe = ""

	menuRepo := &MockMenuSuggestionRepository{
		GetByIDFunc: func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
			return suggestion, nil
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return nil, fmt.Errorf("Gemini APIエラー")
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/suggestion-3", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp menuSuggestionResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "", resp.Recipe) // レシピなしで返る
}

func TestHandleGet_UpdateRecipeFails_ReturnsOK(t *testing.T) {
	// UpdateRecipe失敗もノンクリティカル: レシピなしで200 OKを返す
	suggestion := sampleMenuSuggestion("suggestion-4")
	suggestion.Recipe = ""

	menuRepo := &MockMenuSuggestionRepository{
		GetByIDFunc: func(ctx context.Context, userID string, id string) (*repository.MenuSuggestion, error) {
			return suggestion, nil
		},
		UpdateRecipeFunc: func(ctx context.Context, userID string, id string, recipe string) error {
			return fmt.Errorf("Firestore更新エラー")
		},
	}
	recipeJSON := `{"recipe": "生成されたレシピ"}`
	mockHTTPClient := &gemini.MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *gemini.Schema) (*gemini.Response, error) {
			return &gemini.Response{Response: recipeJSON}, nil
		},
	}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions/suggestion-4", nil)
	w := httptest.NewRecorder()
	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp menuSuggestionResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "", resp.Recipe) // 保存失敗のためレシピなしで返る
}

// --- HandleAccept 追加テスト ---

func TestHandleAccept_Unauthorized(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := httptest.NewRequest(http.MethodPost, "/api/menu/suggestions/suggestion-1/accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleAccept_InvalidPath(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions//accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleAccept_InternalError(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		AcceptFunc: func(ctx context.Context, userID string, id string) (*repository.AcceptMenuSuggestionResult, error) {
			return nil, fmt.Errorf("Firestoreトランザクションエラー")
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1/accept", nil)
	w := httptest.NewRecorder()
	handler.HandleAccept(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- HandleDismiss 追加テスト ---

func TestHandleDismiss_InvalidPath(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions//dismiss", nil)
	w := httptest.NewRecorder()
	handler.HandleDismiss(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDismiss_InternalError(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		DismissFunc: func(ctx context.Context, userID string, id string) error {
			return fmt.Errorf("Firestoreエラー")
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodPost, "/api/menu/suggestions/suggestion-1/dismiss", nil)
	w := httptest.NewRecorder()
	handler.HandleDismiss(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- HandleList 追加テスト ---

func TestHandleList_InternalError(t *testing.T) {
	menuRepo := &MockMenuSuggestionRepository{
		ListFunc: func(ctx context.Context, userID string, status string, limit int) ([]repository.MenuSuggestion, error) {
			return nil, fmt.Errorf("Firestore接続エラー")
		},
	}
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(menuRepo, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions", nil)
	w := httptest.NewRecorder()
	handler.HandleList(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleList_InvalidLimit(t *testing.T) {
	mockHTTPClient := &gemini.MockGeminiHTTPClient{}
	handler := newTestMenuSuggestionHandler(&MockMenuSuggestionRepository{}, &MockIngredientRepository{}, &MockNutritionGoalRepository{}, &MockWeightRecordRepository{}, &MockAnalysisRepository{}, mockHTTPClient)

	tests := []struct {
		name  string
		limit string
	}{
		{"文字列", "abc"},
		{"ゼロ", "0"},
		{"負の数", "-1"},
		{"上限超え", "51"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newMenuRequest(t, http.MethodGet, "/api/menu/suggestions?limit="+tt.limit, nil)
			w := httptest.NewRecorder()
			handler.HandleList(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
