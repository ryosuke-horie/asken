package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

// NutritionInfo は栄養素情報を表す構造体
type NutritionInfo struct {
	Name            string  `json:"name"`
	EstimatedAmount string  `json:"estimated_amount"`
	Calories        float64 `json:"calories_kcal"`
	Protein         float64 `json:"protein_g"`
	Fat             float64 `json:"fat_g"`
	Carbohydrates   float64 `json:"carbohydrates_g"`
	Source          string  `json:"source"` // "database" or "gemini"
}

// CalculateNutritionConfig は栄養素算出の設定
type CalculateNutritionConfig struct {
	FoodItems []FoodItem
	Timeout   time.Duration
}

// CalculateNutritionWithGemini はGemini CLIを使って栄養素を算出する
func CalculateNutritionWithGemini(ctx context.Context, config CalculateNutritionConfig) ([]NutritionInfo, error) {
	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// 食材リストをJSON形式に変換
	foodListJSON, err := json.MarshalIndent(config.FoodItems, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("食材リストのJSON化エラー: %w", err)
	}

	// プロンプトを構築（栄養素算出のみに集中）
	prompt := fmt.Sprintf(`以下の食材リストについて、それぞれのカロリーと栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。

食材リスト:
%s

出力フォーマット:
[
  {
    "name": "食材名",
    "estimated_amount": "推定量",
    "calories_kcal": カロリー数値,
    "protein_g": タンパク質グラム数,
    "fat_g": 脂質グラム数,
    "carbohydrates_g": 炭水化物グラム数
  }
]

一般的な食品成分表に基づいて、妥当な値を推定してください。`, string(foodListJSON))

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

	// JSON部分を抽出
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

	// レスポンス内のJSONコードブロックを抽出
	nutritionJSON := response.Response
	if strings.Contains(nutritionJSON, "```") {
		// ```で分割して、コードブロック内のテキストを抽出
		parts := strings.Split(nutritionJSON, "```")
		if len(parts) >= 3 {
			content := parts[1]
			// "json" プレフィックスを除去
			content = strings.TrimPrefix(content, "json")
			content = strings.TrimPrefix(content, "JSON")
			nutritionJSON = strings.TrimSpace(content)
		}
	}

	// 栄養素情報をパース
	var nutritionList []NutritionInfo
	if err := json.Unmarshal([]byte(nutritionJSON), &nutritionList); err != nil {
		return nil, fmt.Errorf("栄養素情報のパースエラー: %w\nデータ: %s", err, nutritionJSON)
	}

	// ソース情報を追加
	for i := range nutritionList {
		nutritionList[i].Source = "gemini"
	}

	return nutritionList, nil
}

// TODO: 将来的にはPostgreSQLから検索する関数を実装
// func SearchNutritionFromDatabase(foodName string, amount string) (*NutritionInfo, error) {
//     // PostgreSQLから食品マスタを検索
//     // ヒットした場合: DB の値を返す（Source = "database"）
//     // ヒットしない場合: nil を返す
// }

func main() {
	// ステップ1の結果を読み込む
	step1ResultPath := "results/step1_classify_result.json"
	step1Data, err := os.ReadFile(step1ResultPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ エラー: ステップ1の結果ファイルが読み込めません: %v\n", err)
		fmt.Fprintf(os.Stderr, "先に step1_classify.go を実行してください\n")
		os.Exit(1)
	}

	// 食材リストをパース
	var foods []FoodItem
	if err := json.Unmarshal(step1Data, &foods); err != nil {
		fmt.Fprintf(os.Stderr, "❌ エラー: ステップ1の結果のパースに失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🥗 ステップ2: 栄養素算出")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("食材数: %d個\n", len(foods))
	fmt.Printf("タイムアウト: 60秒\n\n")

	// 栄養素算出を実行
	ctx := context.Background()
	config := CalculateNutritionConfig{
		FoodItems: foods,
		Timeout:   60 * time.Second,
	}

	nutritionList, err := CalculateNutritionWithGemini(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ エラー: %v\n", err)
		os.Exit(1)
	}

	// 結果を表示
	fmt.Println("✅ 栄養素算出完了\n")

	var totalCalories, totalProtein, totalFat, totalCarbs float64

	for i, nutrition := range nutritionList {
		fmt.Printf("%d. %s (%s)\n", i+1, nutrition.Name, nutrition.EstimatedAmount)
		fmt.Printf("   カロリー: %.0f kcal\n", nutrition.Calories)
		fmt.Printf("   タンパク質: %.1f g\n", nutrition.Protein)
		fmt.Printf("   脂質: %.1f g\n", nutrition.Fat)
		fmt.Printf("   炭水化物: %.1f g\n", nutrition.Carbohydrates)
		fmt.Printf("   ソース: %s\n\n", nutrition.Source)

		totalCalories += nutrition.Calories
		totalProtein += nutrition.Protein
		totalFat += nutrition.Fat
		totalCarbs += nutrition.Carbohydrates
	}

	// 合計を表示
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("合計:\n")
	fmt.Printf("  カロリー: %.0f kcal\n", totalCalories)
	fmt.Printf("  タンパク質: %.1f g\n", totalProtein)
	fmt.Printf("  脂質: %.1f g\n", totalFat)
	fmt.Printf("  炭水化物: %.1f g\n", totalCarbs)

	// 結果をファイルに保存
	resultPath := "results/step2_nutrition_result.json"
	resultData, err := json.MarshalIndent(nutritionList, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  結果のJSON化エラー: %v\n", err)
		return
	}

	if err := os.WriteFile(resultPath, resultData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  ファイル保存エラー: %v\n", err)
		return
	}

	fmt.Printf("\n💾 結果を保存しました: %s\n", resultPath)
	fmt.Println("\n注意: このサンプルではGemini CLIで栄養素を推定しています。")
	fmt.Println("実際のアプリでは、PostgreSQLから検索 → 見つからない場合のみGeminiで推定することを推奨します。")
}
