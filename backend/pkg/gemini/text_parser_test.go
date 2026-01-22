package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTextToFoods_EmptyInput(t *testing.T) {
	parser := NewTextParser(30 * time.Second)
	ctx := context.Background()

	_, err := parser.ParseTextToFoods(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "入力テキストが空です")
}

func TestParseTextToFoods_Success(t *testing.T) {
	skipIfNoGeminiCLI(t)

	parser := NewTextParser(60 * time.Second)
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
	skipIfNoGeminiCLI(t)

	parser := NewTextParser(60 * time.Second)
	ctx := context.Background()

	// 複雑な入力テキストでテスト
	foods, err := parser.ParseTextToFoods(ctx, "チキンカツ定食、味噌汁付き")

	require.NoError(t, err)
	assert.NotEmpty(t, foods)

	// 定食は複数の食材に分解されるはず
	assert.GreaterOrEqual(t, len(foods), 1)
}

func TestParseTextToFoods_Timeout(t *testing.T) {
	skipIfNoGeminiCLI(t)

	// 非常に短いタイムアウトでテスト
	parser := NewTextParser(1 * time.Millisecond)
	ctx := context.Background()

	_, err := parser.ParseTextToFoods(ctx, "ご飯")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "タイムアウト")
}

func TestNewTextParser(t *testing.T) {
	timeout := 30 * time.Second
	parser := NewTextParser(timeout)

	assert.NotNil(t, parser)
	assert.NotNil(t, parser.client)
}
