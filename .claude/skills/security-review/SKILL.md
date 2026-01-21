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

```typescript
const apiKey = "sk-proj-xxxxx"  // ハードコードされたシークレット
const dbPassword = "password123" // ソースコード内
```

#### OK: 必ずこうする

```typescript
const apiKey = process.env.GEMINI_API_KEY

// シークレットの存在を検証
if (!apiKey) {
  throw new Error('GEMINI_API_KEY not configured')
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

```typescript
// バリデーションスキーマを定義
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

#### ファイルアップロードバリデーション

```typescript
function validateFileUpload(file: File) {
  // サイズチェック（最大5MB）
  const maxSize = 5 * 1024 * 1024
  if (file.size > maxSize) {
    throw new Error('File too large (max 5MB)')
  }

  // タイプチェック
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif']
  if (!allowedTypes.includes(file.type)) {
    throw new Error('Invalid file type')
  }

  // 拡張子チェック
  const allowedExtensions = ['.jpg', '.jpeg', '.png', '.gif']
  const extension = file.name.toLowerCase().match(/\.[^.]+$/)?.[0]
  if (!extension || !allowedExtensions.includes(extension)) {
    throw new Error('Invalid file extension')
  }

  return true
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

```typescript
// 危険 - SQLインジェクション脆弱性
const query = `SELECT * FROM foods WHERE name = '${userInput}'`
```

#### OK: パラメータ化クエリ

```typescript
// 安全 - パラメータ化クエリ
await db.query(
  'SELECT * FROM foods WHERE name = $1',
  [userInput]
)
```

#### 確認ステップ

- [ ] すべてのデータベースクエリがパラメータ化されている
- [ ] SQLで文字列連結がない

### 4. コマンドインジェクション防止（Gemini CLI連携）

シェルコマンドにユーザー入力を直接渡すことは危険です。

#### OK: ファイルパスをサニタイズ

```typescript
import path from 'path'

function sanitizeFilePath(filePath: string): string {
  // パストラバーサル防止
  const normalized = path.normalize(filePath)
  if (normalized.includes('..')) {
    throw new Error('Invalid file path')
  }

  // 許可されたディレクトリ内のみ
  const allowedDir = '/uploads/'
  if (!normalized.startsWith(allowedDir)) {
    throw new Error('File path not allowed')
  }

  return normalized
}
```

#### 確認ステップ

- [ ] ファイルパスがサニタイズされている
- [ ] パストラバーサル攻撃が防止されている
- [ ] シェルコマンドにユーザー入力が直接渡されていない
- [ ] execFileを使用（シェルインジェクション防止）

### 5. XSS防止

#### HTMLのサニタイズ

```typescript
import DOMPurify from 'isomorphic-dompurify'

// ユーザー提供のHTMLは常にDOMPurifyでサニタイズ
function sanitizeUserContent(html: string): string {
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p'],
    ALLOWED_ATTR: []
  })
}
```

#### 確認ステップ

- [ ] ユーザー提供のHTMLがサニタイズされている
- [ ] Reactの組み込みXSS保護が使用されている
- [ ] 生のHTML挿入は避ける

### 6. 機密データの露出

#### ロギング

```typescript
// NG: 機密データをログ
console.log('API Key:', apiKey)

// OK: 機密データを編集
console.log('API call made for userId:', userId)
```

#### エラーメッセージ

```typescript
// NG: 内部詳細を露出
catch (error) {
  return { error: error.message, stack: error.stack }
}

// OK: 汎用エラーメッセージ
catch (error) {
  console.error('Internal error:', error)
  return { error: 'An error occurred. Please try again.' }
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
# 脆弱性をチェック
npm audit

# 自動修正可能な問題を修正
npm audit fix

# 依存関係を更新
npm update
```

#### 確認ステップ

- [ ] 依存関係が最新
- [ ] 既知の脆弱性がない（npm auditがクリーン）
- [ ] package-lock.jsonがコミットされている

## デプロイ前セキュリティチェックリスト

本番デプロイ前に必ず確認:

- [ ] シークレット: ハードコードされたシークレットがない、すべて環境変数にある
- [ ] 入力バリデーション: すべてのユーザー入力がバリデートされている
- [ ] SQLインジェクション: すべてのクエリがパラメータ化されている
- [ ] コマンドインジェクション: ファイルパスがサニタイズされている
- [ ] XSS: ユーザーコンテンツがサニタイズされている
- [ ] HTTPS: 本番環境で強制
- [ ] エラーハンドリング: エラーに機密データがない
- [ ] ロギング: 機密データがログされていない
- [ ] 依存関係: 最新で脆弱性がない
- [ ] ファイルアップロード: バリデートされている（サイズ、タイプ）

## リソース

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Next.js Security](https://nextjs.org/docs/security)

---

セキュリティはオプションではありません。1つの脆弱性がプラットフォーム全体を危険にさらす可能性があります。迷ったら、慎重側に倒してください。
