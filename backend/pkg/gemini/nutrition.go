package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	httpClient GeminiHTTPClient
}

// NewNutritionCalculator は新しいNutritionCalculatorを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewNutritionCalculator(timeout time.Duration) *NutritionCalculator {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return &NutritionCalculator{
		httpClient: NewHTTPClient(apiKey, timeout),
	}
}

// NewNutritionCalculatorWithAPIKey はAPIキーを指定してNutritionCalculatorを作成
func NewNutritionCalculatorWithAPIKey(apiKey string, timeout time.Duration) *NutritionCalculator {
	return &NutritionCalculator{
		httpClient: NewHTTPClient(apiKey, timeout),
	}
}

// NewNutritionCalculatorWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewNutritionCalculatorWithHTTPClient(httpClient GeminiHTTPClient) *NutritionCalculator {
	return &NutritionCalculator{
		httpClient: httpClient,
	}
}

// CalculateNutrition はGemini APIを使って栄養素を算出する
func (nc *NutritionCalculator) CalculateNutrition(ctx context.Context, foods []FoodItem) ([]NutritionInfo, error) {
	// 食材リストが空の場合はエラー
	if len(foods) == 0 {
		return nil, fmt.Errorf("画像から食材を認識できませんでした。食べ物が写っている鮮明な画像を使用してください")
	}

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

	// Gemini APIを呼び出す
	response, err := nc.httpClient.Execute(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("Gemini API呼び出しエラー: %w", err)
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
