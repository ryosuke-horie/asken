package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// NutritionInfo は栄養素情報を表す構造体
type NutritionInfo struct {
	Name            string  `json:"name"`
	EstimatedAmount string  `json:"estimated_amount"`
	Calories        float64 `json:"calories_kcal"`
	Protein         float64 `json:"protein_g"`
	Fat             float64 `json:"fat_g"`
	Carbohydrates   float64 `json:"carbohydrates_g"`
}

// NutritionCalculator は栄養素計算を行うクライアント
type NutritionCalculator struct {
	timeout time.Duration
}

// NewNutritionCalculator は新しいNutritionCalculatorを作成
func NewNutritionCalculator(timeout time.Duration) *NutritionCalculator {
	return &NutritionCalculator{
		timeout: timeout,
	}
}

// CalculateNutrition はGemini CLIを使って栄養素を算出する
func (nc *NutritionCalculator) CalculateNutrition(ctx context.Context, foods []FoodItem) ([]NutritionInfo, error) {
	// 食材リストが空の場合はエラー
	if len(foods) == 0 {
		return nil, fmt.Errorf("画像から食材を認識できませんでした。食べ物が写っている鮮明な画像を使用してください")
	}

	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, nc.timeout)
	defer cancel()

	// 食材リストをJSON形式に変換
	foodListJSON, err := json.MarshalIndent(foods, "", "  ")
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
			return nil, fmt.Errorf("タイムアウト: Gemini CLIの実行が%v以内に完了しませんでした", nc.timeout)
		}
		return nil, fmt.Errorf("Gemini CLI実行エラー: %w\n出力: %s", err, string(output))
	}

	// JSON部分を抽出
	jsonData := extractJSON(output)
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("JSON開始位置が見つかりません\n生データ: %s", string(output))
	}

	// JSONレスポンスをパース
	var response Response
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w\n生データ: %s", err, string(jsonData))
	}

	// レスポンス内のJSONコードブロックを抽出
	nutritionJSON := removeCodeBlock(response.Response)

	// 栄養素情報をパース
	var nutritionList []NutritionInfo
	if err := json.Unmarshal([]byte(nutritionJSON), &nutritionList); err != nil {
		return nil, fmt.Errorf("栄養素情報のパースエラー: %w\nデータ: %s", err, nutritionJSON)
	}

	return nutritionList, nil
}
