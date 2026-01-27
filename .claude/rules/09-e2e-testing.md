# E2Eテスト（Playwright）

## 基本方針

- E2EテストはPlaywrightを使用
- テストファイルは`frontend/e2e/tests/`に配置
- ページオブジェクトパターンを使用してテストの保守性を向上

## ディレクトリ構成

```
frontend/e2e/
├── fixtures/         # テストフィクスチャ
│   └── index.ts      # ページオブジェクトの注入
├── pages/            # ページオブジェクト
│   ├── LoginPage.ts
│   └── HomePage.ts
└── tests/            # テストファイル
    └── auth.spec.ts
```

## 実行コマンド

```bash
# ヘッドレスモードで実行
task frontend:e2e

# UIモードで実行（デバッグ用）
task frontend:e2e:ui

# 特定のテストファイルを実行
cd frontend
npx playwright test e2e/tests/auth.spec.ts

# 特定のテストのみ実行
npx playwright test -g "ログインできるべき"
```

## テストの書き方

### 命名規則

```typescript
test.describe('機能名', () => {
  test.describe('シナリオ', () => {
    test('期待される動作を〜〜すべき', async ({ loginPage }) => {
      // テスト実装
    })
  })
})
```

### フィクスチャの使用

```typescript
import { test, expect } from '../fixtures'

test('ログインできるべき', async ({ loginPage, homePage }) => {
  await loginPage.goto()
  await loginPage.login('test@example.com', 'Pass0123')
  await expect(loginPage.page).toHaveURL('/')
})
```

## ページオブジェクトパターン

### ページオブジェクトの作成

```typescript
import { type Page, type Locator } from '@playwright/test'

export class ExamplePage {
  readonly page: Page
  readonly submitButton: Locator

  constructor(page: Page) {
    this.page = page
    this.submitButton = page.locator('button[type="submit"]')
  }

  async goto() {
    await this.page.goto('/example')
  }

  async submit() {
    await this.submitButton.click()
  }
}
```

### ページオブジェクトの設計原則

| 原則 | 説明 |
|:---|:---|
| 単一責任 | 1つのページオブジェクトは1つのページを担当 |
| カプセル化 | セレクタの詳細はページオブジェクト内に隠蔽 |
| 再利用性 | 共通操作はメソッドとして抽出 |
| 読みやすさ | メソッド名は操作の意図を明確に |

## テストデータ管理

### 既存のdb-seedを活用

テストデータは`task db-seed`で作成されるデータを使用:

```typescript
// テストユーザー
const TEST_USER = {
  email: 'test@example.com',
  password: 'Pass0123',
}
```

### テスト前の状態リセット

```typescript
test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.evaluate(() => {
    localStorage.removeItem('asken_auth_token')
    localStorage.removeItem('asken_user')
  })
})
```

## デバッグ方法

### スクリーンショット

テスト失敗時は自動的にスクリーンショットが保存される:
- 保存場所: `frontend/test-results/`

### UIモードでのデバッグ

```bash
npm run e2e:ui
```

UIモードでは:
- ステップごとの実行
- タイムライン表示
- DOM インスペクション
- ネットワークリクエスト確認

### 特定のテストをデバッグ

```bash
# デバッグモードで実行
npx playwright test --debug e2e/tests/auth.spec.ts
```

## ベストプラクティス

### ロケーター選択の優先順位

1. **role**ベース（推奨）: `page.getByRole('button', { name: 'ログイン' })`
2. **テキスト**ベース: `page.getByText('ログイン')`
3. **ラベル**ベース: `page.getByLabel('メールアドレス')`
4. **ID**ベース: `page.locator('#email')`
5. **CSSセレクタ**（最終手段）: `page.locator('.submit-button')`

### 待機処理

```typescript
// 良い例: 明示的な待機
await expect(page).toHaveURL('/')
await expect(element).toBeVisible()

// 悪い例: 固定待機（使用禁止）
await page.waitForTimeout(1000)
```

### アサーション

```typescript
// URLの確認
await expect(page).toHaveURL('/dashboard')

// 要素の可視性
await expect(element).toBeVisible()
await expect(element).toBeHidden()

// テキスト内容
await expect(element).toContainText('成功')

// 要素の状態
await expect(button).toBeEnabled()
await expect(input).toHaveValue('test@example.com')
```

## 禁止事項

- `page.waitForTimeout()` の使用（不安定なテストの原因）
- テスト間の依存関係（各テストは独立して実行可能であるべき）
- 本番環境に対するE2Eテスト実行
- ハードコードされた絶対パス

## 新しいテストの追加手順

1. 必要に応じてページオブジェクトを作成/更新（`e2e/pages/`）
2. フィクスチャに登録（`e2e/fixtures/index.ts`）
3. テストファイルを作成（`e2e/tests/`）
4. ローカルで動作確認（`npm run e2e:ui`）
