---
paths:
  - "backend/**/*.go"
---

# バックエンド開発規約（Golang）

## 命名規則

- パッケージ: 小文字、単数形（例: `food`, `gemini`）
- 変数・関数: camelCase、エクスポートするものはPascalCase（例: `getFoodByID`, `FoodID`）
- 定数: PascalCaseまたはcamelCase（例: `MaxRetries`, `defaultTimeout`）
- インターフェース: PascalCase、`-er`で終わる（例: `FoodRepository`, `GeminiCaller`）

## パッケージ構成

- internal/: プロジェクト内部のみで使用するパッケージ
- pkg/: 外部からインポート可能なパッケージ（Gemini HTTP APIクライアント等の共有パッケージ）
- レイヤードアーキテクチャを採用（Handler → Service → Repository）

## エラーハンドリング

- エラーは必ず処理する - `_`でエラーを無視しない
- エラーはラップして文脈情報を追加する（`fmt.Errorf`使用）
- カスタムエラー型を定義する場合は`errors.New`を使用

```go
// 良い例
food, err := repo.GetFoodByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get food %s: %w", id, err)
}

// 悪い例
food, _ := repo.GetFoodByID(ctx, id)  // エラーを無視
```

## Gemini HTTP API連携

- Gemini APIは`pkg/gemini/http_client.go`のHTTPクライアントで直接通信する
- タイムアウト処理を実装する（context.WithTimeout）
- APIキーはヘッダー（`x-goog-api-key`）で送信する
- レスポンスのパースエラーを適切に処理する
- ステータスコード別のエラーハンドリングを実装する

```go
// 良い例 - HTTP APIクライアントの使用
func (s *GeminiService) AnalyzeImage(ctx context.Context, imageData []byte, mimeType string) (*AnalysisResult, error) {
    resp, err := s.client.ExecuteWithImage(ctx, analyzeImagePrompt, imageData, mimeType)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze image: %w", err)
    }

    var result AnalysisResult
    if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
        return nil, fmt.Errorf("failed to parse gemini response: %w", err)
    }

    return &result, nil
}
```

## コンテキスト使用

- すべての外部呼び出し（DB、Gemini HTTP API）にはcontextを渡す
- context.Backgroundは`main()`または`init()`のみで使用
- タイムアウトとキャンセル処理を適切に実装

## 構造体とインターフェース

- 構造体フィールドはJSONタグを必ず定義
- 小さなインターフェースを定義（1〜3メソッド）
- インターフェースは使用する側で定義

```go
// 良い例
type Food struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Calories    int       `json:"calories" db:"calories"`
    Protein     float64   `json:"protein" db:"protein"`      // g
    Fat         float64   `json:"fat" db:"fat"`              // g
    Carbs       float64   `json:"carbs" db:"carbs"`          // g
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type FoodRepository interface {
    GetByID(ctx context.Context, id string) (*Food, error)
    Search(ctx context.Context, query string) ([]*Food, error)
    Create(ctx context.Context, food *Food) error
}

type GeminiCaller interface {
    AnalyzeImage(ctx context.Context, imagePath string) (*AnalysisResult, error)
    SearchFood(ctx context.Context, foodName string) (*NutritionInfo, error)
}
```

## 並列処理

Gemini API呼び出しなど、独立した操作は並列化を検討する：

```go
// 悪い例 - 逐次実行
result1, err := gemini.AnalyzeFood(ctx, food1)
if err != nil {
    return err
}
result2, err := gemini.AnalyzeFood(ctx, food2)
if err != nil {
    return err
}

// 良い例 - 並列実行
var wg sync.WaitGroup
var mu sync.Mutex
results := make([]*AnalysisResult, 2)
errs := make([]error, 2)

wg.Add(2)
go func() {
    defer wg.Done()
    result, err := gemini.AnalyzeFood(ctx, food1)
    mu.Lock()
    results[0], errs[0] = result, err
    mu.Unlock()
}()

go func() {
    defer wg.Done()
    result, err := gemini.AnalyzeFood(ctx, food2)
    mu.Lock()
    results[1], errs[1] = result, err
    mu.Unlock()
}()

wg.Wait()

// エラーチェック
for _, err := range errs {
    if err != nil {
        return nil, err
    }
}
```
