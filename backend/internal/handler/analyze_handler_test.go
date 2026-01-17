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

	"github.com/ryosuke-horie/asken/backend/internal/service"
	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
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

func TestAnalyzeHandler_Success(t *testing.T) {
	mockService := &MockFoodService{
		AnalyzeFoodImageFunc: func(ctx context.Context, imagePath string) (*service.AnalysisResult, error) {
			return &service.AnalysisResult{
				Foods: []gemini.NutritionInfo{
					{
						Name:            "刺身盛り合わせ",
						EstimatedAmount: "8切れ",
						Calories:        360.0,
						Protein:         30.0,
						Fat:             24.6,
						Carbohydrates:   0.4,
					},
				},
				TotalCalories:      360.0,
				TotalProtein:       30.0,
				TotalFat:           24.6,
				TotalCarbohydrates: 0.4,
			}, nil
		},
	}

	handler := NewAnalyzeHandler(mockService)

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
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result service.AnalysisResult
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Len(t, result.Foods, 1)
	assert.Equal(t, 360.0, result.TotalCalories)
}

func TestAnalyzeHandler_NoImageFile(t *testing.T) {
	mockService := &MockFoodService{}
	handler := NewAnalyzeHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeHandler_InvalidFileType(t *testing.T) {
	mockService := &MockFoodService{}
	handler := NewAnalyzeHandler(mockService)

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

func TestAnalyzeHandler_ServiceError(t *testing.T) {
	mockService := &MockFoodService{
		AnalyzeFoodImageFunc: func(ctx context.Context, imagePath string) (*service.AnalysisResult, error) {
			return nil, assert.AnError
		},
	}

	handler := NewAnalyzeHandler(mockService)

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
