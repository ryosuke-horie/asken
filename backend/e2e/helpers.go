//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const defaultTestUID = "e2e-test-user"

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

	req.Header.Set("Content-Type", "application/json")
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
