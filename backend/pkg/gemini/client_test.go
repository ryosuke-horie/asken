package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteGeminiCLI_Success(t *testing.T) {
	skipIfNoGeminiCLI(t)

	client := NewClient(60 * time.Second)
	ctx := context.Background()

	// 簡単なプロンプトでテスト
	prompt := "こんにちは"

	response, err := client.Execute(ctx, prompt)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.SessionID)
	assert.NotEmpty(t, response.Response)
}

func TestExecuteGeminiCLI_Timeout(t *testing.T) {
	skipIfNoGeminiCLI(t)

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

func TestExecuteGeminiCLI_JSONParse(t *testing.T) {
	skipIfNoGeminiCLI(t)

	client := NewClient(60 * time.Second)
	ctx := context.Background()

	// JSONレスポンスを期待するプロンプト
	prompt := "こんにちは"

	response, err := client.Execute(ctx, prompt)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// JSONフィールドが正しくパースされていることを確認
	assert.NotEmpty(t, response.SessionID)
	assert.NotEmpty(t, response.Response)
}

func TestExecuteGeminiCLI_InvalidCommand(t *testing.T) {
	// Gemini CLIが存在しない場合のテスト
	// このテストは実際の環境では実行されない可能性がある
	t.Skip("Gemini CLIのインストール状況に依存するためスキップ")
}

func TestExtractJSON_Success(t *testing.T) {
	// JSONノイズ除去のテスト
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "余分な出力あり",
			input:    `Loaded cached credentials.{"session_id":"123","response":"test","stats":{}}`,
			expected: `{"session_id":"123","response":"test","stats":{}}`,
		},
		{
			name:     "クリーンなJSON",
			input:    `{"session_id":"123","response":"test","stats":{}}`,
			expected: `{"session_id":"123","response":"test","stats":{}}`,
		},
		{
			name:     "複数行の余分な出力",
			input:    "Line1\nLine2\n{\"session_id\":\"123\",\"response\":\"test\",\"stats\":{}}",
			expected: `{"session_id":"123","response":"test","stats":{}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractJSON([]byte(tc.input))
			assert.Equal(t, tc.expected, string(result))
		})
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	input := "No JSON here"
	result := extractJSON([]byte(input))
	assert.Empty(t, result)
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
