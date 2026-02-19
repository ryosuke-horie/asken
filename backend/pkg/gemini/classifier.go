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
	prompt := fmt.Sprintf(`この画像に写っている料理を特定し、各料理の名前と推定量をJSON形式のリストで出力してください。

料理名は可能な限り具体的に出力してください。
例:
- ラーメン → 「家系ラーメン」「味噌ラーメン」「博多豚骨ラーメン」など
- カレー → 「カレーライス」「キーマカレー」「バターチキンカレー」など
- 丼物 → 「牛丼」「親子丼」「海鮮丼」など

推定量のルール:
- quantity_valueは数値で指定してください（例: 1, 2, 150, 200）
- quantity_unitは以下のいずれかを使用してください:
  %s
- 重量がわかる食材はgを使用してください（例: ご飯 → 200g）
- 飲み物やスープはmlを使用してください（例: 味噌汁 → 200ml）
- 料理は適切な助数詞を選択してください（例: ラーメン → 1杯, カレー → 1皿）

カロリーや栄養素の情報は不要です。料理の特定と量の推定のみを行ってください。`, SupportedUnitsCSV())

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

// magicSignature はマジックバイトによるMIMEタイプ判定のエントリ
type magicSignature struct {
	mimeType string
	offset   int
	magic    []byte
}

// magicSignatures はマジックバイト判定テーブル
// WebPはRIFFヘッダー（offset=0）+WEBPマーカー（offset=8）の2段階で判定するため、
// RIFFヘッダー部分は拡張子フォールバックに委ねる形で、WEBPマーカーのみoffset=8で登録する。
var magicSignatures = []magicSignature{
	{"image/jpeg", 0, []byte{0xFF, 0xD8, 0xFF}},
	{"image/png", 0, []byte{0x89, 0x50, 0x4E, 0x47}},
	{"image/gif", 0, []byte{0x47, 0x49, 0x46}},
	{"image/webp", 8, []byte{0x57, 0x45, 0x42, 0x50}},
}

// extensionMimeMap は拡張子からMIMEタイプへのマッピング
var extensionMimeMap = []struct {
	mimeType   string
	extensions []string
}{
	{"image/jpeg", []string{".jpg", ".jpeg"}},
	{"image/png", []string{".png"}},
	{"image/gif", []string{".gif"}},
	{"image/webp", []string{".webp"}},
}

// detectMimeType は画像ファイルのMIMEタイプを判定する
// サポートする形式: JPEG, PNG, GIF, WebP
// サポート外の形式の場合はエラーを返す
func detectMimeType(filePath string, data []byte) (string, error) {
	// マジックバイトで判定
	for _, sig := range magicSignatures {
		end := sig.offset + len(sig.magic)
		if len(data) >= end && matchBytes(data[sig.offset:end], sig.magic) {
			return sig.mimeType, nil
		}
	}

	// 拡張子で判定（フォールバック）
	for _, entry := range extensionMimeMap {
		if hasExtension(filePath, entry.extensions...) {
			return entry.mimeType, nil
		}
	}

	return "", fmt.Errorf("サポートされていない画像形式です: %s (JPEG, PNG, GIF, WebPのみ対応)", filePath)
}

// matchBytes は2つのバイトスライスが一致するか比較する
func matchBytes(data, pattern []byte) bool {
	for i, b := range pattern {
		if data[i] != b {
			return false
		}
	}
	return true
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
