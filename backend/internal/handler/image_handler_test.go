package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestImageHandler_Handle_Success(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{
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
	mockStorageRepo := &testutil.MockStorageRepository{
		GetSignedURLFunc: func(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
			return "", repository.ErrObjectNotFound
		},
	}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/images/nonexistent.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestImageHandler_Handle_InternalError(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{
		GetSignedURLFunc: func(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
			return "", errors.New("service temporarily unavailable")
		},
	}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/images/test-image.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to generate image URL")
}

func TestImageHandler_Handle_MethodNotAllowed(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/images/test.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestImageHandler_Handle_InvalidURL(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{}

	handler := NewImageHandler(mockStorageRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/images", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImageHandler_Handle_EmptyFilename(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{}

	handler := NewImageHandler(mockStorageRepo)

	// パスの最後に空のセグメントがある場合
	req := httptest.NewRequest(http.MethodGet, "/api/images/", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// URLパースの結果によってはNotFoundまたはBadRequestになる可能性
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
}

func TestImageHandler_Handle_PathTraversal(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{}

	handler := NewImageHandler(mockStorageRepo)

	// パストラバーサル攻撃を試みる
	req := httptest.NewRequest(http.MethodGet, "/api/images/../secret/secret.txt", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// アクセス拒否されるべき
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestImageHandler_Handle_PathTraversal_SlashOnly(t *testing.T) {
	mockStorageRepo := &testutil.MockStorageRepository{}

	handler := NewImageHandler(mockStorageRepo)

	// スラッシュを含むファイル名（サブディレクトリへのアクセス試行）
	req := httptest.NewRequest(http.MethodGet, "/api/images/subdir/file.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// スラッシュを含むパスはアクセス拒否されるべき
	assert.Equal(t, http.StatusForbidden, w.Code)
}

