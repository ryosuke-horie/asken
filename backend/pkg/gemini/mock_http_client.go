package gemini

import "context"

// MockGeminiHTTPClient はGeminiHTTPClientのモック実装
type MockGeminiHTTPClient struct {
	// ExecuteFunc はExecuteメソッドのモック関数
	ExecuteFunc func(ctx context.Context, prompt string) (*Response, error)

	// ExecuteWithImageFunc はExecuteWithImageメソッドのモック関数
	ExecuteWithImageFunc func(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error)
}

// Execute はモック関数を呼び出す
func (m *MockGeminiHTTPClient) Execute(ctx context.Context, prompt string) (*Response, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, prompt)
	}
	// デフォルトの成功レスポンス
	return &Response{Response: `[]`}, nil
}

// ExecuteWithImage はモック関数を呼び出す
func (m *MockGeminiHTTPClient) ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error) {
	if m.ExecuteWithImageFunc != nil {
		return m.ExecuteWithImageFunc(ctx, prompt, imageData, mimeType)
	}
	// デフォルトの成功レスポンス
	return &Response{Response: `[]`}, nil
}
