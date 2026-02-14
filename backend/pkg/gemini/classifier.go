package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// FoodItem は料理情報を表す構造体（分類のみ）
type FoodItem struct {
	Name            string `json:"name"`
	EstimatedAmount string `json:"estimated_amount"`
}

// Classifier は料理分類を行うクライアント
type Classifier struct {
	httpClient GeminiHTTPClient
}

// NewClassifier は新しいClassifierを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewClassifier(timeout time.Duration) (*Classifier, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	httpClient, err := NewHTTPClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier: %w", err)
	}
	return &Classifier{
		httpClient: httpClient,
	}, nil
}

// NewClassifierWithAPIKey はAPIキーを指定してClassifierを作成
func NewClassifierWithAPIKey(apiKey string, timeout time.Duration) (*Classifier, error) {
	httpClient, err := NewHTTPClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier: %w", err)
	}
	return &Classifier{
		httpClient: httpClient,
	}, nil
}

// NewClassifierWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewClassifierWithHTTPClient(httpClient GeminiHTTPClient) *Classifier {
	return &Classifier{
		httpClient: httpClient,
	}
}

// ClassifyFoods は画像から料理を分類する（カロリー・栄養素情報は含まない）
func (c *Classifier) ClassifyFoods(ctx context.Context, imagePath string) ([]FoodItem, error) {
	// 画像パスの存在確認
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("画像ファイルが見つかりません: %s", imagePath)
	}

	// 画像ファイルを読み込む
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("画像ファイル読み込みエラー: %w", err)
	}

	// MIMEタイプを判定
	mimeType, err := detectMimeType(imagePath, imageData)
	if err != nil {
		return nil, err
	}

	return c.ClassifyFoodsFromData(ctx, imageData, mimeType)
}

// ClassifyFoodsFromData はバイトデータから料理を分類する（Cloud Storage等からダウンロードした画像用）
func (c *Classifier) ClassifyFoodsFromData(ctx context.Context, imageData []byte, mimeType string) ([]FoodItem, error) {
	log.Printf("Classifier: 料理分類を開始 (画像サイズ: %d bytes, MIME: %s)", len(imageData), mimeType)

	// プロンプトを構築（料理名の分類に集中）
	prompt := `この画像に写っている料理を特定し、各料理の名前と推定量をJSON形式のリストで出力してください。

料理名は可能な限り具体的に出力してください。
例:
- ラーメン → 「家系ラーメン」「味噌ラーメン」「博多豚骨ラーメン」など
- カレー → 「カレーライス」「キーマカレー」「バターチキンカレー」など
- 丼物 → 「牛丼」「親子丼」「海鮮丼」など

推定量のルール:
- quantity_valueは数値で指定してください（例: 1, 2, 150, 200）
- quantity_unitは以下のいずれかを使用してください:
  g, ml, 杯, 人前, 個, 枚, 本, 切れ, 食, 皿, 膳, 丁, 束, 袋, 缶, 合, 玉, 粒
- 重量がわかる食材はgを使用してください（例: ご飯 → 200g）
- 飲み物やスープはmlを使用してください（例: 味噌汁 → 200ml）
- 料理は適切な助数詞を選択してください（例: ラーメン → 1杯, カレー → 1皿）

カロリーや栄養素の情報は不要です。料理の特定と量の推定のみを行ってください。`

	// responseSchemaで出力形式を強制
	schema := FoodItemSchema()

	// Gemini APIを呼び出す（画像付き）
	response, err := c.httpClient.ExecuteWithImage(ctx, prompt, imageData, mimeType, schema)
	if err != nil {
		log.Printf("Classifier: Gemini API呼び出しエラー: %v", err)
		return nil, fmt.Errorf("Gemini API呼び出しエラー: %w", err)
	}

	// レスポンス内のJSONコードブロックを抽出（Geminiが```json```で囲んでいる場合）
	foodListJSON := removeCodeBlock(response.Response)

	// スキーマ制約付きレスポンスをパース
	var items []classifierResponseItem
	if err := json.Unmarshal([]byte(foodListJSON), &items); err != nil {
		log.Printf("Classifier: 料理リストのJSONパースエラー: %v", err)
		return nil, fmt.Errorf("料理リストのパースエラー: %w\nデータ: %s", err, foodListJSON)
	}

	// FoodItemに変換（quantity_value + quantity_unit → estimated_amount）
	foods := make([]FoodItem, len(items))
	for i, item := range items {
		foods[i] = item.toFoodItem()
	}

	log.Printf("Classifier: 料理分類完了 (%d品を検出)", len(foods))
	return foods, nil
}

// detectMimeType は画像ファイルのMIMEタイプを判定する
// サポートする形式: JPEG, PNG, GIF, WebP
// サポート外の形式の場合はエラーを返す
func detectMimeType(filePath string, data []byte) (string, error) {
	// マジックバイトで判定
	if len(data) >= 4 {
		// JPEG
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg", nil
		}
		// PNG
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return "image/png", nil
		}
		// GIF
		if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
			return "image/gif", nil
		}
		// WebP
		if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
			if data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
				return "image/webp", nil
			}
		}
	}

	// 拡張子で判定（フォールバック）
	switch {
	case hasExtension(filePath, ".jpg", ".jpeg"):
		return "image/jpeg", nil
	case hasExtension(filePath, ".png"):
		return "image/png", nil
	case hasExtension(filePath, ".gif"):
		return "image/gif", nil
	case hasExtension(filePath, ".webp"):
		return "image/webp", nil
	default:
		return "", fmt.Errorf("サポートされていない画像形式です: %s (JPEG, PNG, GIF, WebPのみ対応)", filePath)
	}
}

// hasExtension はファイルパスが指定の拡張子を持つかチェックする
func hasExtension(filePath string, extensions ...string) bool {
	lowerPath := strings.ToLower(filePath)
	for _, ext := range extensions {
		if strings.HasSuffix(lowerPath, ext) {
			return true
		}
	}
	return false
}
