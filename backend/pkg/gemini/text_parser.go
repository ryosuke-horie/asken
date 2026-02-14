package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// TextParser はテキストから食材を解析するクライアント
type TextParser struct {
	client *Client
}

// NewTextParser は新しいTextParserを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewTextParser(timeout time.Duration) (*TextParser, error) {
	client, err := NewClient(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create text parser: %w", err)
	}
	return &TextParser{
		client: client,
	}, nil
}

// NewTextParserWithAPIKey はAPIキーを指定してTextParserを作成
func NewTextParserWithAPIKey(apiKey string, timeout time.Duration) (*TextParser, error) {
	client, err := NewClientWithAPIKey(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create text parser: %w", err)
	}
	return &TextParser{
		client: client,
	}, nil
}

// NewTextParserWithClient はClientインターフェースを受け取るコンストラクタ（テスト用）
func NewTextParserWithClient(client *Client) *TextParser {
	return &TextParser{
		client: client,
	}
}

// ParseTextToFoods はテキストから食材リストを生成する
func (tp *TextParser) ParseTextToFoods(ctx context.Context, inputText string) ([]FoodItem, error) {
	if inputText == "" {
		return nil, fmt.Errorf("入力テキストが空です")
	}

	log.Printf("TextParser: テキスト解析を開始 (入力長: %d文字)", len(inputText))

	prompt := fmt.Sprintf(`以下のテキストから料理やメニューを特定し、各料理の名前と推定量をJSON形式で出力してください。

入力テキスト: %s

推定量のルール:
- quantity_valueは数値で指定してください（例: 1, 2, 150, 200）
- quantity_unitは以下のいずれかを使用してください:
  g, ml, 杯, 人前, 個, 枚, 本, 切れ, 食, 皿, 膳, 丁, 束, 袋, 缶, 合, 玉, 粒
- 重量がわかる食材はgを使用してください（例: ご飯 → 200g）
- 飲み物やスープはmlを使用してください（例: 味噌汁 → 200ml）
- 料理は適切な助数詞を選択してください（例: ラーメン → 1杯, カレー → 1皿）

重要なルール:
- 料理名（例: ラーメン、チキンカツ定食、親子丼、幕の内弁当）はそのまま1つの料理として扱ってください
- 食材に分解しないでください（例: ラーメン → 麺、スープ、チャーシューに分解しない）
- 量が明記されていない場合は一般的な1食分の量を推定してください
- 「大盛り」「おかわり」などの表現はquantity_valueに反映してください
- 個数（2個、3杯など）はquantity_valueとquantity_unitに適切に変換してください
- 日本語の料理名を使用してください`, inputText)

	// responseSchemaで出力形式を強制
	schema := FoodItemSchema()

	response, err := tp.client.Execute(ctx, prompt, schema)
	if err != nil {
		log.Printf("TextParser: Gemini API呼び出しエラー: %v", err)
		return nil, fmt.Errorf("Gemini APIコールエラー: %w", err)
	}

	foodListJSON := removeCodeBlock(response.Response)

	// スキーマ制約付きレスポンスをパース
	var items []classifierResponseItem
	if err := json.Unmarshal([]byte(foodListJSON), &items); err != nil {
		log.Printf("TextParser: 食材リストのJSONパースエラー: %v", err)
		return nil, fmt.Errorf("食材リストのパースエラー: %w\nデータ: %s", err, foodListJSON)
	}

	// FoodItemに変換（quantity_value + quantity_unit → estimated_amount）
	foods := make([]FoodItem, len(items))
	for i, item := range items {
		foods[i] = item.toFoodItem()
	}

	log.Printf("TextParser: テキスト解析完了 (%d品を検出)", len(foods))
	return foods, nil
}
