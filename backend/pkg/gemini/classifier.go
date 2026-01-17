package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// FoodItem は食材情報を表す構造体（分類のみ）
type FoodItem struct {
	Name            string `json:"name"`
	EstimatedAmount string `json:"estimated_amount"`
}

// Classifier は食材分類を行うクライアント
type Classifier struct {
	timeout time.Duration
}

// NewClassifier は新しいClassifierを作成
func NewClassifier(timeout time.Duration) *Classifier {
	return &Classifier{
		timeout: timeout,
	}
}

// ClassifyFoods は画像から食材を分類する（カロリー・栄養素情報は含まない）
func (c *Classifier) ClassifyFoods(ctx context.Context, imagePath string) ([]FoodItem, error) {
	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 画像パスの存在確認
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("画像ファイルが見つかりません: %s", imagePath)
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("絶対パス変換エラー: %w", err)
	}

	// プロンプトを構築（食材分類のみに集中）
	prompt := fmt.Sprintf(`この画像に写っている食材や料理を特定し、各食材の名前と推定量（グラム数または個数）をJSON形式のリストで出力してください。

出力フォーマット:
[
  {
    "name": "食材名",
    "estimated_amount": "推定量（例: 100g, 3切れ, 1杯）"
  }
]

カロリーや栄養素の情報は不要です。食材の特定と量の推定のみを行ってください。

@%s`, absPath)

	// Gemini CLIコマンドを構築
	cmd := exec.CommandContext(ctx, "gemini", "-o", "json", prompt)

	// 標準出力と標準エラー出力をキャプチャ
	output, err := cmd.CombinedOutput()
	if err != nil {
		// コンテキストタイムアウトのチェック
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("タイムアウト: Gemini CLIの実行が%v以内に完了しませんでした", c.timeout)
		}
		return nil, fmt.Errorf("Gemini CLI実行エラー: %w\n出力: %s", err, string(output))
	}

	// JSON部分を抽出（"Loaded cached credentials."などの余分な出力を除去）
	jsonData := extractJSON(output)
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("JSON開始位置が見つかりません\n生データ: %s", string(output))
	}

	// JSONレスポンスをパース
	var response Response
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w\n生データ: %s", err, string(jsonData))
	}

	// レスポンス内のJSONコードブロックを抽出（Geminiが```json```で囲んでいる場合）
	foodListJSON := removeCodeBlock(response.Response)

	// 食材リストをパース
	var foods []FoodItem
	if err := json.Unmarshal([]byte(foodListJSON), &foods); err != nil {
		return nil, fmt.Errorf("食材リストのパースエラー: %w\nデータ: %s", err, foodListJSON)
	}

	return foods, nil
}
