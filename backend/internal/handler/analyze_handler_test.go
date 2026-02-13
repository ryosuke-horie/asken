package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/internal/testutil"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (m *MockAnalysisRepository) UpdateResult(ctx context.Context, userID string, id uuid.UUID, foods []gemini.NutritionInfo) error {
	if m.UpdateResultFunc != nil {
		return m.UpdateResultFunc(ctx, userID, id, foods)
	}
	return nil
}

func TestAnalyzeHandler_Success(t *testing.T) {
	// 非同期処理のため、FoodServiceは呼ばれない
	mockService := &MockFoodService{}

	expectedID := uuid.New()
	mockRepo := &MockAnalysisRepository{
		CreateRequestFunc: func(ctx context.Context, imagePath string, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
			// Cloud Storageのオブジェクト名が渡されることを確認
			assert.Contains(t, imagePath, "uploads/")
			return expectedID, nil
		},
	}
	mockStorageRepo := &testutil.MockStorageRepository{
		UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
			return "uploads/test-uuid.jpg", nil
		},
	}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

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
	_, err = part.Write(jpegData)
	require.NoError(t, err)
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
		ID string `json:"id"`
	}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedID.String(), response.ID)

	// ファイルが削除されていないことを確認（永続化のため）
	// Note: テスト後のクリーンアップは別途必要
}

func TestAnalyzeHandler_NoImageFile(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_InvalidFileType(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	// テキストファイルをアップロード
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("not an image"))
	require.NoError(t, err)
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
		CreateRequestFunc: func(ctx context.Context, imagePath string, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
			return uuid.Nil, assert.AnError
		},
	}
	mockStorageRepo := &testutil.MockStorageRepository{
		UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
			return "uploads/test-uuid.jpg", nil
		},
	}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	// JPEG magic number + 512バイト以上のダミーデータ
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	_, err = part.Write(jpegData)
	require.NoError(t, err)
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

func TestAnalyzeHandler_TextInput_Success(t *testing.T) {
	mockService := &MockFoodService{}

	expectedID := uuid.New()
	mockRepo := &MockAnalysisRepository{
		CreateRequestWithTextFunc: func(ctx context.Context, inputText string, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
			assert.Equal(t, "ご飯二杯, 焼肉", inputText)
			assert.Equal(t, "lunch", mealType)
			assert.Equal(t, "2024-01-15", mealDate)
			return expectedID, nil
		},
	}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	// JSONリクエストボディを作成
	reqBody := map[string]string{
		"input_text": "ご飯二杯, 焼肉",
		"meal_type":  "lunch",
		"meal_date":  "2024-01-15",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedID.String(), response.ID)
}

func TestAnalyzeHandler_TextInput_EmptyText(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	reqBody := map[string]string{
		"input_text": "",
		"meal_type":  "lunch",
		"meal_date":  "2024-01-15",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_TextInput_InvalidMealType(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	reqBody := map[string]string{
		"input_text": "ご飯二杯",
		"meal_type":  "invalid",
		"meal_date":  "2024-01-15",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_TextInput_TooLong(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	// 1001文字のテキストを作成
	longText := make([]byte, 1001)
	for i := range longText {
		longText[i] = 'a'
	}

	reqBody := map[string]string{
		"input_text": string(longText),
		"meal_type":  "lunch",
		"meal_date":  "2024-01-15",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_TextInput_InvalidMealDate(t *testing.T) {
	tests := []struct {
		name     string
		mealDate string
	}{
		{"不正なフォーマット", "2024/01/15"},
		{"日付のみ不正", "2024-13-01"},
		{"テキスト", "yesterday"},
		{"日付区切り不正", "20240115"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockFoodService{}
			mockRepo := &MockAnalysisRepository{}
			mockStorageRepo := &testutil.MockStorageRepository{}
			handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

			reqBody := map[string]string{
				"input_text": "ご飯",
				"meal_type":  "lunch",
				"meal_date":  tt.mealDate,
			}
			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Handle(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "YYYY-MM-DD")
		})
	}
}

func TestAnalyzeHandler_TextInput_MalformedJSON(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	// 不正なJSONを送信
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_StorageUploadError(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{
		UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
			return "", errors.New("Cloud Storage unavailable")
		},
	}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	// JPEG magic number + 512バイト以上のダミーデータ
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	_, err = part.Write(jpegData)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("meal_type", "lunch"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "ファイルの保存に失敗しました")
}

func TestAnalyzeHandler_CleanupOnRepositoryFailure(t *testing.T) {
	mockService := &MockFoodService{}
	deleteCalled := false
	uploadedObjectName := ""

	mockStorageRepo := &testutil.MockStorageRepository{
		UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
			uploadedObjectName = "uploads/test-uuid.jpg"
			return uploadedObjectName, nil
		},
		DeleteFunc: func(ctx context.Context, objectName string) error {
			deleteCalled = true
			assert.Equal(t, uploadedObjectName, objectName)
			return nil
		},
	}
	mockRepo := &MockAnalysisRepository{
		CreateRequestFunc: func(ctx context.Context, imagePath string, mealType string, mealDate string, userID *string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("database error")
		},
	}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	_, err = part.Write(jpegData)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("meal_type", "lunch"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, deleteCalled, "Cloud Storage delete should be called on repository failure")
}

func TestAnalyzeHandler_TextInput_OversizedBody(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	// 4KBを超えるJSONボディを作成
	largeBody := make([]byte, 5000)
	for i := range largeBody {
		largeBody[i] = 'a'
	}
	reqBody := map[string]string{
		"input_text": string(largeBody),
		"meal_type":  "lunch",
		"meal_date":  "2024-01-15",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), "リクエストボディが大きすぎます")
}

func TestTruncateForLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		expected string
	}{
		{"短いテキスト", "hello", 50, "hello"},
		{"ちょうど上限", "12345", 5, "12345"},
		{"上限超え", "123456", 5, "12345..."},
		{"日本語テキスト", "あいうえおかきくけこ", 5, "あいうえお..."},
		{"空文字列", "", 50, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateForLog(tt.input, tt.maxRunes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzeHandler_HandleUploadImage_StorageError(t *testing.T) {
	mockService := &MockFoodService{}
	mockRepo := &MockAnalysisRepository{}
	mockStorageRepo := &testutil.MockStorageRepository{
		UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
			return "", errors.New("Cloud Storage unavailable")
		},
	}
	handler := NewAnalyzeHandler(mockService, mockRepo, mockStorageRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	_, err = part.Write(jpegData)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload-image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.HandleUploadImage(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "ファイルの保存に失敗しました")
}
