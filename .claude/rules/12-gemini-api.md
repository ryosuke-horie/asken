---
paths:
  - "backend/**/*.go"
  - "backend/pkg/gemini/**/*"
---

# Gemini API連携のベストプラクティス

## レスポンス形式の統一

- Gemini APIには特定の形式でレスポンスを返すようプロンプトで指示
- JSON形式を推奨（パースしやすい）

```go
// プロンプト例
const analyzeImagePrompt = `
この画像に写っている食品を分析し、以下のJSON形式で返してください：

{
  "foods": [
    {
      "name": "食品名",
      "estimated_amount": "推定量（g）",
      "calories": 推定カロリー（数値）,
      "protein": タンパク質（g）,
      "fat": 脂質（g）,
      "carbs": 炭水化物（g）
    }
  ],
  "total_calories": 合計カロリー（数値）
}
`
```

## エラーハンドリング

- Gemini APIのエラーを適切に処理
- タイムアウト、レート制限、API障害に対応
- フォールバック処理（データベース検索のみ）を実装

```go
func (s *FoodService) SearchFood(ctx context.Context, foodName string) (*NutritionInfo, error) {
    // 1. データベース検索
    food, err := s.repo.SearchByName(ctx, foodName)
    if err == nil && food != nil {
        return food.ToNutritionInfo(), nil
    }

    // 2. Gemini APIで検索
    geminiResult, err := s.geminiCaller.SearchFood(ctx, foodName)
    if err != nil {
        return nil, fmt.Errorf("food not found in db or gemini: %w", err)
    }

    // 3. データベースにキャッシュ
    newFood := &Food{
        Name:     foodName,
        Calories: geminiResult.Calories,
        Protein:  geminiResult.Protein,
        Fat:      geminiResult.Fat,
        Carbs:    geminiResult.Carbs,
        Source:   "gemini",
    }
    _ = s.repo.Create(ctx, newFood)  // エラーは無視（キャッシュ失敗は許容）

    return geminiResult, nil
}
```

## データベースとの組み合わせ

1. まずデータベースを検索（高速、正確）
2. データベースに存在しない場合はGemini APIを使用
3. Gemini APIの結果をデータベースにキャッシュ（次回以降の高速化）

## セキュリティ

- APIキーはヘッダー（`x-goog-api-key`）で送信する（URLパラメータに含めない）
- HTTP APIリクエストにはタイムアウトを設定する（context.WithTimeout）
- レスポンスのパース失敗を適切に処理する
- エラーレスポンスに機密情報（APIキー等）を含めない

```go
// ✅ 良い例 - APIキーをヘッダーで送信
req.Header.Set("x-goog-api-key", c.apiKey)

// ✅ 良い例 - タイムアウト付きコンテキスト
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```
