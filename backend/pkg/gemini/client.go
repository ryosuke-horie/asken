package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Response はGemini CLIのレスポンスを表す構造体
type Response struct {
	SessionID string          `json:"session_id"`
	Response  string          `json:"response"`
	Stats     json.RawMessage `json:"stats"`
}

// Client はGemini CLIクライアント
type Client struct {
	timeout time.Duration
}

// NewClient は新しいGemini CLIクライアントを作成
func NewClient(timeout time.Duration) *Client {
	return &Client{
		timeout: timeout,
	}
}

// Execute はGemini CLIコマンドを実行し、レスポンスを返す
func (c *Client) Execute(ctx context.Context, prompt string) (*Response, error) {
	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Gemini CLIコマンドを構築
	cmd := exec.CommandContext(ctx, "gemini", "-m", "gemini-3-flash-preview", "-o", "json", prompt)

	// 標準出力と標準エラー出力をキャプチャ
	output, err := cmd.CombinedOutput()
	if err != nil {
		// コンテキストタイムアウトのチェック
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("タイムアウト: Gemini CLIの実行が%v以内に完了しませんでした", c.timeout)
		}
		return nil, fmt.Errorf("Gemini CLI実行エラー: %w\n出力: %s", err, string(output))
	}

	// JSON部分を抽出（"Loaded cached credentials."などの余分な出力を除去）
	jsonData := extractJSON(output)
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("JSON開始位置が見つかりません\n生データ: %s", string(output))
	}

	// JSONレスポンスをパース
	var response Response
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w\n生データ: %s", err, string(jsonData))
	}

	return &response, nil
}

// extractJSON はバイト列からJSON部分を抽出する
// "Loaded cached credentials."などのノイズを除去
func extractJSON(data []byte) []byte {
	jsonStart := bytes.IndexByte(data, '{')
	if jsonStart == -1 {
		return []byte{}
	}
	return data[jsonStart:]
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
