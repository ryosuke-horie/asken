# コーディングスタイル

## 基本原則

| 原則 | 内容 |
|:---|:---|
| KISS | 動作する最もシンプルな解決策を選ぶ |
| DRY | 共通ロジックを関数に抽出、コピペ禁止 |
| YAGNI | 必要になる前に機能を構築しない |

## イミュータビリティ（重要）

オブジェクトの直接変更は禁止。常に新しいオブジェクトを作成すること。

```typescript
// NG: ミューテーション
function updateUser(user: User, name: string) {
  user.name = name  // 直接変更は禁止
  return user
}

// OK: イミュータブル
function updateUser(user: User, name: string) {
  return {
    ...user,
    name
  }
}
```

## ファイル構成

小さなファイルを多数作成することを推奨:

| 指標 | 基準 |
| :--- | :--- |
| ファイル行数（理想） | 200-400行 |
| ファイル行数（最大） | 800行 |
| 関数行数（最大） | 50行 |
| ネスト深度（最大） | 4レベル |

- 高凝集・低結合を意識する
- 大きなコンポーネントからユーティリティを抽出する

## エラーハンドリング

エラーは必ず適切に処理すること:

```typescript
try {
  const result = await riskyOperation()
  return result
} catch (error) {
  // console.errorは適切なエラーロギング用途に限り許容
  console.error('Operation failed:', error)
  throw new Error('ユーザーにわかりやすいエラーメッセージ')
}
```

## 入力検証

ユーザー入力は必ず検証すること:

```typescript
function validateInput(input: unknown) {
  if (typeof input !== 'string') {
    throw new Error('Invalid input type')
  }
  if (input.length > 1000) {
    throw new Error('Input too long')
  }
  return input.trim()
}
```

## コード品質チェックリスト

作業完了前に確認:

- [ ] コードが読みやすく、適切な命名がされている
- [ ] 関数が小さい（50行未満）
- [ ] ファイルが集中している（800行未満）
- [ ] 深いネストがない（4レベル以下）
- [ ] 適切なエラーハンドリングがされている
- [ ] console.log文がない
- [ ] ハードコードされた値がない
- [ ] ミューテーションがない（イミュータブルパターンを使用）

## 非同期処理

独立した処理は並列実行すること:

```typescript
// OK: 並列実行
const [users, meals, stats] = await Promise.all([
  fetchUsers(),
  fetchMeals(),
  fetchStats()
])

// NG: 不必要な順次実行
const users = await fetchUsers()
const meals = await fetchMeals()
const stats = await fetchStats()
```

## 早期リターン

深いネストは早期リターンで解消すること:

```typescript
// NG: 深いネスト
if (user) {
  if (user.isValid) {
    if (meal) {
      // 処理
    }
  }
}

// OK: 早期リターン
if (!user) return
if (!user.isValid) return
if (!meal) return
// 処理
```

## 禁止事項

- any型の使用（型が不明な場合はunknownを使用）
- console.logの本番コードへの残存（console.errorは適切なエラーロギング用途に限り許容）
- 未使用のimport/変数の放置
- コメントアウトされたコードの放置
- マジックナンバー/マジックストリングの直接使用
- 環境変数の直接参照（設定ファイル経由で取得）
