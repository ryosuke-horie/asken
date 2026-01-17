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

// FoodItem は食品情報を表す構造体
type FoodItem struct {
	Name             string  `json:"name"`
	EstimatedAmount  string  `json:"estimated_amount,omitempty"`
	Calories         float64 `json:"calories_kcal"`
	Protein          float64 `json:"protein_g"`
	Fat              float64 `json:"fat_g"`
	Carbohydrates    float64 `json:"carbohydrates_g,omitempty"`
	Carbs            float64 `json:"carbs_g,omitempty"`
	Note             string  `json:"note,omitempty"`
	Description      string  `json:"description,omitempty"`
}

// AnalyzeImageConfig は画像分析の設定
type AnalyzeImageConfig struct {
	ImagePath string
	Timeout   time.Duration
}

// AnalyzeImage はGemini CLIを使って画像から食材を分析する
func AnalyzeImage(ctx context.Context, config AnalyzeImageConfig) (*GeminiResponse, error) {
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

	// プロンプトを構築
	prompt := fmt.Sprintf("この画像に写っている食材や料理を特定し、それぞれのカロリーと主要な栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。 @%s", absPath)

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
	if strings.Contains(response.Response, "```json") {
		start := strings.Index(response.Response, "```json") + 7
		end := strings.LastIndex(response.Response, "```")
		if start > 7 && end > start {
			response.Response = strings.TrimSpace(response.Response[start:end])
		}
	}

	return &response, nil
}

func main() {
	// 設定
	config := AnalyzeImageConfig{
		ImagePath: "images/IMG_0374.JPG",
		Timeout:   60 * time.Second,
	}

	fmt.Println("🔍 Gemini CLIで画像を分析中...")
	fmt.Printf("画像: %s\n", config.ImagePath)
	fmt.Printf("タイムアウト: %v\n\n", config.Timeout)

	// 画像分析を実行
	ctx := context.Background()
	response, err := AnalyzeImage(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ エラー: %v\n", err)
		os.Exit(1)
	}

	// 結果を表示
	fmt.Println("✅ 分析完了")
	fmt.Printf("セッションID: %s\n\n", response.SessionID)
	fmt.Println("📊 レスポンス:")
	fmt.Println(response.Response)

	// 結果をファイルに保存
	resultPath := "results/golang_result.json"
	resultData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  結果のJSON化エラー: %v\n", err)
		return
	}

	if err := os.WriteFile(resultPath, resultData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  ファイル保存エラー: %v\n", err)
		return
	}

	fmt.Printf("\n💾 結果を保存しました: %s\n", resultPath)
}
