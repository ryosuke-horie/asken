# テスト詳細

## 基本方針

- Goのテストは標準ライブラリの`testing`パッケージとtestifyを使用すること
- テストカバレッジは80%以上を目標とすること
- iOSのテストについては`10-ios-testing.md`を参照

## 単体テストの書き方

古典派（Classicist）のテストスタイルを採用する。

- テストは仕様を表現するドキュメントとして機能させる
- モックは外部依存（API、DB）に限定し、内部実装のモックは避ける
- 実際のオブジェクトを使用して振る舞いを検証する

## テスト構造（Go）

テーブル駆動テストを使用:

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

## モックの使用基準

| 対象 | モック可否 | 理由 |
| :--- | :--------- | :--- |
| 外部API | 可 | ネットワーク依存を排除 |
| データベース | 可 | テスト実行速度と独立性 |
| 現在時刻 | 可 | 再現性の確保 |
| 内部クラス | 不可 | 実装詳細への依存を避ける |
| ユーティリティ関数 | 不可 | 実際の振る舞いを検証 |

## 禁止事項

- 実装詳細に依存したテスト（プライベートメソッドの直接テスト等）
- モックだらけで実際の振る舞いを検証していないテスト
- テスト間の依存関係（順序依存、共有状態）

## エージェントサポート

| エージェント | 用途 | 使用タイミング |
| :--- | :--- | :--- |
| tdd-guide | TDD方法論の徹底 | 新機能実装、バグ修正時 |
