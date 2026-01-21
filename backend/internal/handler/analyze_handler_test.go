package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockFoodService はテスト用のモックFoodService
type MockFoodService struct {
	AnalyzeFoodImageFunc func(ctx context.Context, imagePath string) (*service.AnalysisResult, error)
}

func (m *MockFoodService) AnalyzeFoodImage(ctx context.Context, imagePath string) (*service.AnalysisResult, error) {
	if m.AnalyzeFoodImageFunc != nil {
		return m.AnalyzeFoodImageFunc(ctx, imagePath)
	}
	return nil, nil
}

// MockAnalysisRepository はテスト用のモックAnalysisRepository
type MockAnalysisRepository struct {
	CreateRequestFunc      func(ctx context.Context, imagePath string, mealType string, mealDate string) (uuid.UUID, error)
	GetRequestFunc         func(ctx context.Context, id uuid.UUID) (*repository.AnalysisRequest, error)
	UpdateStatusFunc       func(ctx context.Context, id uuid.UUID, status repository.AnalysisStatus, errorMessage string) error
	SaveResultFunc         func(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error
	GetResultFunc          func(ctx context.Context, requestID uuid.UUID) (*service.AnalysisResult, error)
	GetPendingRequestsFunc func(ctx context.Context, limit int) ([]repository.AnalysisRequest, error)
	GetHistoryListFunc     func(ctx context.Context, page, limit int) ([]repository.HistoryItem, int, error)
	GetHistoryDetailFunc   func(ctx context.Context, id uuid.UUID) (*repository.HistoryDetail, error)
	DeleteHistoryFunc      func(ctx context.Context, id uuid.UUID) error
	GetDailyMealsFunc      func(ctx context.Context, date string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error)
}

func (m *MockAnalysisRepository) CreateRequest(ctx context.Context, imagePath string, mealType string, mealDate string) (uuid.UUID, error) {
	if m.CreateRequestFunc != nil {
		return m.CreateRequestFunc(ctx, imagePath, mealType, mealDate)
	}
	return uuid.Nil, nil
}

func (m *MockAnalysisRepository) GetRequest(ctx context.Context, id uuid.UUID) (*repository.AnalysisRequest, error) {
	if m.GetRequestFunc != nil {
		return m.GetRequestFunc(ctx, id)
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

func (m *MockAnalysisRepository) GetResult(ctx context.Context, requestID uuid.UUID) (*service.AnalysisResult, error) {
	if m.GetResultFunc != nil {
		return m.GetResultFunc(ctx, requestID)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) GetPendingRequests(ctx context.Context, limit int) ([]repository.AnalysisRequest, error) {
	if m.GetPendingRequestsFunc != nil {
		return m.GetPendingRequestsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) GetHistoryList(ctx context.Context, page, limit int) ([]repository.HistoryItem, int, error) {
	if m.GetHistoryListFunc != nil {
		return m.GetHistoryListFunc(ctx, page, limit)
	}
	return nil, 0, nil
}

func (m *MockAnalysisRepository) GetHistoryDetail(ctx context.Context, id uuid.UUID) (*repository.HistoryDetail, error) {
	if m.GetHistoryDetailFunc != nil {
		return m.GetHistoryDetailFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockAnalysisRepository) DeleteHistory(ctx context.Context, id uuid.UUID) error {
	if m.DeleteHistoryFunc != nil {
		return m.DeleteHistoryFunc(ctx, id)
	}
	return nil
}

func (m *MockAnalysisRepository) GetDailyMeals(ctx context.Context, date string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
	if m.GetDailyMealsFunc != nil {
		return m.GetDailyMealsFunc(ctx, date)
	}
	return nil, repository.DailyTotal{}, nil
}

func TestAnalyzeHandler_Success(t *testing.T) {
	// 非同期処理のため、FoodServiceは呼ばれない
	mockService := &MockFoodService{}

	expectedID := uuid.New()
	mockRepo := &MockAnalysisRepository{
		CreateRequestFunc: func(ctx context.Context, imagePath string, mealType string, mealDate string) (uuid.UUID, error) {
			// ファイルが永続化されていることを確認
			assert.FileExists(t, imagePath)
			return expectedID, nil
		},
	}

	handler := NewAnalyzeHandler(mockService, mockRepo)

	// テスト用の画像ファイルを作成（JPEGマジックナンバーを含む）
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	// JPEG magic number + 512バイト以上のダミーデータ
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	part.Write(jpegData)
	// meal_typeフィールドを追加
	require.NoError(t, writer.WriteField("meal_type", "lunch"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// 202 Accepted レスポンスを確認
	assert.Equal(t, http.StatusAccepted, w.Code)

	// レスポンス形式を確認
	var response struct {
		AnalysisID string `json:"analysis_id"`
	}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedID.String(), response.AnalysisID)

	// ファイルが削除されていないことを確認（永続化のため）
	// Note: テスト後のクリーンアップは別途必要
}

func TestAnalyzeHandler_NoImageFile(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_InvalidFileType(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo)

	// テキストファイルをアップロード
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.txt")
	require.NoError(t, err)
	part.Write([]byte("not an image"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_RepositoryError(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{
		CreateRequestFunc: func(ctx context.Context, imagePath string, mealType string, mealDate string) (uuid.UUID, error) {
			return uuid.Nil, assert.AnError
		},
	}

	handler := NewAnalyzeHandler(mockService, mockRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	// JPEG magic number + 512バイト以上のダミーデータ
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	part.Write(jpegData)
	// meal_typeフィールドを追加
	require.NoError(t, writer.WriteField("meal_type", "lunch"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestValidateImageFile(t *testing.T) {
	testCases := []struct {
		name      string
		filename  string
		content   []byte
		expectErr bool
	}{
		{
			name:     "JPEG画像",
			filename: "test.jpg",
			content: func() []byte {
				data := make([]byte, 512)
				data[0] = 0xFF
				data[1] = 0xD8
				data[2] = 0xFF
				return data
			}(),
			expectErr: false,
		},
		{
			name:     "PNG画像",
			filename: "test.png",
			content: func() []byte {
				data := make([]byte, 512)
				data[0] = 0x89
				data[1] = 0x50
				data[2] = 0x4E
				data[3] = 0x47
				data[4] = 0x0D
				data[5] = 0x0A
				data[6] = 0x1A
				data[7] = 0x0A
				return data
			}(),
			expectErr: false,
		},
		{
			name:      "テキストファイル",
			filename:  "test.txt",
			content:   []byte("not an image"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 一時ファイルを作成
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tc.filename)
			err := os.WriteFile(tmpFile, tc.content, 0644)
			require.NoError(t, err)

			// ファイルを開く
			file, err := os.Open(tmpFile)
			require.NoError(t, err)
			defer file.Close()

			// ファイルヘッダーを作成
			header := &multipart.FileHeader{
				Filename: tc.filename,
				Size:     int64(len(tc.content)),
			}

			err = validateImageFile(file, header)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
