package gemini

import (
	"context"
	"encoding/json"
	"fmt"
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
	httpClient *HTTPClient
}

// NewClassifier は新しいClassifierを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewClassifier(timeout time.Duration) *Classifier {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return &Classifier{
		httpClient: NewHTTPClient(apiKey, timeout),
	}
}

// NewClassifierWithAPIKey はAPIキーを指定してClassifierを作成
func NewClassifierWithAPIKey(apiKey string, timeout time.Duration) *Classifier {
	return &Classifier{
		httpClient: NewHTTPClient(apiKey, timeout),
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
	mimeType := detectMimeType(imagePath, imageData)

	// プロンプトを構築（料理名の分類に集中）
	prompt := `この画像に写っている料理を特定し、各料理の名前と推定量をJSON形式のリストで出力してください。

料理名は可能な限り具体的に出力してください。
例:
- ラーメン → 「家系ラーメン」「味噌ラーメン」「博多豚骨ラーメン」など
- カレー → 「カレーライス」「キーマカレー」「バターチキンカレー」など
- 丼物 → 「牛丼」「親子丼」「海鮮丼」など

出力フォーマット:
[
  {
    "name": "料理名",
    "estimated_amount": "推定量（例: 1人前, 1杯, 1皿）"
  }
]

カロリーや栄養素の情報は不要です。料理の特定と量の推定のみを行ってください。`

	// Gemini APIを呼び出す（画像付き）
	response, err := c.httpClient.ExecuteWithImage(ctx, prompt, imageData, mimeType)
	if err != nil {
		return nil, fmt.Errorf("Gemini API呼び出しエラー: %w", err)
	}

	// レスポンス内のJSONコードブロックを抽出（Geminiが```json```で囲んでいる場合）
	foodListJSON := removeCodeBlock(response.Response)

	// 料理リストをパース
	var foods []FoodItem
	if err := json.Unmarshal([]byte(foodListJSON), &foods); err != nil {
		return nil, fmt.Errorf("料理リストのパースエラー: %w\nデータ: %s", err, foodListJSON)
	}

	return foods, nil
}

// detectMimeType は画像ファイルのMIMEタイプを判定する
func detectMimeType(filePath string, data []byte) string {
	// マジックバイトで判定
	if len(data) >= 4 {
		// JPEG
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
		// PNG
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return "image/png"
		}
		// GIF
		if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
			return "image/gif"
		}
		// WebP
		if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
			if data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
				return "image/webp"
			}
		}
	}

	// 拡張子で判定（フォールバック）
	switch {
	case hasExtension(filePath, ".jpg", ".jpeg"):
		return "image/jpeg"
	case hasExtension(filePath, ".png"):
		return "image/png"
	case hasExtension(filePath, ".gif"):
		return "image/gif"
	case hasExtension(filePath, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg" // デフォルト
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
