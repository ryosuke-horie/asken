package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyFoods_Success(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// テスト画像のパスを取得（pkg/gemini/testdata内）
	imagePath := filepath.Join("testdata", "images", "IMG_0374.JPG")

	// 画像の存在確認
	_, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		t.Skipf("テスト画像が見つかりません: %s", imagePath)
	}

	classifier := NewClassifier(60 * time.Second)
	ctx := context.Background()

	foods, err := classifier.ClassifyFoods(ctx, imagePath)

	require.NoError(t, err)
	assert.NotEmpty(t, foods)

	// 各料理が適切な構造を持っているか確認
	for _, food := range foods {
		assert.NotEmpty(t, food.Name, "料理名が空です")
		assert.NotEmpty(t, food.EstimatedAmount, "推定量が空です")
	}
}

func TestClassifyFoods_InvalidImagePath(t *testing.T) {
	classifier := NewClassifier(60 * time.Second)
	ctx := context.Background()

	// 存在しない画像パス
	invalidPath := "/path/to/nonexistent/image.jpg"

	_, err := classifier.ClassifyFoods(ctx, invalidPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "画像ファイルが見つかりません")
}

func TestClassifyFoods_UnsupportedImageFormat(t *testing.T) {
	// テスト用の一時ファイルを作成（サポート外の拡張子）
	tmpFile, err := os.CreateTemp("", "test*.bmp")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// BMPのマジックバイトを書き込む
	_, err = tmpFile.Write([]byte{0x42, 0x4D, 0x00, 0x00})
	require.NoError(t, err)
	tmpFile.Close()

	classifier := NewClassifier(60 * time.Second)
	ctx := context.Background()

	_, err = classifier.ClassifyFoods(ctx, tmpFile.Name())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "サポートされていない画像形式")
}

func TestClassifyFoods_EmptyResponse(t *testing.T) {
	// このテストは実際のGemini APIの挙動に依存するためスキップ
	// 実際には、Gemini APIが空のレスポンスを返すことは稀
	t.Skip("Gemini APIの挙動に依存するためスキップ")
}

func TestClassifyFoods_Timeout(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// 非常に短いタイムアウトでテスト
	imagePath := filepath.Join("testdata", "images", "IMG_0374.JPG")

	_, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		t.Skipf("テスト画像が見つかりません: %s", imagePath)
	}

	classifier := NewClassifier(1 * time.Millisecond)
	ctx := context.Background()

	_, err = classifier.ClassifyFoods(ctx, imagePath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}

func TestClassifyFoods_RelativePath(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// 相対パスが絶対パスに変換されることを確認
	imagePath := filepath.Join("testdata", "images", "IMG_0374.JPG")

	_, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		t.Skipf("テスト画像が見つかりません: %s", imagePath)
	}

	classifier := NewClassifier(60 * time.Second)
	ctx := context.Background()

	foods, err := classifier.ClassifyFoods(ctx, imagePath)

	require.NoError(t, err)
	assert.NotEmpty(t, foods)
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		data        []byte
		expected    string
		expectError bool
	}{
		{
			name:        "JPEGマジックバイト",
			filePath:    "test.jpg",
			data:        []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expected:    "image/jpeg",
			expectError: false,
		},
		{
			name:        "PNGマジックバイト",
			filePath:    "test.png",
			data:        []byte{0x89, 0x50, 0x4E, 0x47},
			expected:    "image/png",
			expectError: false,
		},
		{
			name:        "GIFマジックバイト",
			filePath:    "test.gif",
			data:        []byte{0x47, 0x49, 0x46, 0x38},
			expected:    "image/gif",
			expectError: false,
		},
		{
			name:        "WebPマジックバイト",
			filePath:    "test.webp",
			data:        []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expected:    "image/webp",
			expectError: false,
		},
		{
			name:        "拡張子フォールバック_JPG",
			filePath:    "photo.JPG",
			data:        []byte{0x00, 0x00, 0x00, 0x00}, // 不正なマジックバイト
			expected:    "image/jpeg",
			expectError: false,
		},
		{
			name:        "拡張子フォールバック_PNG",
			filePath:    "image.PNG",
			data:        []byte{0x00, 0x00, 0x00, 0x00},
			expected:    "image/png",
			expectError: false,
		},
		{
			name:        "未知の拡張子はエラーを返す",
			filePath:    "file.unknown",
			data:        []byte{0x00, 0x00, 0x00, 0x00},
			expected:    "",
			expectError: true,
		},
		{
			name:        "空のデータは拡張子で判定",
			filePath:    "test.gif",
			data:        []byte{},
			expected:    "image/gif",
			expectError: false,
		},
		{
			name:        "BMPはサポート外でエラー",
			filePath:    "image.bmp",
			data:        []byte{0x42, 0x4D, 0x00, 0x00}, // BMP マジックバイト
			expected:    "",
			expectError: true,
		},
		{
			name:        "TIFFはサポート外でエラー",
			filePath:    "image.tiff",
			data:        []byte{0x49, 0x49, 0x2A, 0x00}, // TIFF (little-endian)
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detectMimeType(tt.filePath, tt.data)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
				assert.Contains(t, err.Error(), "サポートされていない画像形式")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
