// このファイルはテスト専用モックです。
// 複数パッケージのテストから参照されるため _test.go にできません。
// プロダクションコードから MockGeminiHTTPClient を使用しないでください。
package gemini

import "context"

// MockGeminiHTTPClient はGeminiHTTPClientのモック実装
type MockGeminiHTTPClient struct {
	// ExecuteFunc はExecuteメソッドのモック関数
	ExecuteFunc func(ctx context.Context, prompt string, schema *Schema) (*Response, error)

	// ExecuteWithImageFunc はExecuteWithImageメソッドのモック関数
	ExecuteWithImageFunc func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error)
}

// Execute はモック関数を呼び出す
func (m *MockGeminiHTTPClient) Execute(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, prompt, schema)
	}
	// デフォルトの成功レスポンス
	return &Response{Response: `[]`}, nil
}

// ExecuteWithImage はモック関数を呼び出す
func (m *MockGeminiHTTPClient) ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
	if m.ExecuteWithImageFunc != nil {
		return m.ExecuteWithImageFunc(ctx, prompt, imageData, mimeType, schema)
	}
	// デフォルトの成功レスポンス
	return &Response{Response: `[]`}, nil
}
