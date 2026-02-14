package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
func NewNutritionCalculator(timeout time.Duration) (*NutritionCalculator, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	httpClient, err := NewHTTPClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create nutrition calculator: %w", err)
	}
	return &NutritionCalculator{
		httpClient: httpClient,
	}, nil
}

// NewNutritionCalculatorWithAPIKey はAPIキーを指定してNutritionCalculatorを作成
func NewNutritionCalculatorWithAPIKey(apiKey string, timeout time.Duration) (*NutritionCalculator, error) {
	httpClient, err := NewHTTPClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create nutrition calculator: %w", err)
	}
	return &NutritionCalculator{
		httpClient: httpClient,
	}, nil
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

	log.Printf("NutritionCalculator: 栄養素計算を開始 (%d品)", len(foods))

	// 食材リストをJSON形式に変換
	foodListJSON, err := json.MarshalIndent(foods, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("食材リストのJSON化エラー: %w", err)
	}

	// プロンプトを構築（栄養素算出のみに集中）
	prompt := fmt.Sprintf(`以下の食材リストについて、それぞれのカロリーと栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。

食材リスト:
%s

推定量のルール:
- quantity_valueは数値で指定してください（入力の推定量をそのまま数値に変換）
- quantity_unitは以下のいずれかを使用してください:
  g, ml, 杯, 人前, 個, 枚, 本, 切れ, 食, 皿, 膳, 丁, 束, 袋, 缶, 合, 玉, 粒
- 入力の推定量に含まれる単位をそのまま使用してください

一般的な食品成分表に基づいて、妥当な値を推定してください。`, string(foodListJSON))

	// responseSchemaで出力形式を強制
	schema := NutritionInfoSchema()

	// Gemini APIを呼び出す
	response, err := nc.httpClient.Execute(ctx, prompt, schema)
	if err != nil {
		log.Printf("NutritionCalculator: Gemini API呼び出しエラー: %v", err)
		return nil, fmt.Errorf("Gemini API呼び出しエラー: %w", err)
	}

	// レスポンス内のJSONコードブロックを抽出
	nutritionJSON := removeCodeBlock(response.Response)

	// スキーマ制約付きレスポンスをパース
	var items []nutritionResponseItem
	if err := json.Unmarshal([]byte(nutritionJSON), &items); err != nil {
		log.Printf("NutritionCalculator: 栄養素情報のJSONパースエラー: %v", err)
		return nil, fmt.Errorf("栄養素情報のパースエラー: %w\nデータ: %s", err, nutritionJSON)
	}

	// NutritionInfoに変換（quantity_value + quantity_unit → estimated_amount）
	nutritionList := make([]NutritionInfo, len(items))
	for i, item := range items {
		nutritionList[i] = item.toNutritionInfo()
	}

	log.Printf("NutritionCalculator: 栄養素計算完了 (%d品)", len(nutritionList))
	return nutritionList, nil
}
