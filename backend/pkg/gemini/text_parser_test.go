package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTextToFoods_EmptyInput(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{}
	client := NewClientWithHTTPClient(mockHTTPClient)
	parser := NewTextParserWithClient(client)
	ctx := context.Background()

	_, err := parser.ParseTextToFoods(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "入力テキストが空です")
}

func TestParseTextToFoods_Success(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	parser, err := NewTextParser(60 * time.Second)
	require.NoError(t, err)
	ctx := context.Background()

	// 一般的な食事テキストでテスト
	foods, err := parser.ParseTextToFoods(ctx, "ご飯二杯と焼肉")

	require.NoError(t, err)
	assert.NotEmpty(t, foods)

	// 各食材が適切な構造を持っているか確認
	for _, food := range foods {
		assert.NotEmpty(t, food.Name, "食材名が空です")
		assert.NotEmpty(t, food.EstimatedAmount, "推定量が空です")
	}
}

func TestParseTextToFoods_ComplexInput(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	parser, err := NewTextParser(60 * time.Second)
	require.NoError(t, err)
	ctx := context.Background()

	// 複雑な入力テキストでテスト
	foods, err := parser.ParseTextToFoods(ctx, "チキンカツ定食、味噌汁付き")

	require.NoError(t, err)
	assert.NotEmpty(t, foods)

	// 定食は複数の食材に分解されるはず
	assert.GreaterOrEqual(t, len(foods), 1)
}

func TestParseTextToFoods_Timeout(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	// 非常に短いタイムアウトでテスト
	parser, err := NewTextParser(1 * time.Millisecond)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = parser.ParseTextToFoods(ctx, "ご飯")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}

func TestTextParser_ParseTextToFoods_MockSuccess(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{
				Response: `[{"name": "ラーメン", "quantity_value": 1, "quantity_unit": "杯"}]`,
			}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	parser := NewTextParserWithClient(client)
	ctx := context.Background()

	foods, err := parser.ParseTextToFoods(ctx, "ラーメンを食べた")

	require.NoError(t, err)
	assert.Len(t, foods, 1)
	assert.Equal(t, "ラーメン", foods[0].Name)
	assert.Equal(t, "1杯", foods[0].EstimatedAmount)
}

func TestTextParser_ParseTextToFoods_MockAPIError(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return nil, assert.AnError
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	parser := NewTextParserWithClient(client)
	ctx := context.Background()

	_, err := parser.ParseTextToFoods(ctx, "テスト")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini APIコールエラー")
}

func TestTextParser_ParseTextToFoods_MockInvalidJSON(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `{invalid json}`}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	parser := NewTextParserWithClient(client)
	ctx := context.Background()

	_, err := parser.ParseTextToFoods(ctx, "テスト")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "パースエラー")
}

func TestTextParser_ParseTextToFoods_MockCodeBlock(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: "```json\n[{\"name\": \"カレーライス\", \"quantity_value\": 1, \"quantity_unit\": \"皿\"}]\n```"}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	parser := NewTextParserWithClient(client)
	ctx := context.Background()

	foods, err := parser.ParseTextToFoods(ctx, "カレーライス")

	require.NoError(t, err)
	assert.Len(t, foods, 1)
	assert.Equal(t, "カレーライス", foods[0].Name)
	assert.Equal(t, "1皿", foods[0].EstimatedAmount)
}

func TestNewTextParserWithClient(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{}
	client := NewClientWithHTTPClient(mockClient)
	parser := NewTextParserWithClient(client)

	assert.NotNil(t, parser)
	assert.Equal(t, client, parser.client)
}

func TestNewTextParser(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	timeout := 30 * time.Second
	parser, err := NewTextParser(timeout)

	require.NoError(t, err)
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.client)
}

func TestNewTextParser_EmptyAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	parser, err := NewTextParser(60 * time.Second)

	require.Error(t, err)
	assert.Nil(t, parser)
	assert.ErrorIs(t, err, ErrEmptyAPIKey)
}

func TestNewTextParserWithAPIKey_EmptyAPIKey(t *testing.T) {
	parser, err := NewTextParserWithAPIKey("", 60*time.Second)

	require.Error(t, err)
	assert.Nil(t, parser)
	assert.ErrorIs(t, err, ErrEmptyAPIKey)
}
