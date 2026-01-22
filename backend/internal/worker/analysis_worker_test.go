package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/internal/service"
	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
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
	GetPendingRequestsFunc func(ctx context.Context, limit int) ([]repository.AnalysisRequest, error)
	UpdateStatusFunc       func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error
	SaveResultFunc         func(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error
}

func (m *MockAnalysisRepository) CreateRequest(ctx context.Context, imagePath string, mealType string, mealDate string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) CreateRequestWithText(ctx context.Context, inputText string, mealType string, mealDate string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) GetRequest(ctx context.Context, id uuid.UUID) (*repository.AnalysisRequest, error) {
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

func (m *MockAnalysisRepository) GetResult(ctx context.Context, requestID uuid.UUID) (*service.AnalysisResult, error) {
	return nil, nil
}

func (m *MockAnalysisRepository) GetPendingRequests(ctx context.Context, limit int) ([]repository.AnalysisRequest, error) {
	if m.GetPendingRequestsFunc != nil {
		return m.GetPendingRequestsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) GetHistoryList(ctx context.Context, page, limit int) ([]repository.HistoryItem, int, error) {
	return nil, 0, nil
}

func (m *MockAnalysisRepository) GetHistoryDetail(ctx context.Context, id uuid.UUID) (*repository.HistoryDetail, error) {
	return nil, nil
}

func (m *MockAnalysisRepository) DeleteHistory(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockAnalysisRepository) GetDailyMeals(ctx context.Context, date string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
	return nil, repository.DailyTotal{}, nil
}

func TestProcessRequest_Success(t *testing.T) {
	requestID := uuid.New()
	imagePath := "/uploads/test.jpg"

	// AnalyzeFoodImageが成功するケース
	mockService := &MockFoodService{
		AnalyzeFoodImageFunc: func(ctx context.Context, path string) (*service.AnalysisResult, error) {
			assert.Equal(t, imagePath, path)
			return &service.AnalysisResult{
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
				TotalCalories:      252,
				TotalProtein:       3.8,
				TotalFat:           0.5,
				TotalCarbohydrates: 55.7,
			}, nil
		},
	}

	updateStatusCalled := 0
	saveResultCalled := false

	mockRepo := &MockAnalysisRepository{
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error {
			updateStatusCalled++
			if updateStatusCalled == 1 {
				// 最初の呼び出しはprocessingへの更新
				assert.Equal(t, requestID, id)
				assert.Equal(t, repository.StatusProcessing, status)
				assert.Empty(t, errorMessage)
			}
			return nil
		},
		SaveResultFunc: func(ctx context.Context, id uuid.UUID, result *service.AnalysisResult) error {
			saveResultCalled = true
			assert.Equal(t, requestID, id)
			assert.NotNil(t, result)
			assert.Equal(t, 252.0, result.TotalCalories)
			return nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 5*time.Second)

	request := repository.AnalysisRequest{
		ID:        requestID,
		Status:    repository.StatusPending,
		InputType: repository.InputTypeImage,
		ImagePath: imagePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := worker.processRequest(context.Background(), request)

	assert.NoError(t, err)
	assert.Equal(t, 1, updateStatusCalled)
	assert.True(t, saveResultCalled)
}

func TestProcessRequest_AnalysisError(t *testing.T) {
	requestID := uuid.New()
	imagePath := "/uploads/test.jpg"

	// AnalyzeFoodImageが失敗するケース
	analysisError := errors.New("Gemini API タイムアウト")
	mockService := &MockFoodService{
		AnalyzeFoodImageFunc: func(ctx context.Context, path string) (*service.AnalysisResult, error) {
			return nil, analysisError
		},
	}

	updateStatusCalled := 0
	saveResultCalled := false

	mockRepo := &MockAnalysisRepository{
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error {
			updateStatusCalled++
			if updateStatusCalled == 1 {
				// processingへの更新
				assert.Equal(t, repository.StatusProcessing, status)
			} else if updateStatusCalled == 2 {
				// failedへの更新
				assert.Equal(t, requestID, id)
				assert.Equal(t, repository.StatusFailed, status)
				assert.Contains(t, errorMessage, "Gemini API タイムアウト")
			}
			return nil
		},
		SaveResultFunc: func(ctx context.Context, id uuid.UUID, result *service.AnalysisResult) error {
			saveResultCalled = true
			return nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 5*time.Second)

	request := repository.AnalysisRequest{
		ID:        requestID,
		Status:    repository.StatusPending,
		InputType: repository.InputTypeImage,
		ImagePath: imagePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := worker.processRequest(context.Background(), request)

	assert.NoError(t, err) // processRequestは内部でエラーをハンドリングするため、nilを返す
	assert.Equal(t, 2, updateStatusCalled)
	assert.False(t, saveResultCalled)
}

func TestProcessPendingRequests_NoPendingRequests(t *testing.T) {
	mockService := &MockFoodService{}

	getPendingCalled := false
	mockRepo := &MockAnalysisRepository{
		GetPendingRequestsFunc: func(ctx context.Context, limit int) ([]repository.AnalysisRequest, error) {
			getPendingCalled = true
			return []repository.AnalysisRequest{}, nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 5*time.Second)

	err := worker.processPendingRequests(context.Background())

	assert.NoError(t, err)
	assert.True(t, getPendingCalled)
}

func TestProcessPendingRequests_WithPendingRequests(t *testing.T) {
	requestID := uuid.New()

	mockService := &MockFoodService{
		AnalyzeFoodImageFunc: func(ctx context.Context, path string) (*service.AnalysisResult, error) {
			return &service.AnalysisResult{
				Foods:              []gemini.NutritionInfo{},
				TotalCalories:      100,
				TotalProtein:       10,
				TotalFat:           5,
				TotalCarbohydrates: 20,
			}, nil
		},
	}

	mockRepo := &MockAnalysisRepository{
		GetPendingRequestsFunc: func(ctx context.Context, limit int) ([]repository.AnalysisRequest, error) {
			return []repository.AnalysisRequest{
				{
					ID:        requestID,
					Status:    repository.StatusPending,
					InputType: repository.InputTypeImage,
					ImagePath: "/uploads/test.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error {
			return nil
		},
		SaveResultFunc: func(ctx context.Context, id uuid.UUID, result *service.AnalysisResult) error {
			return nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 5*time.Second)

	err := worker.processPendingRequests(context.Background())

	assert.NoError(t, err)
}

func TestWorkerStart_StopsOnContextCancel(t *testing.T) {
	mockService := &MockFoodService{}

	callCount := 0
	mockRepo := &MockAnalysisRepository{
		GetPendingRequestsFunc: func(ctx context.Context, limit int) ([]repository.AnalysisRequest, error) {
			callCount++
			return []repository.AnalysisRequest{}, nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	// ワーカーを別ゴルーチンで起動
	go worker.Start(ctx)

	// 少し待機してから停止
	time.Sleep(250 * time.Millisecond)
	cancel()

	// 停止後、少し待機
	time.Sleep(100 * time.Millisecond)

	// 少なくとも1回は呼ばれているはず（100msごとにポーリング）
	assert.GreaterOrEqual(t, callCount, 1)
}

func TestProcessRequest_TextInput_Success(t *testing.T) {
	requestID := uuid.New()
	inputText := "ご飯二杯, 焼肉"

	// AnalyzeFoodTextが成功するケース
	mockService := &MockFoodService{
		AnalyzeFoodTextFunc: func(ctx context.Context, text string) (*service.AnalysisResult, error) {
			assert.Equal(t, inputText, text)
			return &service.AnalysisResult{
				Foods: []gemini.NutritionInfo{
					{
						Name:            "白米",
						EstimatedAmount: "300g",
						Calories:        504,
						Protein:         7.6,
						Fat:             1.0,
						Carbohydrates:   111.4,
					},
					{
						Name:            "焼肉",
						EstimatedAmount: "100g",
						Calories:        371,
						Protein:         17.1,
						Fat:             32.9,
						Carbohydrates:   0.1,
					},
				},
				TotalCalories:      875,
				TotalProtein:       24.7,
				TotalFat:           33.9,
				TotalCarbohydrates: 111.5,
			}, nil
		},
	}

	updateStatusCalled := 0
	saveResultCalled := false

	mockRepo := &MockAnalysisRepository{
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error {
			updateStatusCalled++
			if updateStatusCalled == 1 {
				assert.Equal(t, requestID, id)
				assert.Equal(t, repository.StatusProcessing, status)
				assert.Empty(t, errorMessage)
			}
			return nil
		},
		SaveResultFunc: func(ctx context.Context, id uuid.UUID, result *service.AnalysisResult) error {
			saveResultCalled = true
			assert.Equal(t, requestID, id)
			assert.NotNil(t, result)
			assert.Equal(t, 875.0, result.TotalCalories)
			return nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 5*time.Second)

	request := repository.AnalysisRequest{
		ID:        requestID,
		Status:    repository.StatusPending,
		InputType: repository.InputTypeText,
		InputText: inputText,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := worker.processRequest(context.Background(), request)

	assert.NoError(t, err)
	assert.Equal(t, 1, updateStatusCalled)
	assert.True(t, saveResultCalled)
}

func TestProcessRequest_TextInput_AnalysisError(t *testing.T) {
	requestID := uuid.New()
	inputText := "ご飯二杯"

	// AnalyzeFoodTextが失敗するケース
	analysisError := errors.New("Gemini API タイムアウト")
	mockService := &MockFoodService{
		AnalyzeFoodTextFunc: func(ctx context.Context, text string) (*service.AnalysisResult, error) {
			return nil, analysisError
		},
	}

	updateStatusCalled := 0
	saveResultCalled := false

	mockRepo := &MockAnalysisRepository{
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error {
			updateStatusCalled++
			if updateStatusCalled == 1 {
				assert.Equal(t, repository.StatusProcessing, status)
			} else if updateStatusCalled == 2 {
				assert.Equal(t, requestID, id)
				assert.Equal(t, repository.StatusFailed, status)
				assert.Contains(t, errorMessage, "Gemini API タイムアウト")
			}
			return nil
		},
		SaveResultFunc: func(ctx context.Context, id uuid.UUID, result *service.AnalysisResult) error {
			saveResultCalled = true
			return nil
		},
	}

	worker := NewAnalysisWorker(mockService, mockRepo, 5*time.Second)

	request := repository.AnalysisRequest{
		ID:        requestID,
		Status:    repository.StatusPending,
		InputType: repository.InputTypeText,
		InputText: inputText,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := worker.processRequest(context.Background(), request)

	assert.NoError(t, err)
	assert.Equal(t, 2, updateStatusCalled)
	assert.False(t, saveResultCalled)
}
