---
paths:
  - "**/*test*.{go,swift}"
  - "**/tests/**/*"
  - "**/__tests__/**/*"
---

# テスト駆動開発（TDD）

## テストの原則

1. 新機能実装時は必ずテストを先に書く（Red → Green → Refactor）
2. テストカバレッジは80%以上を目標とする
3. テストは独立性を保ち、実行順序に依存しない
4. テストコードもリファクタリング対象とする

## バックエンドテスト（Go）

### ツール

- 標準ライブラリ: `testing`パッケージ
- testify: アサーションとモック

### テスト方針

- テーブル駆動テストを使用
- 依存性注入でモックを容易にする
- Gemini HTTP APIはモックする（実際のAPI呼び出しはテストしない）

```go
// ✅ 良い例 - Gemini HTTP APIのモック
type mockGeminiCaller struct {
    mock.Mock
}

func (m *mockGeminiCaller) AnalyzeImage(ctx context.Context, imagePath string) (*AnalysisResult, error) {
    args := m.Called(ctx, imagePath)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*AnalysisResult), args.Error(1)
}

func TestFoodService_AnalyzeFoodImage(t *testing.T) {
    mockGemini := new(mockGeminiCaller)
    mockGemini.On("AnalyzeImage", mock.Anything, "/path/to/image.jpg").
        Return(&AnalysisResult{
            Foods: []string{"ごはん", "味噌汁"},
            Calories: 450,
        }, nil)

    service := NewFoodService(mockGemini, nil)
    result, err := service.AnalyzeFoodImage(context.Background(), "/path/to/image.jpg")

    require.NoError(t, err)
    assert.Equal(t, 450, result.Calories)
    mockGemini.AssertExpectations(t)
}
```

## iOSテスト

iOSテストの詳細は`10-ios-testing.md`を参照。

### 概要

- Swift Testing: ユニットテスト
- XCUITest: UIテスト
- Mockolo: モック生成
- swift-snapshot-testing: スナップショットテスト

## テスト実行

```bash
# バックエンド
cd backend
go test ./...         # すべてのテスト
go test -cover ./...  # カバレッジ付き
go test -v ./...      # 詳細モード

# iOS
task ios:test
```
