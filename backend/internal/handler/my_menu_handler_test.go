package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
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
