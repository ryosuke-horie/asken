package repository

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockStorageRepository はテスト用のモックStorageRepository
type MockStorageRepository struct {
	UploadFunc       func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
	GetSignedURLFunc func(ctx context.Context, objectName string, expiration time.Duration) (string, error)
	DeleteFunc       func(ctx context.Context, objectName string) error
}

func (m *MockStorageRepository) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file, filename, contentType)
	}
	return "uploads/test-uuid.jpg", nil
}

func (m *MockStorageRepository) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	if m.GetSignedURLFunc != nil {
		return m.GetSignedURLFunc(ctx, objectName, expiration)
	}
	return "https://storage.googleapis.com/bucket/uploads/test-uuid.jpg?signature=xxx", nil
}

func (m *MockStorageRepository) Delete(ctx context.Context, objectName string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, objectName)
	}
	return nil
}

func TestMockStorageRepository_Upload(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		mockReturn  string
		mockError   error
		wantErr     bool
	}{
		{
			name:        "正常なアップロード",
			filename:    "test.jpg",
			contentType: "image/jpeg",
			mockReturn:  "uploads/uuid-123.jpg",
			wantErr:     false,
		},
		{
			name:        "PNG画像のアップロード",
			filename:    "image.png",
			contentType: "image/png",
			mockReturn:  "uploads/uuid-456.png",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStorageRepository{
				UploadFunc: func(_ context.Context, _ io.Reader, _ string, _ string) (string, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			reader := strings.NewReader("test image data")
			result, err := mock.Upload(context.Background(), reader, tt.filename, tt.contentType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
		})
	}
}

func TestMockStorageRepository_GetSignedURL(t *testing.T) {
	tests := []struct {
		name       string
		objectName string
		expiration time.Duration
		mockReturn string
		wantErr    bool
	}{
		{
			name:       "正常な署名付きURL生成",
			objectName: "uploads/test-uuid.jpg",
			expiration: 15 * time.Minute,
			mockReturn: "https://storage.googleapis.com/bucket/uploads/test-uuid.jpg?signature=xxx",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStorageRepository{
				GetSignedURLFunc: func(_ context.Context, _ string, _ time.Duration) (string, error) {
					return tt.mockReturn, nil
				},
			}

			result, err := mock.GetSignedURL(context.Background(), tt.objectName, tt.expiration)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
		})
	}
}

func TestMockStorageRepository_Delete(t *testing.T) {
	tests := []struct {
		name       string
		objectName string
		wantErr    bool
	}{
		{
			name:       "正常な削除",
			objectName: "uploads/test-uuid.jpg",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStorageRepository{}

			err := mock.Delete(context.Background(), tt.objectName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
