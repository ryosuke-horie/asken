package handler

import (
	"context"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

var (
	_ FoodService                    = (*MockFoodService)(nil)
	_ repository.AnalysisRepository = (*MockAnalysisRepository)(nil)
)

// MockFoodService はテスト用のモックFoodService
type MockFoodService struct {
	AnalyzeFoodImageFunc func(ctx context.Context, imagePath string) (*service.AnalysisResult, error)
	AnalyzeFoodTextFunc  func(ctx context.Context, inputText string) (*service.AnalysisResult, error)
}

func (m *MockFoodService) AnalyzeFoodImage(ctx context.Context, imagePath string) (*service.AnalysisResult, error) {
	if m.AnalyzeFoodImageFunc != nil {
		return m.AnalyzeFoodImageFunc(ctx, imagePath)
	}
	return nil, nil
}

func (m *MockFoodService) AnalyzeFoodText(ctx context.Context, inputText string) (*service.AnalysisResult, error) {
	if m.AnalyzeFoodTextFunc != nil {
		return m.AnalyzeFoodTextFunc(ctx, inputText)
	}
	return nil, nil
}

// MockAnalysisRepository はテスト用のモックAnalysisRepository
type MockAnalysisRepository struct {
	CreateRequestFunc           func(ctx context.Context, imagePath string, mealType string, mealDate string, userID *string) (uuid.UUID, error)
	CreateRequestWithTextFunc   func(ctx context.Context, inputText string, mealType string, mealDate string, userID *string) (uuid.UUID, error)
	GetRequestFunc              func(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error)
	UpdateStatusFunc            func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error
	SaveResultFunc              func(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error
	GetResultFunc               func(ctx context.Context, userID string, requestID uuid.UUID) (*service.AnalysisResult, error)
	GetPendingRequestsFunc      func(ctx context.Context, limit int) ([]repository.AnalysisRequest, error)
	GetHistoryListFunc          func(ctx context.Context, userID string, page, limit int) ([]repository.HistoryItem, int, error)
	GetHistoryDetailFunc        func(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error)
	DeleteHistoryFunc           func(ctx context.Context, userID string, id uuid.UUID) error
	GetDailyMealsFunc           func(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error)
	CreateRequestFromMylistFunc func(ctx context.Context, inputText string, mealType string, mealDate string, userID *string, result *service.AnalysisResult) (uuid.UUID, error)
	CreateSkippedMealFunc       func(ctx context.Context, mealType string, mealDate string, userID *string) (uuid.UUID, error)
	UpdateResultFunc            func(ctx context.Context, userID string, historyID uuid.UUID, foods []gemini.NutritionInfo) error
}

func (m *MockAnalysisRepository) CreateRequest(ctx context.Context, imagePath string, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
	if m.CreateRequestFunc != nil {
		return m.CreateRequestFunc(ctx, imagePath, mealType, mealDate, userID)
	}
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) CreateRequestWithText(ctx context.Context, inputText string, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
	if m.CreateRequestWithTextFunc != nil {
		return m.CreateRequestWithTextFunc(ctx, inputText, mealType, mealDate, userID)
	}
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) GetRequest(ctx context.Context, userID string, id uuid.UUID) (*repository.AnalysisRequest, error) {
	if m.GetRequestFunc != nil {
		return m.GetRequestFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status, errorMessage)
	}
	return nil
}

func (m *MockAnalysisRepository) SaveResult(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error {
	if m.SaveResultFunc != nil {
		return m.SaveResultFunc(ctx, requestID, result)
	}
	return nil
}

func (m *MockAnalysisRepository) GetResult(ctx context.Context, userID string, requestID uuid.UUID) (*service.AnalysisResult, error) {
	if m.GetResultFunc != nil {
		return m.GetResultFunc(ctx, userID, requestID)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) GetPendingRequests(ctx context.Context, limit int) ([]repository.AnalysisRequest, error) {
	if m.GetPendingRequestsFunc != nil {
		return m.GetPendingRequestsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) GetHistoryList(ctx context.Context, userID string, page, limit int) ([]repository.HistoryItem, int, error) {
	if m.GetHistoryListFunc != nil {
		return m.GetHistoryListFunc(ctx, userID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockAnalysisRepository) GetHistoryDetail(ctx context.Context, userID string, id uuid.UUID) (*repository.HistoryDetail, error) {
	if m.GetHistoryDetailFunc != nil {
		return m.GetHistoryDetailFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) DeleteHistory(ctx context.Context, userID string, id uuid.UUID) error {
	if m.DeleteHistoryFunc != nil {
		return m.DeleteHistoryFunc(ctx, userID, id)
	}
	return nil
}

func (m *MockAnalysisRepository) GetDailyMeals(ctx context.Context, userID string, date string, tz string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
	if m.GetDailyMealsFunc != nil {
		return m.GetDailyMealsFunc(ctx, userID, date, tz)
	}
	return nil, repository.DailyTotal{}, nil
}

func (m *MockAnalysisRepository) CreateRequestFromMylist(ctx context.Context, inputText string, mealType string, mealDate string, userID *string, result *service.AnalysisResult) (uuid.UUID, error) {
	if m.CreateRequestFromMylistFunc != nil {
		return m.CreateRequestFromMylistFunc(ctx, inputText, mealType, mealDate, userID, result)
	}
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) CreateSkippedMeal(ctx context.Context, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
	if m.CreateSkippedMealFunc != nil {
		return m.CreateSkippedMealFunc(ctx, mealType, mealDate, userID)
	}
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) UpdateResult(ctx context.Context, userID string, historyID uuid.UUID, foods []gemini.NutritionInfo) error {
	if m.UpdateResultFunc != nil {
		return m.UpdateResultFunc(ctx, userID, historyID, foods)
	}
	return nil
}
