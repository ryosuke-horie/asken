# プラン: Gemini HTTPClient にインターフェース導入

## Linear Issue
- Issue: EDG-669
- URL: https://linear.app/ryosuke-horie/issue/EDG-669/テスト評価-pkggemini-httpclient-にインターフェース導入

## 概要

`backend/pkg/gemini/` パッケージの `HTTPClient` にインターフェースを導入し、モックベースのユニットテストを可能にする。

## 現状の課題

1. `HTTPClient` が構造体として実装されており、インターフェースがない
2. `Client`, `Classifier`, `NutritionCalculator` が `*HTTPClient` を直接保持
3. `httptest.NewServer` を使った統合テストに依存しており、CI環境で実行可能なテストが限られている

## 解決策

`GeminiHTTPClient` インターフェースを定義し、各クライアント構造体がこのインターフェースに依存するように変更する。後方互換性を維持しつつ、テスト用にモックを注入可能にする。

---

## 実装計画

### フェーズ1: インターフェースとモックの定義

#### 1.1 `http_client.go` にインターフェース追加

```go
// GeminiHTTPClient はGemini APIのHTTPクライアントを表すインターフェース
type GeminiHTTPClient interface {
    Execute(ctx context.Context, prompt string) (*Response, error)
    ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error)
}
```

#### 1.2 `mock_http_client.go` を新規作成

```go
package gemini

import "context"

// MockGeminiHTTPClient はGeminiHTTPClientのモック実装
type MockGeminiHTTPClient struct {
    ExecuteFunc          func(ctx context.Context, prompt string) (*Response, error)
    ExecuteWithImageFunc func(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error)
}

func (m *MockGeminiHTTPClient) Execute(ctx context.Context, prompt string) (*Response, error) {
    if m.ExecuteFunc != nil {
        return m.ExecuteFunc(ctx, prompt)
    }
    return &Response{Response: `[]`}, nil
}

func (m *MockGeminiHTTPClient) ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error) {
    if m.ExecuteWithImageFunc != nil {
        return m.ExecuteWithImageFunc(ctx, prompt, imageData, mimeType)
    }
    return &Response{Response: `[]`}, nil
}
```

### フェーズ2: Client の修正

#### 2.1 `client.go` の `Client` 構造体を修正

```go
// 変更前
type Client struct {
    httpClient *HTTPClient
}

// 変更後
type Client struct {
    httpClient GeminiHTTPClient
}
```

#### 2.2 `client.go` にテスト用コンストラクタ追加

```go
// NewClientWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewClientWithHTTPClient(httpClient GeminiHTTPClient) *Client {
    return &Client{
        httpClient: httpClient,
    }
}
```

### フェーズ3: Classifier の修正

#### 3.1 `classifier.go` の `Classifier` 構造体を修正

```go
// 変更前
type Classifier struct {
    httpClient *HTTPClient
}

// 変更後
type Classifier struct {
    httpClient GeminiHTTPClient
}
```

#### 3.2 `classifier.go` にテスト用コンストラクタ追加

```go
// NewClassifierWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewClassifierWithHTTPClient(httpClient GeminiHTTPClient) *Classifier {
    return &Classifier{
        httpClient: httpClient,
    }
}
```

### フェーズ4: NutritionCalculator の修正

#### 4.1 `nutrition.go` の `NutritionCalculator` 構造体を修正

```go
// 変更前
type NutritionCalculator struct {
    httpClient *HTTPClient
}

// 変更後
type NutritionCalculator struct {
    httpClient GeminiHTTPClient
}
```

#### 4.2 `nutrition.go` にテスト用コンストラクタ追加

```go
// NewNutritionCalculatorWithHTTPClient はHTTPClientインターフェースを受け取るコンストラクタ（テスト用）
func NewNutritionCalculatorWithHTTPClient(httpClient GeminiHTTPClient) *NutritionCalculator {
    return &NutritionCalculator{
        httpClient: httpClient,
    }
}
```

### フェーズ5: TextParser の修正

`TextParser` は `Client` を経由して `HTTPClient` を使用しているため、`Client` のみ `NewTextParserWithClient()` を追加。

```go
// NewTextParserWithClient はClientインターフェースを受け取るコンストラクタ（テスト用）
func NewTextParserWithClient(client *Client) *TextParser {
    return &TextParser{
        client: client,
    }
}
```

### フェーズ6: テスト追加

#### 6.1 `classifier_test.go` にモックベーステスト追加

```go
func TestClassifier_ClassifyFoodsFromData_MockSuccess(t *testing.T) {
    mockHTTPClient := &MockGeminiHTTPClient{
        ExecuteWithImageFunc: func(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error) {
            return &Response{
                Response: `[{"name": "味噌ラーメン", "estimated_amount": "1杯"}]`,
            }, nil
        },
    }

    classifier := NewClassifierWithHTTPClient(mockHTTPClient)
    ctx := context.Background()

    imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
    foods, err := classifier.ClassifyFoodsFromData(ctx, imageData, "image/jpeg")

    require.NoError(t, err)
    assert.Len(t, foods, 1)
    assert.Equal(t, "味噌ラーメン", foods[0].Name)
}
```

#### 6.2 `nutrition_test.go` にモックベーステスト追加

```go
func TestNutritionCalculator_CalculateNutrition_MockSuccess(t *testing.T) {
    mockHTTPClient := &MockGeminiHTTPClient{
        ExecuteFunc: func(ctx context.Context, prompt string) (*Response, error) {
            return &Response{
                Response: `[{"name": "刺身盛り合わせ", "calories_kcal": 200}]`,
            }, nil
        },
    }

    calculator := NewNutritionCalculatorWithHTTPClient(mockHTTPClient)
    // ... テスト実装
}
```

---

## 技術的な考慮事項

1. **後方互換性**: 既存の `NewClient()`, `NewClassifier()`, `NewNutritionCalculator()` は維持
2. **既存パターンの準拠**: `MockTokenVerifier` と同様の関数ベースモックを使用
3. **小さなインターフェース**: 2メソッドのみを定義（`Execute`, `ExecuteWithImage`）
4. **モック配置**: `pkg/gemini/mock_http_client.go` に配置

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|:---|:---|
| `backend/pkg/gemini/http_client.go` | `GeminiHTTPClient` インターフェース追加 |
| `backend/pkg/gemini/mock_http_client.go` | 新規作成（モック実装） |
| `backend/pkg/gemini/client.go` | `Client.httpClient` 型変更、`NewClientWithHTTPClient()` 追加 |
| `backend/pkg/gemini/classifier.go` | `Classifier.httpClient` 型変更、`NewClassifierWithHTTPClient()` 追加 |
| `backend/pkg/gemini/nutrition.go` | `NutritionCalculator.httpClient` 型変更、`NewNutritionCalculatorWithHTTPClient()` 追加 |
| `backend/pkg/gemini/text_parser.go` | `NewTextParserWithClient()` 追加 |
| `backend/pkg/gemini/client_test.go` | モックベーステスト追加 |
| `backend/pkg/gemini/classifier_test.go` | モックベーステスト追加 |
| `backend/pkg/gemini/nutrition_test.go` | モックベーステスト追加 |

---

## 検証方法

1. 既存テストがすべてパスすることを確認
   ```bash
   task test
   ```

2. 新しいモックベーステストが CI 環境で実行できることを確認

3. カバレッジが向上していることを確認
   ```bash
   task test:coverage
   ```

---

## 参考ファイル

- `backend/internal/testutil/mock_token_verifier.go` - 関数ベースモックのパターン参考
- `backend/internal/middleware/auth.go` - インターフェースベース DI のパターン参考
