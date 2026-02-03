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

	client := NewClient(60 * time.Second)
	ctx := context.Background()

	// 簡単なプロンプトでテスト
	prompt := "こんにちは"

	response, err := client.Execute(ctx, prompt)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response)
}

func TestExecuteGeminiAPI_Timeout(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// 非常に短いタイムアウトを設定
	client := NewClient(1 * time.Millisecond)
	ctx := context.Background()

	// 時間がかかりそうなプロンプト
	prompt := "長い応答を生成してください"

	_, err := client.Execute(ctx, prompt)

	// タイムアウトエラーが発生することを確認
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}

func TestExecuteGeminiAPI_JSONParse(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	client := NewClient(60 * time.Second)
	ctx := context.Background()

	// JSONレスポンスを期待するプロンプト
	prompt := "こんにちは"

	response, err := client.Execute(ctx, prompt)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// レスポンスが正しくパースされていることを確認
	assert.NotEmpty(t, response.Response)
}

func TestNewClient(t *testing.T) {
	timeout := 30 * time.Second
	client := NewClient(timeout)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
}

func TestNewClientWithAPIKey(t *testing.T) {
	timeout := 30 * time.Second
	apiKey := "test-api-key"
	client := NewClientWithAPIKey(apiKey, timeout)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
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
