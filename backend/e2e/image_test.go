//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testImageData は1x1ピクセルの透明なPNG画像データ
var testImageData = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // width=1, height=1
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, // bit depth=8, color type=6 (RGBA)
	0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, // IDAT chunk
	0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, // compressed image data
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, // CRC
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, // IEND chunk
	0x42, 0x60, 0x82, // CRC
}

func TestUploadImage_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.UploadImage(ctx, "/api/upload-image", testImageData, "test.png")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]string
	err = resp.JSON(&body)
	require.NoError(t, err)
	assert.NotEmpty(t, body["image_path"], "Response should contain image_path")
}

func TestUploadImage_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 認証なしでリクエスト
	resp, err := testClient.UploadImage(ctx, "/api/upload-image", testImageData, "test.png")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetImage_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず画像をアップロード
	uploadResp, err := client.UploadImage(ctx, "/api/upload-image", testImageData, "test.png")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, uploadResp.StatusCode)

	var uploadBody map[string]string
	err = uploadResp.JSON(&uploadBody)
	require.NoError(t, err)
	imagePath := uploadBody["image_path"]
	require.NotEmpty(t, imagePath)

	// アップロードした画像を取得（認証必須）
	getResp, err := client.Get(ctx, "/api/images/"+imagePath)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "image/")
	assert.NotEmpty(t, getResp.Body, "Response should contain image data")
}

func TestGetImage_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 10*time.Second)

	// 存在しないファイルを取得（認証必須エンドポイント）
	resp, err := client.Get(ctx, "/api/images/nonexistent.png")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetImage_PathTraversal(t *testing.T) {
	client, ctx := authenticatedClient(t, 10*time.Second)

	// ".."を含むファイル名でパストラバーサルを試行
	// GoのHTTPクライアントはURL正規化で"/../"を除去するため、
	// ファイル名に".."を含むパターンでテスト
	resp, err := client.Get(ctx, "/api/images/..etc..passwd")
	require.NoError(t, err)

	// サーバーは403 Forbidden、400 Bad Request、または404 Not Foundを返すべき
	assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
		"Expected 403, 400 or 404 for path traversal attempt, got %d", resp.StatusCode)
}

func TestUploadImage_FileSizeTooLarge(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 11MBの画像データを作成（10MB制限を超える）
	largeImageData := make([]byte, 11<<20) // 11MB
	copy(largeImageData, testImageData)    // 先頭に有効なPNGヘッダをコピー

	resp, err := client.UploadImage(ctx, "/api/upload-image", largeImageData, "large.png")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "ファイルサイズが大きすぎます")
}

func TestUploadImage_UnsupportedFileType(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// GIF形式の画像データ（サポート外）
	gifData := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // GIF89a header
		0x01, 0x00, 0x01, 0x00, 0x00, 0x00, // width=1, height=1
		0x00, 0x2C, 0x00, 0x00, 0x00, 0x00, // image descriptor
		0x01, 0x00, 0x01, 0x00, 0x00, 0x02, // more descriptor
		0x02, 0x44, 0x01, 0x00, 0x3B,       // image data + trailer
	}

	resp, err := client.UploadImage(ctx, "/api/upload-image", gifData, "test.gif")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "サポートされていないファイル形式")
}
