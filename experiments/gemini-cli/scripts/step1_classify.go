package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GeminiResponse はGemini CLIのレスポンスを表す構造体
type GeminiResponse struct {
	SessionID string          `json:"session_id"`
	Response  string          `json:"response"`
	Stats     json.RawMessage `json:"stats"`
}

// FoodItem は食材情報を表す構造体（分類のみ）
type FoodItem struct {
	Name            string `json:"name"`
	EstimatedAmount string `json:"estimated_amount"`
}

// ClassifyImageConfig は画像分類の設定
type ClassifyImageConfig struct {
	ImagePath string
	Timeout   time.Duration
}

// ClassifyFoods は画像から食材を分類する（カロリー・栄養素情報は含まない）
func ClassifyFoods(ctx context.Context, config ClassifyImageConfig) ([]FoodItem, error) {
	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// 画像パスの存在確認
	if _, err := os.Stat(config.ImagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("画像ファイルが見つかりません: %s", config.ImagePath)
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(config.ImagePath)
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
			return nil, fmt.Errorf("タイムアウト: Gemini CLIの実行が%v以内に完了しませんでした", config.Timeout)
		}
		return nil, fmt.Errorf("Gemini CLI実行エラー: %w\n出力: %s", err, string(output))
	}

	// JSON部分を抽出（"Loaded cached credentials."などの余分な出力を除去）
	jsonStart := bytes.IndexByte(output, '{')
	if jsonStart == -1 {
		return nil, fmt.Errorf("JSON開始位置が見つかりません\n生データ: %s", string(output))
	}
	jsonData := output[jsonStart:]

	// JSONレスポンスをパース
	var response GeminiResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w\n生データ: %s", err, string(jsonData))
	}

	// レスポンス内のJSONコードブロックを抽出（Geminiが```json```で囲んでいる場合）
	foodListJSON := response.Response
	if strings.Contains(foodListJSON, "```") {
		// ```で分割して、コードブロック内のテキストを抽出
		parts := strings.Split(foodListJSON, "```")
		if len(parts) >= 3 {
			// parts[0]: コードブロック前のテキスト
			// parts[1]: "json\n[...]" (コードブロック内)
			// parts[2]: コードブロック後のテキスト
			content := parts[1]
			// "json" プレフィックスを除去
			content = strings.TrimPrefix(content, "json")
			content = strings.TrimPrefix(content, "JSON")
			foodListJSON = strings.TrimSpace(content)
		}
	}

	// 食材リストをパース
	var foods []FoodItem
	if err := json.Unmarshal([]byte(foodListJSON), &foods); err != nil {
		return nil, fmt.Errorf("食材リストのパースエラー: %w\nデータ: %s", err, foodListJSON)
	}

	return foods, nil
}

func main() {
	// 設定
	config := ClassifyImageConfig{
		ImagePath: "images/IMG_0374.JPG",
		Timeout:   60 * time.Second,
	}

	fmt.Println("📸 ステップ1: 食材分類")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("画像: %s\n", config.ImagePath)
	fmt.Printf("タイムアウト: %v\n\n", config.Timeout)

	// 食材分類を実行
	ctx := context.Background()
	foods, err := ClassifyFoods(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ エラー: %v\n", err)
		os.Exit(1)
	}

	// 結果を表示
	fmt.Println("✅ 分類完了")
	fmt.Printf("認識した食材数: %d個\n\n", len(foods))

	for i, food := range foods {
		fmt.Printf("%d. %s (%s)\n", i+1, food.Name, food.EstimatedAmount)
	}

	// 結果をファイルに保存
	resultPath := "results/step1_classify_result.json"
	resultData, err := json.MarshalIndent(foods, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  結果のJSON化エラー: %v\n", err)
		return
	}

	if err := os.WriteFile(resultPath, resultData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  ファイル保存エラー: %v\n", err)
		return
	}

	fmt.Printf("\n💾 結果を保存しました: %s\n", resultPath)
	fmt.Println("\n次のステップ: step2_nutrition.go を実行して栄養素を算出")
}
