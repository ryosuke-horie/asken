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
	return []gemini.ReceiptIngredient{}, nil
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
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

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

func TestScanReceiptHandler_Handle_FileTooLarge(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	// 10MB + 1バイトのデータを作成（JPEGマジックバイト付き）
	largeData := make([]byte, 10<<20+1)
	largeData[0] = 0xFF
	largeData[1] = 0xD8
	largeData[2] = 0xFF

	req := createReceiptMultipartRequest(t, largeData, "large.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanReceiptHandler_Handle_InvalidMultipartBody(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	req := httptest.NewRequest(http.MethodPost, "/api/ingredients/scan-receipt", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=broken")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanReceiptHandler_Handle_InvalidImageData(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	// 3バイト未満のデータ（マジックバイト判定不能）
	req := createReceiptMultipartRequest(t, []byte{0xFF}, "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ファイルの形式が正しくありません")
}

func TestScanReceiptHandler_Handle_JpegExtensionSuccess(t *testing.T) {
	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			assert.Equal(t, "image/jpeg", mimeType)
			return []gemini.ReceiptIngredient{
				{Name: "卵", Category: "other", Quantity: 6, Unit: "個"},
			}, nil
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestJPEGData(), "receipt.jpeg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ScanReceiptResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Ingredients, 1)
	assert.Equal(t, "卵", resp.Ingredients[0].Name)
}

func TestScanReceiptHandler_Handle_ExtensionMagicByteMismatch(t *testing.T) {
	h := NewScanReceiptHandler(&MockReceiptParserClient{})

	// .jpg拡張子だが内容がJPEGでもPNGでもないデータ（拡張子チェックは通過するがMIME判定で失敗）
	invalidData := make([]byte, 20) // ゼロバイト列（JPEG/PNGマジックバイトなし）
	req := createReceiptMultipartRequest(t, invalidData, "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ファイルの形式が正しくありません")
}

func TestScanReceiptHandler_Handle_ContextCanceled(t *testing.T) {
	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			return nil, fmt.Errorf("operation aborted: %w", context.Canceled)
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestJPEGData(), "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	// クライアント切断時はレスポンスボディを書かずに終了する（HTTPデフォルトの200）
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestScanReceiptHandler_Handle_ContextDeadlineExceeded(t *testing.T) {
	mock := &MockReceiptParserClient{
		ParseReceiptImageFunc: func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error) {
			return nil, fmt.Errorf("timeout waiting for response: %w", context.DeadlineExceeded)
		},
	}

	h := NewScanReceiptHandler(mock)
	req := createReceiptMultipartRequest(t, createTestJPEGData(), "receipt.jpg", "test-user")
	w := httptest.NewRecorder()
	h.Handle(w, req)

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "タイムアウト")
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
		expectedMime string
		wantErr      bool
	}{
		{
			name:         "JPEGマジックバイト",
			data:         createTestJPEGData(),
			expectedMime: "image/jpeg",
		},
		{
			name:         "PNGマジックバイト",
			data:         createTestPNGData(),
			expectedMime: "image/png",
		},
		{
			name:         "3バイトJPEGマジックバイト - 最小サイズ",
			data:         []byte{0xFF, 0xD8, 0xFF},
			expectedMime: "image/jpeg",
		},
		{
			name:    "2バイト - データ不足",
			data:    []byte{0xFF, 0xD8},
			wantErr: true,
		},
		{
			name:    "1バイト - データ不足",
			data:    []byte{0xFF},
			wantErr: true,
		},
		{
			name:    "マジックバイト不一致（ゼロバイト列はJPEGでもPNGでもない）",
			data:    make([]byte, 10),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime, err := detectReceiptImageMimeType(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMime, mime)
			}
		})
	}
}
