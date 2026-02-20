package gemini

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceiptParser_ParseReceiptImage_Success(t *testing.T) {
	mockResponse := `[
		{"name":"鶏むね肉","category":"meat","quantity_value":300,"quantity_unit":"g"},
		{"name":"牛乳","category":"dairy","quantity_value":1,"quantity_unit":"パック"}
	]`

	mockClient := &MockGeminiHTTPClient{
		ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
			assert.Equal(t, "image/jpeg", mimeType)
			assert.NotEmpty(t, prompt)
			assert.NotNil(t, schema)
			return &Response{Response: mockResponse}, nil
		},
	}

	parser := NewReceiptParserWithHTTPClient(mockClient)
	ingredients, err := parser.ParseReceiptImage(context.Background(), []byte("test-image"), "image/jpeg")

	require.NoError(t, err)
	require.Len(t, ingredients, 2)

	assert.Equal(t, "鶏むね肉", ingredients[0].Name)
	assert.Equal(t, "meat", ingredients[0].Category)
	assert.Equal(t, float64(300), ingredients[0].Quantity)
	assert.Equal(t, "g", ingredients[0].Unit)

	assert.Equal(t, "牛乳", ingredients[1].Name)
	assert.Equal(t, "dairy", ingredients[1].Category)
	assert.Equal(t, float64(1), ingredients[1].Quantity)
	assert.Equal(t, "パック", ingredients[1].Unit)
}

func TestReceiptParser_ParseReceiptImage_EmptyResult(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
			return &Response{Response: `[]`}, nil
		},
	}

	parser := NewReceiptParserWithHTTPClient(mockClient)
	ingredients, err := parser.ParseReceiptImage(context.Background(), []byte("test-image"), "image/jpeg")

	require.NoError(t, err)
	assert.Empty(t, ingredients)
}

func TestReceiptParser_ParseReceiptImage_APIError(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	parser := NewReceiptParserWithHTTPClient(mockClient)
	_, err := parser.ParseReceiptImage(context.Background(), []byte("test-image"), "image/jpeg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini API呼び出しエラー")
}

func TestReceiptParser_ParseReceiptImage_InvalidJSON(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{
		ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
			return &Response{Response: `invalid json`}, nil
		},
	}

	parser := NewReceiptParserWithHTTPClient(mockClient)
	_, err := parser.ParseReceiptImage(context.Background(), []byte("test-image"), "image/jpeg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "食材リストのパースエラー")
}

func TestReceiptParser_ParseReceiptImage_CodeBlockResponse(t *testing.T) {
	// Geminiが```json```で囲んだレスポンスを返す場合
	mockResponse := "```json\n[{\"name\":\"トマト\",\"category\":\"vegetable\",\"quantity_value\":3,\"quantity_unit\":\"個\"}]\n```"

	mockClient := &MockGeminiHTTPClient{
		ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
			return &Response{Response: mockResponse}, nil
		},
	}

	parser := NewReceiptParserWithHTTPClient(mockClient)
	ingredients, err := parser.ParseReceiptImage(context.Background(), []byte("test-image"), "image/png")

	require.NoError(t, err)
	require.Len(t, ingredients, 1)
	assert.Equal(t, "トマト", ingredients[0].Name)
	assert.Equal(t, "vegetable", ingredients[0].Category)
	assert.Equal(t, float64(3), ingredients[0].Quantity)
	assert.Equal(t, "個", ingredients[0].Unit)
}

func TestReceiptIngredientSchema(t *testing.T) {
	schema := ReceiptIngredientSchema()

	assert.Equal(t, SchemaTypeArray, schema.Type)
	assert.NotNil(t, schema.Items)
	assert.Equal(t, SchemaTypeObject, schema.Items.Type)

	props := schema.Items.Properties
	assert.Contains(t, props, "name")
	assert.Contains(t, props, "category")
	assert.Contains(t, props, "quantity_value")
	assert.Contains(t, props, "quantity_unit")

	// カテゴリのEnum確認
	assert.Contains(t, props["category"].Enum, "meat")
	assert.Contains(t, props["category"].Enum, "vegetable")
	assert.Contains(t, props["category"].Enum, "other")

	// 必須フィールドの確認
	assert.Contains(t, schema.Items.Required, "name")
	assert.Contains(t, schema.Items.Required, "category")
	assert.Contains(t, schema.Items.Required, "quantity_value")
	assert.Contains(t, schema.Items.Required, "quantity_unit")
}

func TestReceiptParserResponseItem_toReceiptIngredient(t *testing.T) {
	item := receiptParserResponseItem{
		Name:          "鶏むね肉",
		Category:      "meat",
		QuantityValue: 500,
		QuantityUnit:  "g",
	}

	result := item.toReceiptIngredient()

	assert.Equal(t, "鶏むね肉", result.Name)
	assert.Equal(t, "meat", result.Category)
	assert.Equal(t, float64(500), result.Quantity)
	assert.Equal(t, "g", result.Unit)
}
