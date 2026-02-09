---
name: tdd-workflow
description: Goバックエンドの新機能実装、バグ修正、リファクタリング時に使用するスキル。テスト駆動開発を強制し、80%以上のカバレッジを確保。
---

# テスト駆動開発ワークフロー

> `注意: iOSテストは一時無効化されています`
>
> このスキルは Go バックエンドの実装時にのみ使用してください。
> iOS の開発時には使用しないでください。
>
> macOS/Xcode バージョン問題により、iOS テストを一時的に無効化しています。
> 詳細は `.claude/rules/ios-testing-policy.md` を参照してください。

このスキルは Go バックエンドのコード開発が TDD 原則に従い、包括的なテストカバレッジを確保することを保証します。

## 使用タイミング

- Go バックエンドの新機能実装時
- Go バックエンドのバグ修正時
- Go バックエンドのリファクタリング時
- APIエンドポイントの追加時

## 基本原則

### 1. コードの前にテスト

常にテストを先に書き、その後テストを通すコードを実装する。

### 2. カバレッジ要件

- 最低80%カバレッジ（ユニット + 統合）
- すべてのエッジケースをカバー
- エラーシナリオをテスト
- 境界条件を検証

### 3. テストの種類

#### ユニットテスト

- Go: `testing`パッケージ + testify

#### 統合テスト

- APIエンドポイント（httptest）
- Firestoreエミュレータを使用したデータベース操作

## TDDワークフローステップ

### ステップ1: テストを先に書く（RED）

```go
func TestFoodService_SearchFood(t *testing.T) {
    mockRepo := new(mockFoodRepository)
    mockRepo.On("SearchByName", mock.Anything, "ごはん").
        Return(&Food{Name: "ごはん", Calories: 250}, nil)
    service := NewFoodService(mockRepo, nil)

    result, err := service.SearchFood(context.Background(), "ごはん")

    require.NoError(t, err)
    assert.Equal(t, 250, result.Calories)
}
```

### ステップ2: テスト実行（失敗を確認）

```bash
task test
```

### ステップ3: テストを通す実装を書く（GREEN）

テストを通すコードを書く。過剰な先回り実装は避け、テストが求める振る舞いに集中する。

### ステップ4: テスト実行（成功を確認）

```bash
task test
```

### ステップ5: リファクタリング

テストをグリーンに保ちながらコード品質を改善。

### ステップ6: カバレッジを確認

```bash
task test:coverage
```

## テストパターン

### Goテーブル駆動テスト

```go
func TestCalculateCalories(t *testing.T) {
    tests := []struct {
        name     string
        input    Food
        expected int
    }{
        {"valid food", Food{Protein: 10, Fat: 5, Carbs: 20}, 165},
        {"zero values", Food{}, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateCalories(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## 外部サービスのモック

### Gemini APIモック（Go）

```go
type mockGeminiCaller struct {
    mock.Mock
}

func (m *mockGeminiCaller) AnalyzeImage(ctx context.Context, data []byte, mimeType string) (*AnalysisResult, error) {
    args := m.Called(ctx, data, mimeType)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*AnalysisResult), args.Error(1)
}
```

## テストカバレッジ確認

```bash
task test:coverage
```

## ベストプラクティス

1. テストを先に書く - 常にTDD
2. 1テスト1アサート - 単一の振る舞いに集中
3. 説明的なテスト名 - 何をテストしているか説明
4. Arrange-Act-Assert - 明確なテスト構造
5. 外部依存をモック - ユニットテストを分離
6. エッジケースをテスト - nil、ゼロ値、空、大きな値
7. エラーパスをテスト - ハッピーパスだけでなく
8. テストを高速に - ユニットテストは各50ms未満

## 成功指標

- 80%以上のコードカバレッジ達成
- すべてのテストがパス（グリーン）
- スキップまたは無効化されたテストがない
- 高速なテスト実行

---

テストはオプションではありません。自信を持ったリファクタリング、迅速な開発、本番の信頼性を可能にするセーフティネットです。
