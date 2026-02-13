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
	"time"
)

const defaultTestUID = "e2e-test-user"

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
