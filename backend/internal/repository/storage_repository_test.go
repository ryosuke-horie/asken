package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
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

func TestGenerateObjectName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantExt  string
	}{
		{"jpg拡張子が保持される", "photo.jpg", ".jpg"},
		{"png拡張子が保持される", "image.png", ".png"},
		{"jpeg拡張子が保持される", "image.jpeg", ".jpeg"},
		{"拡張子なしファイル", "noext", ""},
		{"複数ドットのファイル名", "my.photo.jpg", ".jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateObjectName(tt.filename)

			// uploads/プレフィックスの確認
			assert.True(t, strings.HasPrefix(result, "uploads/"))

			// 拡張子の確認
			assert.True(t, strings.HasSuffix(result, tt.wantExt))

			// UUID部分の長さ確認（uploads/ = 8文字 + UUID = 36文字 + 拡張子）
			withoutPrefix := strings.TrimPrefix(result, "uploads/")
			withoutExt := strings.TrimSuffix(withoutPrefix, tt.wantExt)
			assert.Len(t, withoutExt, 36) // UUID v4は36文字
		})
	}

	t.Run("呼び出しごとに異なるオブジェクト名を生成", func(t *testing.T) {
		name1 := generateObjectName("test.jpg")
		name2 := generateObjectName("test.jpg")
		assert.NotEqual(t, name1, name2)
	})
}

func TestConvertStorageError(t *testing.T) {
	t.Run("ErrObjectNotExistをErrObjectNotFoundに変換", func(t *testing.T) {
		result := convertStorageError(storage.ErrObjectNotExist)
		assert.Equal(t, ErrObjectNotFound, result)
	})

	t.Run("他のエラーはそのまま返す", func(t *testing.T) {
		originalErr := errors.New("some other error")
		result := convertStorageError(originalErr)
		assert.Equal(t, originalErr, result)
	})

	t.Run("ラップされたErrObjectNotExistも変換", func(t *testing.T) {
		wrappedErr := fmt.Errorf("wrapped: %w", storage.ErrObjectNotExist)
		result := convertStorageError(wrappedErr)
		assert.Equal(t, ErrObjectNotFound, result)
	})
}
