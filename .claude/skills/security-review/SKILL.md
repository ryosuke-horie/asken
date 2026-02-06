---
name: security-review
description: 認証の追加、ユーザー入力の処理、シークレットの扱い、APIエンドポイントの作成、機密機能の実装時に使用するスキル。包括的なセキュリティチェックリストとパターンを提供。
---

# セキュリティレビュースキル

このスキルはすべてのコードがセキュリティベストプラクティスに従い、潜在的な脆弱性を特定することを保証します。

## 使用タイミング

- 認証または認可の実装時
- ユーザー入力やファイルアップロードの処理時
- 新しいAPIエンドポイントの作成時
- シークレットや認証情報の扱い時
- 機密データの保存や送信時
- サードパーティAPIの統合時（Gemini API等）

## セキュリティチェックリスト

### 1. シークレット管理

#### NG: 絶対にやってはいけない

```go
apiKey := "AIzaSy-xxxxx"  // ハードコードされたシークレット
dbPassword := "password123" // ソースコード内
```

#### OK: 必ずこうする

```go
apiKey := os.Getenv("GEMINI_API_KEY")

// シークレットの存在を検証
if apiKey == "" {
    log.Fatal("GEMINI_API_KEY not configured")
}
```

#### 確認ステップ

- [ ] ハードコードされたAPIキー、トークン、パスワードがない
- [ ] すべてのシークレットが環境変数にある
- [ ] `.env.local`が.gitignoreにある
- [ ] git履歴にシークレットがない
- [ ] 本番シークレットがホスティングプラットフォームにある

### 2. 入力バリデーション

#### 常にユーザー入力をバリデート

```go
func validateInput(input string) (string, error) {
    if len(input) == 0 {
        return "", fmt.Errorf("input is empty")
    }
    if len(input) > 1000 {
        return "", fmt.Errorf("input too long")
    }
    return strings.TrimSpace(input), nil
}
```

#### ファイルアップロードバリデーション

```go
const maxFileSize = 5 * 1024 * 1024 // 5MB

var allowedContentTypes = map[string]bool{
    "image/jpeg": true,
    "image/png":  true,
    "image/gif":  true,
}

func validateFileUpload(header *multipart.FileHeader) error {
    // サイズチェック
    if header.Size > maxFileSize {
        return fmt.Errorf("file too large (max 5MB)")
    }

    // Content-Typeチェック
    contentType := header.Header.Get("Content-Type")
    if !allowedContentTypes[contentType] {
        return fmt.Errorf("invalid file type: %s", contentType)
    }

    // 拡張子チェック
    ext := strings.ToLower(filepath.Ext(header.Filename))
    allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
    if !allowedExts[ext] {
        return fmt.Errorf("invalid file extension: %s", ext)
    }

    return nil
}
```

#### 確認ステップ

- [ ] すべてのユーザー入力がバリデートされている
- [ ] ファイルアップロードが制限されている（サイズ、タイプ、拡張子）
- [ ] ユーザー入力がクエリで直接使用されていない
- [ ] ホワイトリストバリデーション（ブラックリストではない）
- [ ] エラーメッセージが機密情報を漏洩しない

### 3. SQLインジェクション防止

#### NG: SQLの文字列結合

```go
// 危険 - SQLインジェクション脆弱性
query := fmt.Sprintf("SELECT * FROM foods WHERE name = '%s'", userInput)
```

#### OK: パラメータ化クエリ

```go
// 安全 - パラメータ化クエリ
row := db.QueryRowContext(ctx,
    "SELECT * FROM foods WHERE name = $1",
    userInput,
)
```

#### 確認ステップ

- [ ] すべてのデータベースクエリがパラメータ化されている
- [ ] SQLで文字列連結がない

### 4. Gemini HTTP API連携のセキュリティ

Gemini APIとの連携時にはHTTPリクエストのセキュリティを確保すること。

#### OK: サーバーサイドでのAPI呼び出し

```go
// APIキーはサーバーサイドのみで使用
func NewGeminiClient(apiKey string) *GeminiClient {
    return &GeminiClient{
        httpClient: &http.Client{
            Timeout: 30 * time.Second, // タイムアウト必須
        },
        apiKey: apiKey,
    }
}
```

#### 確認ステップ

- [ ] APIキーがサーバーサイド（Go）のみで使用されている
- [ ] HTTPリクエストにタイムアウトが設定されている
- [ ] レスポンスのバリデーションが実装されている
- [ ] APIキーがiOSアプリに埋め込まれていない

### 5. APIレスポンスの安全な処理

#### レスポンスバリデーション

```go
// Gemini APIレスポンスのバリデーション
func parseGeminiResponse(body []byte) (*AnalysisResult, error) {
    var result AnalysisResult
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("invalid response format: %w", err)
    }

    // レスポンス内容のバリデーション
    if result.Calories < 0 || result.Calories > 10000 {
        return nil, fmt.Errorf("unreasonable calorie value: %d", result.Calories)
    }

    return &result, nil
}
```

#### 確認ステップ

- [ ] 外部APIレスポンスがバリデートされている
- [ ] 不正なデータがデータベースに保存されない
- [ ] エラーレスポンスが適切に処理されている

### 6. 機密データの露出

#### ロギング

```go
// NG: 機密データをログ
log.Printf("API Key: %s", apiKey)

// OK: 機密データを編集
log.Printf("API call made for userId: %s", userId)
```

#### エラーメッセージ

```go
// NG: 内部詳細を露出
if err != nil {
    return fmt.Errorf("database error: %v, query: %s", err, query)
}

// OK: 汎用エラーメッセージ
if err != nil {
    log.Printf("internal error: %v", err)
    return fmt.Errorf("an error occurred, please try again")
}
```

#### 確認ステップ

- [ ] ログにパスワード、トークン、シークレットがない
- [ ] ユーザー向けエラーメッセージが汎用的
- [ ] 詳細エラーはサーバーログのみ
- [ ] スタックトレースがユーザーに露出していない

### 7. 依存関係セキュリティ

#### 定期的な更新

```bash
# Go依存関係の脆弱性チェック
go list -m all
govulncheck ./...

# Go依存関係の更新
go get -u ./...
go mod tidy
```

#### 確認ステップ

- [ ] 依存関係が最新
- [ ] 既知の脆弱性がない
- [ ] go.sumがコミットされている
- [ ] SPMのPackage.resolvedがコミットされている

## デプロイ前セキュリティチェックリスト

本番デプロイ前に必ず確認:

- [ ] シークレット: ハードコードされたシークレットがない、すべて環境変数にある
- [ ] 入力バリデーション: すべてのユーザー入力がバリデートされている
- [ ] SQLインジェクション: すべてのクエリがパラメータ化されている
- [ ] Gemini API: APIキーがサーバーサイドのみで使用、タイムアウト設定済み
- [ ] レスポンスバリデーション: 外部APIレスポンスが検証されている
- [ ] HTTPS: 本番環境で強制
- [ ] エラーハンドリング: エラーに機密データがない
- [ ] ロギング: 機密データがログされていない
- [ ] 依存関係: 最新で脆弱性がない
- [ ] ファイルアップロード: バリデートされている（サイズ、タイプ）

## リソース

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/doc/security/best-practices)

---

セキュリティはオプションではありません。1つの脆弱性がプラットフォーム全体を危険にさらす可能性があります。迷ったら、慎重側に倒してください。
