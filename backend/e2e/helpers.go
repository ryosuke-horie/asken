//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"
)

const defaultTestUID = "e2e-test-user"

// minimalJPEG は最小限の有効なJPEGバイト列（テスト用）
var minimalJPEG = []byte{
	0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
	0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
	0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
	0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
	0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
	0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
	0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
	0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0A, 0x0B, 0xFF, 0xC4, 0x00, 0xB5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
	0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
	0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xA1, 0x08,
	0x23, 0x42, 0xB1, 0xC1, 0x15, 0x52, 0xD1, 0xF0, 0x24, 0x33, 0x62, 0x72,
	0x82, 0x09, 0x0A, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x25, 0x26, 0x27, 0x28,
	0x29, 0x2A, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x43, 0x44, 0x45,
	0x46, 0x47, 0x48, 0x49, 0x4A, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
	0x5A, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x73, 0x74, 0x75,
	0x76, 0x77, 0x78, 0x79, 0x7A, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
	0x8A, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0xA2, 0xA3,
	0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6,
	0xB7, 0xB8, 0xB9, 0xBA, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9,
	0xCA, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xE1, 0xE2,
	0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xF1, 0xF2, 0xF3, 0xF4,
	0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01,
	0x00, 0x00, 0x3F, 0x00, 0xFB, 0xD2, 0x8A, 0x28, 0x03, 0xFF, 0xD9,
}

// waitForUserRateLimit はユーザーレート制限のリセットを待つ
//
// レート制限設定: UserRateLimit=2 (2 RPS), UserBurstSize=5
// 全認証済みエンドポイントで共有されるため、前のテストでバケットが
// 枯渇する場合がある。3秒待機でバケットを回復させる。
func waitForUserRateLimit() {
	time.Sleep(3 * time.Second)
}

// waitForGeminiRateLimit はGemini APIのレート制限リセットを待つ
//
// レート制限設定: GeminiRateLimit=0.2 (5秒に1回), GeminiBurstSize=2
// バーストを使い切った後は10秒待つ（余裕を持たせるため5秒→10秒に増量）
func waitForGeminiRateLimit() {
	time.Sleep(10 * time.Second)
}

// skipIfGeminiDisabled はGemini APIを使用するテストをスキップする
//
// デフォルトではGemini APIを呼び出すテストをスキップし、APIコストとレート制限を回避する。
// E2E_RUN_GEMINI=true を設定することでGeminiテストを有効化できる。
func skipIfGeminiDisabled(t *testing.T) {
	t.Helper()
	if os.Getenv("E2E_RUN_GEMINI") != "true" {
		t.Skip("Skipping Gemini API test (set E2E_RUN_GEMINI=true to enable)")
	}
}

// testUID はE2Eテスト用のユーザーIDを返す
func testUID() string {
	if uid := os.Getenv("E2E_TEST_UID"); uid != "" {
		return uid
	}
	return defaultTestUID
}

// Client はE2Eテスト用のHTTPクライアント
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewClient は新しいE2Eテストクライアントを作成する
func NewClient() (*Client, error) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("E2E_BASE_URL environment variable is required")
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// WithAuthToken は認証トークンを設定した新しいクライアントを返す
// テスト間で状態を共有しないために、元のクライアントは変更しない
func (c *Client) WithAuthToken(token string) *Client {
	return &Client{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		authToken:  token,
	}
}

// Request はHTTPリクエストを実行する
func (c *Client) Request(ctx context.Context, method, path string, body any) (*Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// Get はGETリクエストを実行する
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

// Post はPOSTリクエストを実行する
func (c *Client) Post(ctx context.Context, path string, body any) (*Response, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

// Response はHTTPレスポンスを表す
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// JSON はレスポンスボディをJSONとしてデコードする
func (r *Response) JSON(v any) error {
	if err := json.Unmarshal(r.Body, v); err != nil {
		bodyPreview := string(r.Body)
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..."
		}
		return fmt.Errorf("failed to decode response (status=%d, body=%s): %w", r.StatusCode, bodyPreview, err)
	}
	return nil
}

// UploadImage は画像ファイルをアップロードする（multipart/form-data）
func (c *Client) UploadImage(ctx context.Context, path string, imageData []byte, filename string) (*Response, error) {
	body := &bytes.Buffer{}

	// mime/multipart.Writerを使用して安全にmultipart/form-dataを構築
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile("image", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write image data: %w", err)
	}

	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// createTestIngredient はテスト用の食材を作成するヘルパー
// 作成した食材のIDを返す
func createTestIngredient(t *testing.T, client *Client, ctx context.Context, name, category string, quantity float64, unit string) string {
	t.Helper()

	reqBody := map[string]any{
		"name":     name,
		"category": category,
		"quantity": quantity,
		"unit":     unit,
		"source":   "manual",
	}
	resp, err := client.Post(ctx, "/api/ingredients", reqBody)
	if err != nil {
		t.Fatalf("createTestIngredient: failed to create ingredient: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createTestIngredient: expected 201, got %d (body: %s)", resp.StatusCode, string(resp.Body))
	}

	var body map[string]any
	if err := resp.JSON(&body); err != nil {
		t.Fatalf("createTestIngredient: failed to parse response: %v", err)
	}

	id, ok := body["id"].(string)
	if !ok || id == "" {
		t.Fatalf("createTestIngredient: ingredient id is empty or not a string")
	}
	return id
}
