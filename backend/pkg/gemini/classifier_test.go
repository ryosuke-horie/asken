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
		name     string
		filePath string
		data     []byte
		expected string
	}{
		{
			name:     "JPEGマジックバイト",
			filePath: "test.jpg",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expected: "image/jpeg",
		},
		{
			name:     "PNGマジックバイト",
			filePath: "test.png",
			data:     []byte{0x89, 0x50, 0x4E, 0x47},
			expected: "image/png",
		},
		{
			name:     "GIFマジックバイト",
			filePath: "test.gif",
			data:     []byte{0x47, 0x49, 0x46, 0x38},
			expected: "image/gif",
		},
		{
			name:     "WebPマジックバイト",
			filePath: "test.webp",
			data:     []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expected: "image/webp",
		},
		{
			name:     "拡張子フォールバック_JPG",
			filePath: "photo.JPG",
			data:     []byte{0x00, 0x00, 0x00, 0x00}, // 不正なマジックバイト
			expected: "image/jpeg",
		},
		{
			name:     "拡張子フォールバック_PNG",
			filePath: "image.PNG",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			expected: "image/png",
		},
		{
			name:     "未知の拡張子はJPEGにフォールバック",
			filePath: "file.unknown",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			expected: "image/jpeg",
		},
		{
			name:     "空のデータは拡張子で判定",
			filePath: "test.gif",
			data:     []byte{},
			expected: "image/gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMimeType(tt.filePath, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}
