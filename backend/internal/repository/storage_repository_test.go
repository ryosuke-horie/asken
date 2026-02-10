package repository

import (
	"context"
	"errors"
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

// ============================================================================
// Cloud Storage Repository 単体テスト用モック
// ============================================================================

// mockStorageClient は storageClient のモック実装
type mockStorageClient struct {
	bucketFunc             func(name string) storageBucket
	bucketWithSignedURLFunc func(name string) storageBucketWithSignedURL
}

func (m *mockStorageClient) Bucket(name string) storageBucket {
	if m.bucketFunc != nil {
		return m.bucketFunc(name)
	}
	return &mockStorageBucket{}
}

func (m *mockStorageClient) BucketWithSignedURL(name string) storageBucketWithSignedURL {
	if m.bucketWithSignedURLFunc != nil {
		return m.bucketWithSignedURLFunc(name)
	}
	return &mockStorageBucketWithSignedURL{}
}

// mockStorageBucket は storageBucket のモック実装
type mockStorageBucket struct {
	objectFunc func(name string) storageObject
}

func (m *mockStorageBucket) Object(name string) storageObject {
	if m.objectFunc != nil {
		return m.objectFunc(name)
	}
	return &mockStorageObject{}
}

// mockStorageBucketWithSignedURL は storageBucketWithSignedURL のモック実装
type mockStorageBucketWithSignedURL struct {
	signedURLFunc func(objectName string, opts *storage.SignedURLOptions) (string, error)
}

func (m *mockStorageBucketWithSignedURL) SignedURL(objectName string, opts *storage.SignedURLOptions) (string, error) {
	if m.signedURLFunc != nil {
		return m.signedURLFunc(objectName, opts)
	}
	return "https://storage.example.com/signed-url", nil
}

// mockStorageObject は storageObject のモック実装
type mockStorageObject struct {
	newWriterFunc func(ctx context.Context) storageObjectWriter
	newReaderFunc func(ctx context.Context) (storageObjectReader, error)
	attrsFunc     func(ctx context.Context) (*storage.ObjectAttrs, error)
	deleteFunc    func(ctx context.Context) error
}

func (m *mockStorageObject) NewWriter(ctx context.Context) storageObjectWriter {
	if m.newWriterFunc != nil {
		return m.newWriterFunc(ctx)
	}
	return &mockStorageWriter{}
}

func (m *mockStorageObject) NewReader(ctx context.Context) (storageObjectReader, error) {
	if m.newReaderFunc != nil {
		return m.newReaderFunc(ctx)
	}
	return &mockStorageReader{data: []byte("test data")}, nil
}

func (m *mockStorageObject) Attrs(ctx context.Context) (*storage.ObjectAttrs, error) {
	if m.attrsFunc != nil {
		return m.attrsFunc(ctx)
	}
	return &storage.ObjectAttrs{}, nil
}

func (m *mockStorageObject) Delete(ctx context.Context) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx)
	}
	return nil
}

// mockStorageWriter は storageObjectWriter のモック実装
type mockStorageWriter struct {
	writeFunc   func(p []byte) (int, error)
	closeFunc   func() error
	contentType string
}

func (m *mockStorageWriter) Write(p []byte) (int, error) {
	if m.writeFunc != nil {
		return m.writeFunc(p)
	}
	return len(p), nil
}

func (m *mockStorageWriter) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockStorageWriter) SetContentType(contentType string) {
	m.contentType = contentType
}

// mockStorageReader は storageObjectReader のモック実装
type mockStorageReader struct {
	data       []byte
	readPos    int
	readFunc   func(p []byte) (int, error)
	closeFunc  func() error
	closeCalled bool
}

func (m *mockStorageReader) Read(p []byte) (int, error) {
	if m.readFunc != nil {
		return m.readFunc(p)
	}
	if m.readPos >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockStorageReader) Close() error {
	m.closeCalled = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// ============================================================================
// CloudStorageRepository Upload メソッドのテスト
// ============================================================================

func TestCloudStorageRepository_Upload_Success(t *testing.T) {
	writer := &mockStorageWriter{}
	object := &mockStorageObject{
		newWriterFunc: func(ctx context.Context) storageObjectWriter {
			return writer
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			assert.Equal(t, "test-bucket", name)
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	reader := strings.NewReader("test image data")
	result, err := repo.Upload(context.Background(), reader, "test.jpg", "image/jpeg")

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "uploads/"))
	assert.True(t, strings.HasSuffix(result, ".jpg"))
	assert.Equal(t, "image/jpeg", writer.contentType)
}

func TestCloudStorageRepository_Upload_CopyError(t *testing.T) {
	object := &mockStorageObject{
		newWriterFunc: func(ctx context.Context) storageObjectWriter {
			return &mockStorageWriter{
				writeFunc: func(p []byte) (int, error) {
					return 0, errors.New("write error")
				},
			}
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	reader := strings.NewReader("test image data")
	_, err := repo.Upload(context.Background(), reader, "test.jpg", "image/jpeg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud Storageへのアップロードに失敗")
}

func TestCloudStorageRepository_Upload_CloseError(t *testing.T) {
	object := &mockStorageObject{
		newWriterFunc: func(ctx context.Context) storageObjectWriter {
			return &mockStorageWriter{
				closeFunc: func() error {
					return errors.New("close error")
				},
			}
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	reader := strings.NewReader("test image data")
	_, err := repo.Upload(context.Background(), reader, "test.jpg", "image/jpeg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud Storageへのアップロード完了に失敗")
}

// ============================================================================
// CloudStorageRepository Download メソッドのテスト
// ============================================================================

func TestCloudStorageRepository_Download_Success(t *testing.T) {
	expectedData := []byte("test image data")
	object := &mockStorageObject{
		newReaderFunc: func(ctx context.Context) (storageObjectReader, error) {
			return &mockStorageReader{data: expectedData}, nil
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	data, err := repo.Download(context.Background(), "uploads/test-uuid.jpg")

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
}

func TestCloudStorageRepository_Download_ObjectNotFound(t *testing.T) {
	object := &mockStorageObject{
		newReaderFunc: func(ctx context.Context) (storageObjectReader, error) {
			return nil, storage.ErrObjectNotExist
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	_, err := repo.Download(context.Background(), "uploads/test-uuid.jpg")

	assert.Error(t, err)
	assert.Equal(t, ErrObjectNotFound, err)
}

func TestCloudStorageRepository_Download_ReadError(t *testing.T) {
	object := &mockStorageObject{
		newReaderFunc: func(ctx context.Context) (storageObjectReader, error) {
			return &mockStorageReader{
				readFunc: func(p []byte) (int, error) {
					return 0, errors.New("read error")
				},
			}, nil
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	_, err := repo.Download(context.Background(), "uploads/test-uuid.jpg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud Storageからのデータ読み取りに失敗")
}

func TestCloudStorageRepository_Download_SizeLimitExceeded(t *testing.T) {
	// 15MB + 1 バイトのデータを生成
	largeData := make([]byte, maxDownloadSize+1)
	for i := range largeData {
		largeData[i] = 'x'
	}

	object := &mockStorageObject{
		newReaderFunc: func(ctx context.Context) (storageObjectReader, error) {
			return &mockStorageReader{data: largeData}, nil
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	_, err := repo.Download(context.Background(), "uploads/large-file.jpg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ファイルサイズが上限")
}

// ============================================================================
// CloudStorageRepository GetSignedURL メソッドのテスト
// ============================================================================

func TestCloudStorageRepository_GetSignedURL_Success(t *testing.T) {
	expectedURL := "https://storage.googleapis.com/bucket/uploads/test-uuid.jpg?signature=xxx"

	object := &mockStorageObject{
		attrsFunc: func(ctx context.Context) (*storage.ObjectAttrs, error) {
			return &storage.ObjectAttrs{}, nil
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	bucketWithSignedURL := &mockStorageBucketWithSignedURL{
		signedURLFunc: func(objectName string, opts *storage.SignedURLOptions) (string, error) {
			assert.Equal(t, "uploads/test-uuid.jpg", objectName)
			assert.Equal(t, "GET", opts.Method)
			return expectedURL, nil
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
		bucketWithSignedURLFunc: func(name string) storageBucketWithSignedURL {
			return bucketWithSignedURL
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	url, err := repo.GetSignedURL(context.Background(), "uploads/test-uuid.jpg", 15*time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)
}

func TestCloudStorageRepository_GetSignedURL_ObjectNotFound(t *testing.T) {
	object := &mockStorageObject{
		attrsFunc: func(ctx context.Context) (*storage.ObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	_, err := repo.GetSignedURL(context.Background(), "uploads/test-uuid.jpg", 15*time.Minute)

	assert.Error(t, err)
	assert.Equal(t, ErrObjectNotFound, err)
}

func TestCloudStorageRepository_GetSignedURL_AttrsError(t *testing.T) {
	object := &mockStorageObject{
		attrsFunc: func(ctx context.Context) (*storage.ObjectAttrs, error) {
			return nil, errors.New("attrs error")
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	_, err := repo.GetSignedURL(context.Background(), "uploads/test-uuid.jpg", 15*time.Minute)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "オブジェクト情報の取得に失敗")
}

func TestCloudStorageRepository_GetSignedURL_SignedURLError(t *testing.T) {
	object := &mockStorageObject{
		attrsFunc: func(ctx context.Context) (*storage.ObjectAttrs, error) {
			return &storage.ObjectAttrs{}, nil
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	bucketWithSignedURL := &mockStorageBucketWithSignedURL{
		signedURLFunc: func(objectName string, opts *storage.SignedURLOptions) (string, error) {
			return "", errors.New("signed url error")
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
		bucketWithSignedURLFunc: func(name string) storageBucketWithSignedURL {
			return bucketWithSignedURL
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	_, err := repo.GetSignedURL(context.Background(), "uploads/test-uuid.jpg", 15*time.Minute)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "署名付きURLの生成に失敗")
}

// ============================================================================
// CloudStorageRepository Delete メソッドのテスト
// ============================================================================

func TestCloudStorageRepository_Delete_Success(t *testing.T) {
	deleteCalled := false
	object := &mockStorageObject{
		deleteFunc: func(ctx context.Context) error {
			deleteCalled = true
			return nil
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	err := repo.Delete(context.Background(), "uploads/test-uuid.jpg")

	assert.NoError(t, err)
	assert.True(t, deleteCalled)
}

func TestCloudStorageRepository_Delete_AlreadyDeleted(t *testing.T) {
	object := &mockStorageObject{
		deleteFunc: func(ctx context.Context) error {
			return storage.ErrObjectNotExist
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	err := repo.Delete(context.Background(), "uploads/test-uuid.jpg")

	// 既に削除済みの場合はエラーにならない
	assert.NoError(t, err)
}

func TestCloudStorageRepository_Delete_Error(t *testing.T) {
	object := &mockStorageObject{
		deleteFunc: func(ctx context.Context) error {
			return errors.New("delete error")
		},
	}
	bucket := &mockStorageBucket{
		objectFunc: func(name string) storageObject {
			return object
		},
	}
	client := &mockStorageClient{
		bucketFunc: func(name string) storageBucket {
			return bucket
		},
	}

	repo := &cloudStorageRepository{
		client:     client,
		bucketName: "test-bucket",
	}

	err := repo.Delete(context.Background(), "uploads/test-uuid.jpg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud Storageからの削除に失敗")
}
