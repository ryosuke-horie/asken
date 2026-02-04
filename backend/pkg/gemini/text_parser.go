package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TextParser はテキストから食材を解析するクライアント
type TextParser struct {
	client *Client
}

// NewTextParser は新しいTextParserを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewTextParser(timeout time.Duration) *TextParser {
	return &TextParser{
		client: NewClient(timeout),
	}
}

// NewTextParserWithAPIKey はAPIキーを指定してTextParserを作成
func NewTextParserWithAPIKey(apiKey string, timeout time.Duration) *TextParser {
	return &TextParser{
		client: NewClientWithAPIKey(apiKey, timeout),
	}
}

// ParseTextToFoods はテキストから食材リストを生成する
func (tp *TextParser) ParseTextToFoods(ctx context.Context, inputText string) ([]FoodItem, error) {
	if inputText == "" {
		return nil, fmt.Errorf("入力テキストが空です")
	}

	prompt := fmt.Sprintf(`以下のテキストから料理やメニューを特定し、各料理の名前と推定量をJSON形式で出力してください。

入力テキスト: %s

出力フォーマット:
[
  {
    "name": "料理名",
    "estimated_amount": "推定量（例: 1人前, 1杯）"
  }
]

重要なルール:
- 料理名（例: ラーメン、チキンカツ定食、親子丼、幕の内弁当）はそのまま1つの料理として扱ってください
- 食材に分解しないでください（例: ラーメン → 麺、スープ、チャーシューに分解しない）
- 量が明記されていない場合は一般的な1食分の量を推定してください
- 「大盛り」「おかわり」などの表現は量に反映してください
- 個数（2個、3杯など）は適切な量に変換してください
- 日本語の料理名を使用してください`, inputText)

	response, err := tp.client.Execute(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIコールエラー: %w", err)
	}

	foodListJSON := removeCodeBlock(response.Response)

	var foods []FoodItem
	if err := json.Unmarshal([]byte(foodListJSON), &foods); err != nil {
		return nil, fmt.Errorf("食材リストのパースエラー: %w\nデータ: %s", err, foodListJSON)
	}

	return foods, nil
}
