# Git/GitHubワークフロー詳細

## タスク管理ツール

- タスク管理はLinearを使用すること
- 作業は必ずLinearのタスクベースで行うこと

## ブランチ命名規則

- ブランチ名は`edg-{タスクID}`の形式とすること
  - 例: `edg-305`, `edg-421`
  - タスクIDはLinearで採番されたIDを使用する

## コミットメッセージ形式

Conventional Commits形式を使用:

```
<type>: <description>

<optional body>
```

### type一覧

| type | 用途 |
| :--- | :--- |
| feat | 新機能 |
| fix | バグ修正 |
| refactor | リファクタリング |
| docs | ドキュメント |
| test | テスト |
| chore | ビルド、設定等 |
| perf | パフォーマンス改善 |
| ci | CI/CD関連 |

### 例

```bash
git commit -m "feat: 食事画像アップロード機能を追加"
git commit -m "fix: 栄養素計算のバリデーションエラーを修正"
git commit -m "refactor: NutritionDisplayコンポーネントを分割"
```

## プルリクエスト作成時のルール

### PR descriptionの記載

- PR descriptionには必ずLinearのタスクIDを記載すること
  - 形式: `Fixes EDG-{タスクID}`または`Closes EDG-{タスクID}`
  - 例: `Fixes EDG-305`
  - これによりPRがマージされたときにLinearのタスクが自動的に完了状態になる

## 作業フロー例

```bash
# 1. Linearでタスクを確認 (例: EDG-305)

# 2. edg-xxxブランチを作成
git checkout -b edg-305

# 3. 作業を実施

# 4. コミット
git add .
git commit -m "feat: 機能の説明"

# 5. プッシュ
git push -u origin edg-305

# 6. PRを作成（descriptionにFixes EDG-xxxを含める）
gh pr create --title "タイトル" --body "Fixes EDG-305"
```
