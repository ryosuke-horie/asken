package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockStorageRepositoryForImage はテスト用のモックStorageRepository
type MockStorageRepositoryForImage struct {
	UploadFunc       func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
	GetSignedURLFunc func(ctx context.Context, objectName string, expiration time.Duration) (string, error)
	DeleteFunc       func(ctx context.Context, objectName string) error
}

func (m *MockStorageRepositoryForImage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file, filename, contentType)
	}
	return "uploads/test-uuid.jpg", nil
}

func (m *MockStorageRepositoryForImage) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	if m.GetSignedURLFunc != nil {
		return m.GetSignedURLFunc(ctx, objectName, expiration)
	}
	return "https://storage.googleapis.com/bucket/uploads/test-uuid.jpg?signature=xxx", nil
}

func (m *MockStorageRepositoryForImage) Delete(ctx context.Context, objectName string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, objectName)
	}
	return nil
}

func TestImageHandler_Handle_Success(t *testing.T) {
	mockStorageRepo := &MockStorageRepositoryForImage{
		GetSignedURLFunc: func(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
			assert.Equal(t, "uploads/test-image.jpg", objectName)
			assert.Equal(t, 15*time.Minute, expiration)
			return "https://storage.googleapis.com/bucket/uploads/test-image.jpg?signature=xxx", nil
		},
	}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/images/test-image.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// 署名付きURLへのリダイレクト（302 Found）
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://storage.googleapis.com/bucket/uploads/test-image.jpg?signature=xxx", w.Header().Get("Location"))
}

func TestImageHandler_Handle_NotFound(t *testing.T) {
	mockStorageRepo := &MockStorageRepositoryForImage{
		GetSignedURLFunc: func(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
			return "", errors.New("オブジェクトが見つかりません: " + objectName)
		},
	}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/images/nonexistent.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestImageHandler_Handle_MethodNotAllowed(t *testing.T) {
	mockStorageRepo := &MockStorageRepositoryForImage{}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/images/test.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestImageHandler_Handle_InvalidURL(t *testing.T) {
	mockStorageRepo := &MockStorageRepositoryForImage{}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/images", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImageHandler_Handle_EmptyFilename(t *testing.T) {
	mockStorageRepo := &MockStorageRepositoryForImage{}

	handler := NewImageHandler(mockStorageRepo)

	// パスの最後に空のセグメントがある場合
	req := httptest.NewRequest(http.MethodGet, "/api/images/", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// URLパースの結果によってはNotFoundまたはBadRequestになる可能性
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
}

func TestImageHandler_Handle_PathTraversal(t *testing.T) {
	mockStorageRepo := &MockStorageRepositoryForImage{}

	handler := NewImageHandler(mockStorageRepo)

	// パストラバーサル攻撃を試みる
	req := httptest.NewRequest(http.MethodGet, "/api/images/../secret/secret.txt", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// アクセス拒否されるべき
	assert.Equal(t, http.StatusForbidden, w.Code)
}

