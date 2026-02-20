package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIngredientRepository はテスト用のモックIngredientRepository
type MockIngredientRepository struct {
	CreateFunc  func(ctx context.Context, userID string, input repository.CreateIngredientInput) (*repository.Ingredient, error)
	ListFunc    func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error)
	GetByIDFunc func(ctx context.Context, userID string, ingredientID string) (*repository.Ingredient, error)
	UpdateFunc  func(ctx context.Context, userID string, ingredientID string, input repository.UpdateIngredientInput) (*repository.Ingredient, error)
	DeleteFunc  func(ctx context.Context, userID string, ingredientID string) error
}

func (m *MockIngredientRepository) Create(ctx context.Context, userID string, input repository.CreateIngredientInput) (*repository.Ingredient, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, input)
	}
	return nil, nil
}

func (m *MockIngredientRepository) List(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID, category)
	}
	return nil, nil
}

func (m *MockIngredientRepository) GetByID(ctx context.Context, userID string, ingredientID string) (*repository.Ingredient, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, userID, ingredientID)
	}
	return nil, nil
}

func (m *MockIngredientRepository) Update(ctx context.Context, userID string, ingredientID string, input repository.UpdateIngredientInput) (*repository.Ingredient, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, userID, ingredientID, input)
	}
	return nil, nil
}

func (m *MockIngredientRepository) Delete(ctx context.Context, userID string, ingredientID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, userID, ingredientID)
	}
	return nil
}

// newIngredientRequest はIngredientHandlerへのリクエストを生成するヘルパー
func newIngredientRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, url, &buf)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user-id")
	return req.WithContext(ctx)
}

// sampleIngredient はテスト用のIngredientを生成するヘルパー
func sampleIngredient(id string) *repository.Ingredient {
	now := time.Now()
	return &repository.Ingredient{
		ID:        id,
		Name:      "鶏むね肉",
		Category:  "meat",
		Quantity:  500,
		Unit:      "g",
		Source:    "manual",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewIngredientHandler_NilPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewIngredientHandler(nil)
	})
}

func TestIngredientHandler_MethodNotAllowed(t *testing.T) {
	h := NewIngredientHandler(&MockIngredientRepository{})
	ingredientID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name   string
		method string
		url    string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{
			name:   "HandleList はPOST以外も不可",
			method: http.MethodPost,
			url:    "/api/ingredients",
			handle: h.HandleList,
		},
		{
			name:   "HandleCreate はGET不可",
			method: http.MethodGet,
			url:    "/api/ingredients",
			handle: h.HandleCreate,
		},
		{
			name:   "HandleUpdate はGET不可",
			method: http.MethodGet,
			url:    "/api/ingredients/" + ingredientID,
			handle: h.HandleUpdate,
		},
		{
			name:   "HandleDelete はGET不可",
			method: http.MethodGet,
			url:    "/api/ingredients/" + ingredientID,
			handle: h.HandleDelete,
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

func TestIngredientHandler_Unauthorized(t *testing.T) {
	h := NewIngredientHandler(&MockIngredientRepository{})
	ingredientID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name   string
		method string
		url    string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{"HandleList 未認証", http.MethodGet, "/api/ingredients", h.HandleList},
		{"HandleCreate 未認証", http.MethodPost, "/api/ingredients", h.HandleCreate},
		{"HandleUpdate 未認証", http.MethodPut, "/api/ingredients/" + ingredientID, h.HandleUpdate},
		{"HandleDelete 未認証", http.MethodDelete, "/api/ingredients/" + ingredientID, h.HandleDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			tt.handle(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestIngredientHandler_HandleList(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		mockList       func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error)
		expectedStatus int
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name: "食材一覧を正常取得できる",
			url:  "/api/ingredients",
			mockList: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
				return []repository.Ingredient{*sampleIngredient("id-1")}, nil
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp IngredientsListResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Ingredients, 1)
				assert.Equal(t, "鶏むね肉", resp.Ingredients[0].Name)
			},
		},
		{
			name: "categoryフィルタで絞り込める",
			url:  "/api/ingredients?category=meat",
			mockList: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
				assert.Equal(t, "meat", category)
				return []repository.Ingredient{*sampleIngredient("id-1")}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "不正なcategoryでBadRequest",
			url:  "/api/ingredients?category=invalid",
			mockList: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
				return nil, nil
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "空の食材一覧を返す",
			url:  "/api/ingredients",
			mockList: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
				return []repository.Ingredient{}, nil
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp IngredientsListResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Empty(t, resp.Ingredients)
			},
		},
		{
			name: "リポジトリエラーで500",
			url:  "/api/ingredients",
			mockList: func(ctx context.Context, userID string, category string) ([]repository.Ingredient, error) {
				return nil, fmt.Errorf("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockIngredientRepository{ListFunc: tt.mockList}
			h := NewIngredientHandler(mock)

			req := newIngredientRequest(t, http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			h.HandleList(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestIngredientHandler_HandleCreate(t *testing.T) {
	validBody := CreateIngredientRequest{
		Name:     "鶏むね肉",
		Category: "meat",
		Quantity: 500,
		Unit:     "g",
		Source:   "manual",
	}

	tests := []struct {
		name           string
		body           interface{}
		mockCreate     func(ctx context.Context, userID string, input repository.CreateIngredientInput) (*repository.Ingredient, error)
		expectedStatus int
	}{
		{
			name: "食材を正常作成できる",
			body: validBody,
			mockCreate: func(ctx context.Context, userID string, input repository.CreateIngredientInput) (*repository.Ingredient, error) {
				assert.Equal(t, "鶏むね肉", input.Name)
				assert.Equal(t, "meat", input.Category)
				assert.Equal(t, float64(500), input.Quantity)
				assert.Equal(t, "g", input.Unit)
				assert.Equal(t, "manual", input.Source)
				return sampleIngredient("new-id"), nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "nameが空でBadRequest",
			body:           CreateIngredientRequest{Category: "meat", Quantity: 100, Unit: "g", Source: "manual"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "nameが100文字超でBadRequest",
			body:           CreateIngredientRequest{Name: string(make([]byte, 101)), Category: "meat", Quantity: 100, Unit: "g", Source: "manual"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "categoryが不正でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "invalid", Quantity: 100, Unit: "g", Source: "manual"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "quantityが0でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 0, Unit: "g", Source: "manual"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "quantityが負でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "meat", Quantity: -1, Unit: "g", Source: "manual"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unitが不正でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 100, Unit: "invalid", Source: "manual"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "sourceが不正でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 100, Unit: "g", Source: "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "購入日・消費期限あり",
			body: CreateIngredientRequest{
				Name:         "鶏むね肉",
				Category:     "meat",
				Quantity:     500,
				Unit:         "g",
				Source:       "receipt",
				PurchaseDate: "2026-02-18",
				ExpiryDate:   "2026-02-22",
			},
			mockCreate: func(ctx context.Context, userID string, input repository.CreateIngredientInput) (*repository.Ingredient, error) {
				assert.NotNil(t, input.PurchaseDate)
				assert.NotNil(t, input.ExpiryDate)
				return sampleIngredient("new-id"), nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "不正な日付形式でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 100, Unit: "g", Source: "manual", PurchaseDate: "invalid-date"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "expiryDateが不正形式でBadRequest",
			body:           CreateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 100, Unit: "g", Source: "manual", ExpiryDate: "invalid-date"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "リポジトリエラーで500",
			body: validBody,
			mockCreate: func(ctx context.Context, userID string, input repository.CreateIngredientInput) (*repository.Ingredient, error) {
				return nil, fmt.Errorf("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockIngredientRepository{CreateFunc: tt.mockCreate}
			h := NewIngredientHandler(mock)

			req := newIngredientRequest(t, http.MethodPost, "/api/ingredients", tt.body)
			w := httptest.NewRecorder()
			h.HandleCreate(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestIngredientHandler_HandleUpdate(t *testing.T) {
	ingredientID := "550e8400-e29b-41d4-a716-446655440000"
	validBody := UpdateIngredientRequest{
		Name:     "鶏むね肉",
		Category: "meat",
		Quantity: 300,
		Unit:     "g",
	}

	tests := []struct {
		name           string
		url            string
		body           interface{}
		mockUpdate     func(ctx context.Context, userID string, ingredientID string, input repository.UpdateIngredientInput) (*repository.Ingredient, error)
		expectedStatus int
	}{
		{
			name: "食材を正常更新できる",
			url:  "/api/ingredients/" + ingredientID,
			body: validBody,
			mockUpdate: func(ctx context.Context, userID string, id string, input repository.UpdateIngredientInput) (*repository.Ingredient, error) {
				assert.Equal(t, ingredientID, id)
				assert.Equal(t, "鶏むね肉", input.Name)
				assert.Equal(t, float64(300), input.Quantity)
				return sampleIngredient(id), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "不正なIDでBadRequest",
			url:            "/api/ingredients/not-a-uuid",
			body:           validBody,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "存在しない食材でNotFound",
			url:  "/api/ingredients/" + ingredientID,
			body: validBody,
			mockUpdate: func(ctx context.Context, userID string, id string, input repository.UpdateIngredientInput) (*repository.Ingredient, error) {
				return nil, fmt.Errorf("食材が見つかりません: %w", repository.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "nameが空でBadRequest",
			url:            "/api/ingredients/" + ingredientID,
			body:           UpdateIngredientRequest{Category: "meat", Quantity: 100, Unit: "g"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "quantityが0でBadRequest",
			url:            "/api/ingredients/" + ingredientID,
			body:           UpdateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 0, Unit: "g"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "categoryが不正でBadRequest",
			url:            "/api/ingredients/" + ingredientID,
			body:           UpdateIngredientRequest{Name: "テスト", Category: "invalid", Quantity: 100, Unit: "g"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unitが不正でBadRequest",
			url:            "/api/ingredients/" + ingredientID,
			body:           UpdateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 100, Unit: "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "purchaseDateが不正形式でBadRequest",
			url:            "/api/ingredients/" + ingredientID,
			body:           UpdateIngredientRequest{Name: "テスト", Category: "meat", Quantity: 100, Unit: "g", PurchaseDate: "invalid-date"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "リポジトリエラーで500",
			url:  "/api/ingredients/" + ingredientID,
			body: validBody,
			mockUpdate: func(ctx context.Context, userID string, id string, input repository.UpdateIngredientInput) (*repository.Ingredient, error) {
				return nil, fmt.Errorf("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockIngredientRepository{UpdateFunc: tt.mockUpdate}
			h := NewIngredientHandler(mock)

			req := newIngredientRequest(t, http.MethodPut, tt.url, tt.body)
			w := httptest.NewRecorder()
			h.HandleUpdate(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestIngredientHandler_HandleDelete(t *testing.T) {
	ingredientID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name           string
		url            string
		mockDelete     func(ctx context.Context, userID string, ingredientID string) error
		expectedStatus int
	}{
		{
			name: "食材を正常削除できる",
			url:  "/api/ingredients/" + ingredientID,
			mockDelete: func(ctx context.Context, userID string, id string) error {
				assert.Equal(t, ingredientID, id)
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "不正なIDでBadRequest",
			url:            "/api/ingredients/not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "存在しない食材でNotFound",
			url:  "/api/ingredients/" + ingredientID,
			mockDelete: func(ctx context.Context, userID string, id string) error {
				return fmt.Errorf("食材が見つかりません: %w", repository.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "リポジトリエラーで500",
			url:  "/api/ingredients/" + ingredientID,
			mockDelete: func(ctx context.Context, userID string, id string) error {
				return fmt.Errorf("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockIngredientRepository{DeleteFunc: tt.mockDelete}
			h := NewIngredientHandler(mock)

			req := newIngredientRequest(t, http.MethodDelete, tt.url, nil)
			w := httptest.NewRecorder()
			h.HandleDelete(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestIngredientHandler_toIngredientResponse(t *testing.T) {
	purchaseDate := time.Date(2026, 2, 18, 0, 0, 0, 0, time.UTC)
	expiryDate := time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC)

	ingredient := repository.Ingredient{
		ID:           "test-id",
		Name:         "鶏むね肉",
		Category:     "meat",
		Quantity:     500,
		Unit:         "g",
		PurchaseDate: &purchaseDate,
		ExpiryDate:   &expiryDate,
		Source:       "receipt",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	resp := toIngredientResponse(ingredient)

	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "鶏むね肉", resp.Name)
	assert.Equal(t, "2026-02-18", resp.PurchaseDate)
	assert.Equal(t, "2026-02-22", resp.ExpiryDate)
}

func TestIngredientHandler_toIngredientResponse_NilDates(t *testing.T) {
	ingredient := repository.Ingredient{
		ID:        "test-id",
		Name:      "鶏むね肉",
		Category:  "meat",
		Quantity:  500,
		Unit:      "g",
		Source:    "manual",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	resp := toIngredientResponse(ingredient)

	assert.Empty(t, resp.PurchaseDate)
	assert.Empty(t, resp.ExpiryDate)
}

func TestValidateIngredientName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"正常", "鶏むね肉", false},
		{"空文字", "", true},
		{"100文字", string(make([]rune, 100)), false},
		{"101文字", string(make([]rune, 101)), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIngredientName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIngredientQuantity(t *testing.T) {
	tests := []struct {
		name     string
		quantity float64
		wantErr  bool
	}{
		{"正の値", 100, false},
		{"小数", 0.1, false},
		{"0", 0, true},
		{"負の値", -1, true},
		{"NaN", math.NaN(), true},
		{"正の無限大", math.Inf(1), true},
		{"上限超過", 1000000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIngredientQuantity(tt.quantity)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIngredientUnit(t *testing.T) {
	tests := []struct {
		name    string
		unit    string
		wantErr bool
	}{
		{"g", "g", false},
		{"ml", "ml", false},
		{"個", "個", false},
		{"不正な単位", "oz", true},
		{"空文字", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIngredientUnit(tt.unit)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseOptionalDate(t *testing.T) {
	tests := []struct {
		name      string
		dateStr   string
		fieldName string
		wantNil   bool
		wantErr   bool
	}{
		{"空文字はnil", "", "purchaseDate", true, false},
		{"正常な日付", "2026-02-18", "purchaseDate", false, false},
		{"不正な日付", "invalid", "purchaseDate", false, true},
		{"形式不正", "2026/02/18", "purchaseDate", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseOptionalDate(tt.dateStr, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
			}
		})
	}
}

func TestIngredientHandler_HandleCreate_InvalidJSON(t *testing.T) {
	mock := &MockIngredientRepository{}
	h := NewIngredientHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/api/ingredients", bytes.NewBufferString("{invalid json}"))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user-id")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIngredientHandler_HandleUpdate_InvalidJSON(t *testing.T) {
	ingredientID := "550e8400-e29b-41d4-a716-446655440000"
	mock := &MockIngredientRepository{}
	h := NewIngredientHandler(mock)

	req := httptest.NewRequest(http.MethodPut, "/api/ingredients/"+ingredientID, bytes.NewBufferString("{invalid json}"))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user-id")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.HandleUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExtractIngredientID(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectedID  string
		expectError bool
	}{
		{
			name:       "正常にIDを抽出できる",
			path:       "/api/ingredients/550e8400-e29b-41d4-a716-446655440000",
			expectedID: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:        "UUIDでないIDでエラー",
			path:        "/api/ingredients/not-a-uuid",
			expectError: true,
		},
		{
			name:        "空IDでエラー（二重スラッシュ）",
			path:        "/api/ingredients//",
			expectError: true,
		},
		{
			name:        "パスが短すぎてエラー",
			path:        "/api",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := extractIngredientID(tt.path)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}
