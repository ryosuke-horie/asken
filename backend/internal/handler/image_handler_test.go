package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageHandler_Handle_Success(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用の画像ファイルを作成
	testFilename := "test-image.jpg"
	testFilePath := filepath.Join(tmpDir, testFilename)
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10} // JPEGヘッダー
	err := os.WriteFile(testFilePath, jpegData, 0644)
	require.NoError(t, err)

	handler := NewImageHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/images/"+testFilename, nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
}

func TestImageHandler_Handle_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	handler := NewImageHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/images/nonexistent.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestImageHandler_Handle_MethodNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()

	handler := NewImageHandler(tmpDir)

	req := httptest.NewRequest(http.MethodPost, "/api/images/test.jpg", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestImageHandler_Handle_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()

	handler := NewImageHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/images", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImageHandler_Handle_EmptyFilename(t *testing.T) {
	tmpDir := t.TempDir()

	handler := NewImageHandler(tmpDir)

	// パスの最後に空のセグメントがある場合
	req := httptest.NewRequest(http.MethodGet, "/api/images/", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// URLパースの結果によってはNotFoundまたはBadRequestになる可能性
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
}

func TestImageHandler_Handle_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// 親ディレクトリに秘密ファイルを作成
	secretDir := filepath.Join(tmpDir, "secret")
	require.NoError(t, os.MkdirAll(secretDir, 0755))
	secretFile := filepath.Join(secretDir, "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("secret data"), 0644))

	// uploadsディレクトリを作成
	uploadsDir := filepath.Join(tmpDir, "uploads")
	require.NoError(t, os.MkdirAll(uploadsDir, 0755))

	handler := NewImageHandler(uploadsDir)

	// パストラバーサル攻撃を試みる
	req := httptest.NewRequest(http.MethodGet, "/api/images/../secret/secret.txt", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	// アクセス拒否されるべき
	assert.Equal(t, http.StatusForbidden, w.Code)
}
