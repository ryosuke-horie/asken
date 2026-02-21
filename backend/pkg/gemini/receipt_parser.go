package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// ReceiptIngredient はレシートから抽出した食材
type ReceiptIngredient struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

// ReceiptParser はレシート画像から食材を解析するクライアント
type ReceiptParser struct {
	httpClient GeminiHTTPClient
}

// NewReceiptParser は新しいReceiptParserを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewReceiptParser(timeout time.Duration) (*ReceiptParser, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("failed to create receipt parser: 環境変数 GEMINI_API_KEY が設定されていません")
	}
	httpClient, err := NewHTTPClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create receipt parser: %w", err)
	}
	return &ReceiptParser{httpClient: httpClient}, nil
}

// NewReceiptParserWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewReceiptParserWithHTTPClient(httpClient GeminiHTTPClient) *ReceiptParser {
	if httpClient == nil {
		panic("receipt parser: httpClient must not be nil")
	}
	return &ReceiptParser{httpClient: httpClient}
}

// ParseReceiptImage はレシート画像から食材リストを抽出する
// imageData: 画像バイナリデータ（JPEG or PNG）
// mimeType: 画像のMIMEタイプ（image/jpeg or image/png）
func (p *ReceiptParser) ParseReceiptImage(ctx context.Context, imageData []byte, mimeType string) ([]ReceiptIngredient, error) {
	log.Printf("ReceiptParser: レシート解析を開始 (画像サイズ: %d bytes, MIME: %s)", len(imageData), mimeType)

	prompt := fmt.Sprintf(`このレシート画像から食材リストを抽出してください。

抽出ルール:
- レシートに記載された商品名から食材名を推定する
- パッケージ商品（例: 「若鶏もも 2P」）は食材名と数量に分解する
- 食材以外の商品（日用品、洗剤、ポイント関連、袋代、消費税等）は除外する
- カテゴリは食材の種類から自動判別する（meat/fish/vegetable/fruit/dairy/grain/seasoning/beverage/other）
- 数量が不明な場合は一般的な1パックの量を推定する

数量のルール:
- quantity_valueは数値で指定する（例: 1, 500, 2）
- quantity_unitは以下のいずれかを使用する:
  %s
- 重量がわかる食材はgを使用する（例: 鶏むね肉 → 300g）
- 個数がわかる場合は「個」を使用する（例: 卵 → 6個）
- 不明な場合は「個」または「パック」を使用する`, SupportedUnitsCSV())

	schema := ReceiptIngredientSchema()

	response, err := p.httpClient.ExecuteWithImage(ctx, prompt, imageData, mimeType, schema)
	if err != nil {
		log.Printf("ReceiptParser: Gemini API呼び出しエラー: %v", err)
		return nil, fmt.Errorf("Gemini API呼び出しエラー: %w", err)
	}

	listJSON := removeCodeBlock(response.Response)

	var items []receiptParserResponseItem
	if err := json.Unmarshal([]byte(listJSON), &items); err != nil {
		log.Printf("ReceiptParser: 食材リストのJSONパースエラー: %v", err)
		return nil, fmt.Errorf("食材リストのパースエラー: %w", err)
	}

	ingredients := make([]ReceiptIngredient, 0, len(items))
	for _, item := range items {
		if item.Name == "" {
			log.Printf("ReceiptParser: 警告 - name が空の食材をスキップします")
			continue
		}
		if item.QuantityValue < 0 {
			log.Printf("ReceiptParser: 警告 - quantity_value が負値の食材をスキップします: name=%s, value=%v", item.Name, item.QuantityValue)
			continue
		}
		ingredients = append(ingredients, item.toReceiptIngredient())
	}

	if len(ingredients) == 0 {
		log.Printf("ReceiptParser: 警告 - レシートから食材を検出できませんでした (imageSize=%d bytes, mimeType=%s)", len(imageData), mimeType)
	} else {
		log.Printf("ReceiptParser: レシート解析完了 (%d品を検出, imageSize=%d bytes)", len(ingredients), len(imageData))
	}
	return ingredients, nil
}
