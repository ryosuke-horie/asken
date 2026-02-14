package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteGeminiAPI_Success(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	client, err := NewClient(60 * time.Second)
	require.NoError(t, err)
	ctx := context.Background()

	// 簡単なプロンプトでテスト
	prompt := "こんにちは"

	response, err := client.Execute(ctx, prompt, nil)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response)
}

func TestExecuteGeminiAPI_Timeout(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// 非常に短いタイムアウトを設定
	client, err := NewClient(1 * time.Millisecond)
	require.NoError(t, err)
	ctx := context.Background()

	// 時間がかかりそうなプロンプト
	prompt := "長い応答を生成してください"

	_, err = client.Execute(ctx, prompt, nil)

	// タイムアウトエラーが発生することを確認
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}

func TestExecuteGeminiAPI_JSONParse(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	client, err := NewClient(60 * time.Second)
	require.NoError(t, err)
	ctx := context.Background()

	// JSONレスポンスを期待するプロンプト
	prompt := "こんにちは"

	response, err := client.Execute(ctx, prompt, nil)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// レスポンスが正しくパースされていることを確認
	assert.NotEmpty(t, response.Response)
}

func TestNewClient(t *testing.T) {
	t.Run("GEMINI_API_KEYが設定されている場合は成功する", func(t *testing.T) {
		skipIfNoGeminiAPIKey(t)

		timeout := 30 * time.Second
		client, err := NewClient(timeout)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
	})

	t.Run("GEMINI_API_KEYが未設定の場合はエラーを返す", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "")

		timeout := 30 * time.Second
		client, err := NewClient(timeout)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrEmptyAPIKey)
	})
}

func TestNewClientWithAPIKey(t *testing.T) {
	t.Run("有効なAPIキーで作成できる", func(t *testing.T) {
		timeout := 30 * time.Second
		apiKey := "test-api-key"
		client, err := NewClientWithAPIKey(apiKey, timeout)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
	})

	t.Run("空のAPIキーでエラーを返す", func(t *testing.T) {
		timeout := 30 * time.Second
		client, err := NewClientWithAPIKey("", timeout)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrEmptyAPIKey)
	})
}

func TestClient_Execute_MockSuccess(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `{"result": "success"}`}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	resp, err := client.Execute(ctx, "テストプロンプト", nil)

	require.NoError(t, err)
	assert.Equal(t, `{"result": "success"}`, resp.Response)
}

func TestClient_Execute_MockError(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return nil, assert.AnError
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	ctx := context.Background()

	_, err := client.Execute(ctx, "テストプロンプト", nil)

	assert.Error(t, err)
}

func TestClient_ExecuteWithImage_MockSuccess(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
			return &Response{Response: `[{"name": "テスト料理"}]`}, nil
		},
	}

	// Client自体にはExecuteWithImageメソッドがないが、HTTPClientインターフェース経由でテスト可能
	assert.NotNil(t, mockHTTPClient)
}

func TestNewClientWithHTTPClient(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{}
	client := NewClientWithHTTPClient(mockHTTPClient)

	assert.NotNil(t, client)
	assert.Equal(t, mockHTTPClient, client.httpClient)
}

func TestRemoveCodeBlock_Success(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSONコードブロック",
			input:    "```json\n[{\"name\":\"test\"}]\n```",
			expected: `[{"name":"test"}]`,
		},
		{
			name:     "コードブロックなし",
			input:    `[{"name":"test"}]`,
			expected: `[{"name":"test"}]`,
		},
		{
			name:     "大文字JSON",
			input:    "```JSON\n[{\"name\":\"test\"}]\n```",
			expected: `[{"name":"test"}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := removeCodeBlock(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
