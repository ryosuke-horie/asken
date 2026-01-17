---
paths:
  - "frontend/**/*"
  - "backend/**/*"
---

# セキュリティガイドライン

## フロントエンド

### XSS対策

- ユーザー入力は必ずエスケープ（Reactはデフォルトでエスケープ）
- HTMLを直接レンダリングする機能（dangerouslySetInnerHTML等）の使用は避ける

```typescript
// ✅ 良い例 - Reactは自動的にエスケープする
<div>{foodName}</div>
```

### 機密情報の管理

- **機密情報**: クライアントサイドに機密情報を保存しない
- **環境変数**: `NEXT_PUBLIC_`接頭辞のついた変数のみクライアントに公開される

```typescript
// ✅ 良い例
const apiUrl = process.env.NEXT_PUBLIC_API_URL;  // クライアントで利用可能
const dbPassword = process.env.DB_PASSWORD;       // サーバーサイドのみ
```

## バックエンド

### SQLインジェクション対策

- **プレースホルダ**を使用（パラメータバインディング）
- クエリ文字列に直接ユーザー入力を埋め込まない

```go
// ✅ 良い例 - SQLインジェクション対策
query := "SELECT * FROM foods WHERE name LIKE $1"
rows, err := db.QueryContext(ctx, query, "%"+searchTerm+"%")

// ❌ 悪い例 - SQLインジェクション脆弱性
query := fmt.Sprintf("SELECT * FROM foods WHERE name LIKE '%%%s%%'", searchTerm)
rows, err := db.QueryContext(ctx, query)
```

### 入力検証

- すべてのユーザー入力をバリデーション
- ホワイトリスト方式を採用

```go
// ✅ 良い例 - ホワイトリストバリデーション
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

- **機密情報**: 環境変数で管理し、コードに直接記述しない
- APIキー、パスワードはコミットしない

```go
// ✅ 良い例 - 環境変数から取得
apiKey := os.Getenv("GEMINI_API_KEY")

// ❌ 悪い例 - ハードコード
apiKey := "sk-1234567890abcdef"
```

### コマンドインジェクション対策

- **Gemini CLI実行**: ファイルパスのサニタイズ
- ユーザー入力を直接コマンド引数に渡さない

```go
// ✅ 良い例 - コマンドインジェクション対策
func sanitizeImagePath(path string) (string, error) {
    // パスのバリデーション
    if !filepath.IsAbs(path) {
        return "", errors.New("path must be absolute")
    }

    // ディレクトリトラバーサル対策
    cleanPath := filepath.Clean(path)
    if !strings.HasPrefix(cleanPath, allowedUploadDir) {
        return "", errors.New("path outside allowed directory")
    }

    return cleanPath, nil
}
```

## 個人利用のための簡略化

- **認証・認可**: MVP段階では不要（exe.dev環境内のみでアクセス）
- **CORS**: 必要に応じて設定（フロントエンドとバックエンドが同一ドメインの場合は不要）

## ロギング

- **機密情報**をログに出力しない（APIキー、パスワード）
- エラーログには十分な文脈情報を含める

```go
// ✅ 良い例
log.Printf("failed to fetch food: foodID=%s, error=%v", foodID, err)

// ❌ 悪い例 - 機密情報をログに出力
log.Printf("API key: %s", apiKey)
```
