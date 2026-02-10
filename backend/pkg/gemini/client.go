package gemini

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// Response はGemini APIのレスポンスを表す構造体
// SessionIDとStatsはCLI時代の名残で、HTTP APIでは使用されないが後方互換性のため保持
type Response struct {
	SessionID string          `json:"session_id"`
	Response  string          `json:"response"`
	Stats     json.RawMessage `json:"stats"`
}

// Client はGemini APIクライアント
type Client struct {
	httpClient GeminiHTTPClient
}

// NewClient は新しいGemini APIクライアントを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewClient(timeout time.Duration) *Client {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return &Client{
		httpClient: NewHTTPClient(apiKey, timeout),
	}
}

// NewClientWithAPIKey はAPIキーを指定してGemini APIクライアントを作成
func NewClientWithAPIKey(apiKey string, timeout time.Duration) *Client {
	return &Client{
		httpClient: NewHTTPClient(apiKey, timeout),
	}
}

// NewClientWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewClientWithHTTPClient(httpClient GeminiHTTPClient) *Client {
	return &Client{
		httpClient: httpClient,
	}
}

// Execute はGemini APIを呼び出し、レスポンスを返す
func (c *Client) Execute(ctx context.Context, prompt string) (*Response, error) {
	return c.httpClient.Execute(ctx, prompt)
}

// removeCodeBlock はMarkdownコードブロック（```json```）を除去する
func removeCodeBlock(text string) string {
	if !strings.Contains(text, "```") {
		return text
	}

	// ```で分割して、コードブロック内のテキストを抽出
	parts := strings.Split(text, "```")
	if len(parts) >= 3 {
		// parts[0]: コードブロック前のテキスト
		// parts[1]: "json\n[...]" (コードブロック内)
		// parts[2]: コードブロック後のテキスト
		content := parts[1]
		// "json" または "JSON" プレフィックスを除去
		content = strings.TrimPrefix(content, "json")
		content = strings.TrimPrefix(content, "JSON")
		return strings.TrimSpace(content)
	}

	return text
}
