---
name: tdd-workflow
description: 新機能の実装、バグ修正、リファクタリング時に使用するスキル。テスト駆動開発を強制し、80%以上のカバレッジを確保。ユニット、統合、E2Eテストを含む。
---

# テスト駆動開発ワークフロー

このスキルはすべてのコード開発がTDD原則に従い、包括的なテストカバレッジを確保することを保証します。

## 使用タイミング

- 新機能や機能の実装時
- バグや問題の修正時
- 既存コードのリファクタリング時
- APIエンドポイントの追加時
- 新しいコンポーネントの作成時

## 基本原則

### 1. コードの前にテスト

常にテストを先に書き、その後テストを通すコードを実装する。

### 2. カバレッジ要件

- 最低80%カバレッジ（ユニット + 統合 + E2E）
- すべてのエッジケースをカバー
- エラーシナリオをテスト
- 境界条件を検証

### 3. テストの種類

#### ユニットテスト

- 個別の関数とユーティリティ
- コンポーネントロジック
- 純粋関数
- ヘルパーとユーティリティ

#### 統合テスト

- APIエンドポイント
- データベース操作
- サービス間の相互作用
- 外部API呼び出し

#### E2Eテスト

- 重要なユーザーフロー
- 完全なワークフロー
- ブラウザ自動化
- UI操作

## TDDワークフローステップ

### ステップ1: ユーザージャーニーを書く

```
[役割]として、[アクション]をしたい、[メリット]のために

例:
ユーザーとして、食事画像をアップロードしたい、
カロリーと栄養素を自動計算してもらうために。
```

### ステップ2: テストケースを生成

各ユーザージャーニーに対して包括的なテストケースを作成:

```typescript
describe('画像アップロード', () => {
  it('有効な画像でカロリーを返すべき', async () => {
    // テスト実装
  })

  it('無効なファイル形式でエラーを返すべき', async () => {
    // エッジケースをテスト
  })

  it('API障害時にフォールバックすべき', async () => {
    // フォールバック動作をテスト
  })
})
```

### ステップ3: テストを実行（失敗すべき）

```bash
npm test
# テストは失敗すべき - まだ実装していない
```

### ステップ4: コードを実装

テストを通す最小限のコードを書く:

```typescript
// テストに導かれた実装
export async function analyzeImage(imagePath: string) {
  // ここに実装
}
```

### ステップ5: 再度テストを実行

```bash
npm test
# テストが通るはず
```

### ステップ6: リファクタリング

テストをグリーンに保ちながらコード品質を改善:

- 重複を削除
- 命名を改善
- パフォーマンスを最適化
- 可読性を向上

### ステップ7: カバレッジを確認

```bash
npm test -- --coverage
# 80%以上のカバレッジを確認
```

## テストパターン

### ユニットテストパターン（Jest）

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from './Button'

describe('Buttonコンポーネント', () => {
  it('正しいテキストでレンダリングすべき', () => {
    render(<Button>クリック</Button>)
    expect(screen.getByText('クリック')).toBeInTheDocument()
  })

  it('クリック時にonClickを呼び出すべき', () => {
    const handleClick = jest.fn()
    render(<Button onClick={handleClick}>クリック</Button>)

    fireEvent.click(screen.getByRole('button'))

    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('disabledプロパティがtrueの場合は無効化されるべき', () => {
    render(<Button disabled>クリック</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })
})
```

### API統合テストパターン

```typescript
describe('GET /api/meals', () => {
  it('食事データを正常に返すべき', async () => {
    const response = await fetch('http://localhost:3000/api/meals')
    const data = await response.json()

    expect(response.status).toBe(200)
    expect(data.success).toBe(true)
    expect(Array.isArray(data.data)).toBe(true)
  })

  it('クエリパラメータをバリデートすべき', async () => {
    const response = await fetch('http://localhost:3000/api/meals?date=invalid')

    expect(response.status).toBe(400)
  })
})
```

## 外部サービスのモック

### Gemini APIモック

```typescript
jest.mock('@/lib/gemini', () => ({
  analyzeFood: jest.fn(() => Promise.resolve({
    foods: [{ name: 'ご飯', calories: 250 }],
    success: true
  }))
}))
```

### Firestoreモック

```typescript
jest.mock('@/lib/firestore', () => ({
  getDoc: jest.fn(() => Promise.resolve({
    exists: true,
    data: () => ({ id: '1', name: 'テストデータ' })
  }))
}))
```

## テストカバレッジ確認

### カバレッジレポートを実行

```bash
npm test -- --coverage
```

### カバレッジしきい値

```json
{
  "jest": {
    "coverageThreshold": {
      "global": {
        "branches": 80,
        "functions": 80,
        "lines": 80,
        "statements": 80
      }
    }
  }
}
```

## よくあるテストの間違い

### NG: 実装詳細をテスト

```typescript
// 内部状態をテストしない
expect(component.state.count).toBe(5)
```

### OK: ユーザーに見える振る舞いをテスト

```typescript
// ユーザーが見るものをテスト
expect(screen.getByText('カウント: 5')).toBeInTheDocument()
```

### NG: テスト分離なし

```typescript
// テストが相互依存
test('ユーザーを作成', () => { /* ... */ })
test('同じユーザーを更新', () => { /* 前のテストに依存 */ })
```

### OK: 独立したテスト

```typescript
// 各テストが自分のデータをセットアップ
test('ユーザーを作成', () => {
  const user = createTestUser()
  // テストロジック
})

test('ユーザーを更新', () => {
  const user = createTestUser()
  // 更新ロジック
})
```

## 継続的テスト

### 開発中のウォッチモード

```bash
npm test -- --watch
# ファイル変更時にテストが自動実行
```

### プリコミットフック

```bash
# 各コミット前に実行
npm test && npm run lint
```

## ベストプラクティス

1. テストを先に書く - 常にTDD
2. 1テスト1アサート - 単一の振る舞いに集中
3. 説明的なテスト名 - 何をテストしているか説明
4. Arrange-Act-Assert - 明確なテスト構造
5. 外部依存をモック - ユニットテストを分離
6. エッジケースをテスト - null、undefined、空、大きな値
7. エラーパスをテスト - ハッピーパスだけでなく
8. テストを高速に - ユニットテストは各50ms未満
9. テスト後にクリーンアップ - 副作用なし
10. カバレッジレポートをレビュー - ギャップを特定

## 成功指標

- 80%以上のコードカバレッジ達成
- すべてのテストがパス（グリーン）
- スキップまたは無効化されたテストがない
- 高速なテスト実行（ユニットテストは30秒未満）
- テストが本番前にバグをキャッチ

---

テストはオプションではありません。自信を持ったリファクタリング、迅速な開発、本番の信頼性を可能にするセーフティネットです。
