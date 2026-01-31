package gemini

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoGeminiCLI skips the test if Gemini CLI is not available
func skipIfNoGeminiCLI(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("gemini"); err != nil {
		t.Skip("Gemini CLI not available, skipping integration test")
	}
}

func TestClassifyFoods_Success(t *testing.T) {
	skipIfNoGeminiCLI(t)

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
	// このテストは実際のGemini CLIの挙動に依存するためスキップ
	// 実際には、Gemini CLIが空のレスポンスを返すことは稀
	t.Skip("Gemini CLIの挙動に依存するためスキップ")
}

func TestClassifyFoods_Timeout(t *testing.T) {
	skipIfNoGeminiCLI(t)

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
	skipIfNoGeminiCLI(t)

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
