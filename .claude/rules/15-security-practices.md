---
paths:
  - "backend/**/*"
  - "ios/**/*"
---

# セキュリティガイドライン

## バックエンド

### SQLインジェクション対策

- プレースホルダを使用（パラメータバインディング）
- クエリ文字列に直接ユーザー入力を埋め込まない

```go
// 良い例 - SQLインジェクション対策
query := "SELECT * FROM foods WHERE name LIKE $1"
rows, err := db.QueryContext(ctx, query, "%"+searchTerm+"%")

// 悪い例 - SQLインジェクション脆弱性
query := fmt.Sprintf("SELECT * FROM foods WHERE name LIKE '%%%s%%'", searchTerm)
rows, err := db.QueryContext(ctx, query)
```

### 入力検証

- すべてのユーザー入力をバリデーション
- ホワイトリスト方式を採用

```go
// 良い例 - ホワイトリストバリデーション
func validateSortOrder(order string) error {
    allowedOrders := map[string]bool{
        "asc":  true,
        "desc": true,
    }
    if !allowedOrders[order] {
        return errors.New("invalid sort order")
    }
    return nil
}
```

### 機密情報の管理

- 機密情報: 環境変数で管理し、コードに直接記述しない
- APIキー、パスワードはコミットしない

```go
// 良い例 - 環境変数から取得
apiKey := os.Getenv("GEMINI_API_KEY")

// 悪い例 - ハードコード
apiKey := "sk-1234567890abcdef"
```

## 認証・CORS

- 認証・認可: Cloud Run環境でFirebase認証済みユーザーのみアクセス
- CORS: iOSアプリからのアクセスのため、必要に応じて設定

## ロギング

- 機密情報をログに出力しない（APIキー、パスワード）
- エラーログには十分な文脈情報を含める

```go
// 良い例
log.Printf("failed to fetch food: foodID=%s, error=%v", foodID, err)

// 悪い例 - 機密情報をログに出力
log.Printf("API key: %s", apiKey)
```
