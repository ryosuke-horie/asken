package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockReceiptParserClient はテスト用のモックReceiptParserClient
type MockReceiptParserClient struct {
	ParseReceiptImageFunc func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error)
}

func (m *MockReceiptParserClient) ParseReceiptImage(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
	if m.ParseReceiptImageFunc != nil {
		return m.ParseReceiptImageFunc(ctx, imageData, mimeType)
	}
	return nil, nil
}

// createReceiptMultipartRequest はレシートスキャン用のmultipartリクエストを生成するヘルパー
func createReceiptMultipartRequest(t *testing.T, imageData []byte, filename string, userID string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filename)
	require.NoError(t, err)
	_, err = part.Write(imageData)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/ingredients/scan-receipt", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if userID != "" {
		ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
		req = req.WithContext(ctx)
	}
	return req
}

func TestNewScanReceiptHandler_NilPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewScanReceiptHandler(nil)
	})
}

func TestScanReceiptHandler_MethodNotAllowed(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	req := httptest.NewRequest(http.MethodGet, "/api/ingredients/scan-receipt", nil)
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestScanReceiptHandler_Unauthorized(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	req := httptest.NewRequest(http.MethodPost, "/api/ingredients/scan-receipt", nil)
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestScanReceiptHandler_Handle_Success(t *testing.T) {
	mockIngredients := []gemini.ReceiptIngredient{
		{Name: "鶏むね肉", Category: "meat", Quantity: 300, Unit: "g"},
		{Name: "牛乳", Category: "dairy", Quantity: 1, Unit: "パック"},
	}

	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			assert.Equal(t, "image/jpeg", mimeType)
			return mockIngredients, nil
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestJPEGData(), "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ScanReceiptResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Ingredients, 2)

	assert.Equal(t, "鶏むね肉", resp.Ingredients[0].Name)
	assert.Equal(t, "meat", resp.Ingredients[0].Category)
	assert.Equal(t, float64(300), resp.Ingredients[0].Quantity)
	assert.Equal(t, "g", resp.Ingredients[0].Unit)
	assert.Equal(t, "receipt", resp.Ingredients[0].Source)

	assert.Equal(t, "牛乳", resp.Ingredients[1].Name)
	assert.Equal(t, "dairy", resp.Ingredients[1].Category)
}

func TestScanReceiptHandler_Handle_PNGSuccess(t *testing.T) {
	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			assert.Equal(t, "image/png", mimeType)
			return []gemini.ReceiptIngredient{
				{Name: "トマト", Category: "vegetable", Quantity: 3, Unit: "個"},
			}, nil
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestPNGData(), "receipt.png", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ScanReceiptResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Ingredients, 1)
	assert.Equal(t, "トマト", resp.Ingredients[0].Name)
}

func TestScanReceiptHandler_Handle_EmptyResult(t *testing.T) {
	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			return []gemini.ReceiptIngredient{}, nil
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestJPEGData(), "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ScanReceiptResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.Ingredients)
}

func TestScanReceiptHandler_Handle_NoImageField(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/ingredients/scan-receipt", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanReceiptHandler_Handle_UnsupportedFormat(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	// HEICファイルを送信（非対応）
	req := createReceiptMultipartRequest(t, createTestHEICData(), "receipt.heic", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "JPEG, PNGのみ")
}

func TestScanReceiptHandler_Handle_GeminiError(t *testing.T) {
	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			return nil, fmt.Errorf("Gemini API error")
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestJPEGData(), "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDetectReceiptImageMimeType(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		ext          string
		expectedMime string
		wantErr      bool
	}{
		{
			name:         "JPEGマジックバイト",
			data:         createTestJPEGData(),
			ext:          ".jpg",
			expectedMime: "image/jpeg",
		},
		{
			name:         "PNGマジックバイト",
			data:         createTestPNGData(),
			ext:          ".png",
			expectedMime: "image/png",
		},
		{
			name:         "拡張子フォールバック - JPEG",
			data:         make([]byte, 10),
			ext:          ".jpg",
			expectedMime: "image/jpeg",
		},
		{
			name:         "拡張子フォールバック - PNG",
			data:         make([]byte, 10),
			ext:          ".png",
			expectedMime: "image/png",
		},
		{
			name:    "データ不足",
			data:    []byte{0xFF},
			ext:     ".txt",
			wantErr: true,
		},
		{
			name:    "不明な形式",
			data:    make([]byte, 10),
			ext:     ".txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime, err := detectReceiptImageMimeType(tt.data, tt.ext)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMime, mime)
			}
		})
	}
}
