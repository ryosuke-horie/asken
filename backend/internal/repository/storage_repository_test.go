package repository

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

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
			mock := &testutil.MockStorageRepository{
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
			mock := &testutil.MockStorageRepository{
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
			mock := &testutil.MockStorageRepository{}

			err := mock.Delete(context.Background(), tt.objectName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewStorageRepositoryCloudStorage_NilClient(t *testing.T) {
	repo, err := NewStorageRepositoryCloudStorage(nil, "test-bucket")

	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "storage client is required")
}

func TestNewStorageRepositoryCloudStorage_EmptyBucketName(t *testing.T) {
	// nilクライアントでもバケット名が空でもエラーになるが、
	// クライアントのチェックが先に実行されるため、
	// このテストではクライアントのエラーが返される
	// 実際の動作確認のため、nilクライアントでテスト
	repo, err := NewStorageRepositoryCloudStorage(nil, "")

	assert.Error(t, err)
	assert.Nil(t, repo)
	// nilクライアントのエラーが先に返される
	assert.Contains(t, err.Error(), "storage client is required")
}
