package testutil

import (
	"context"
	"io"
	"time"
)

// MockStorageRepository はStorageRepositoryのモック実装
type MockStorageRepository struct {
	UploadFunc       func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
	GetSignedURLFunc func(ctx context.Context, objectName string, expiration time.Duration) (string, error)
	DeleteFunc       func(ctx context.Context, objectName string) error
}

// Upload はアップロード操作のモック
func (m *MockStorageRepository) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file, filename, contentType)
	}
	return "uploads/test.jpg", nil
}

// GetSignedURL は署名付きURL取得のモック
func (m *MockStorageRepository) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	if m.GetSignedURLFunc != nil {
		return m.GetSignedURLFunc(ctx, objectName, expiration)
	}
	return "https://storage.example.com/signed-url", nil
}

// Delete は削除操作のモック
func (m *MockStorageRepository) Delete(ctx context.Context, objectName string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, objectName)
	}
	return nil
}
