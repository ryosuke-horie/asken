# ADR-003: 自炊メニュー機能のアーキテクチャ設計

## コンテキスト

ウチコミは現在、食事画像やテキストから栄養素を自動推定し記録する機能を持つ。しかし「何を食べるべきか」の提案機能がなく、ユーザーは自分で食事内容を決める必要がある。

自炊メニュー機能は、冷蔵庫の食材を管理し、ユーザーの身体データ・食事履歴・栄養目標に基づいて最適なメニューと調理方法をサジェストする機能である。

### 解決すべき課題

- 食材の在庫把握が困難（何がどれくらい残っているか分からない）
- 栄養目標に沿った食事の献立を考える負担が大きい
- 食材のロス（消費期限切れ）を防ぎたい

## 決定

### 機能構成

| 機能 | 概要 |
|:---|:---|
| 食材管理 | レシート撮影・手動入力で食材を登録し、在庫を管理する |
| メニューサジェスト | 在庫食材・栄養目標・食事履歴・体重推移を考慮してメニューを提案 |
| レシピ提示 | 提案メニューの具体的な調理手順を表示 |
| 食事記録連動 | サジェスト採用時に食事記録を作成し、使用食材を自動控除 |

### データモデル（新規Firestoreコレクション）

```
users/{userId}/
  ├── ingredients/{ingredientId}    # 食材在庫
  └── menuSuggestions/{suggestionId} # メニューサジェスト履歴
```

既存コレクション（`analysisRequests`, `nutritionGoal`, `weightRecords`, `weightGoal`）は変更なし。食事記録連動時は既存の `analysisRequests` に新しい `inputType: "suggestion"` として記録する。

### Gemini API 拡張

既存の2段階パイプライン（画像分類またはテキスト解析→栄養計算）に加え、3つの新しいプロンプトを追加する。

| プロンプト | 入力 | 出力 |
|:---|:---|:---|
| レシート解析 | レシート画像 | 食材リスト（名前、数量、単位、カテゴリ） |
| メニュー提案 | 在庫食材 + 栄養目標 + 食事履歴 + 体重推移 | メニュー候補（タイトル、食材、推定栄養素） |
| レシピ生成 | メニュータイトル + 使用食材 | 調理手順 |

### API設計方針

既存のルーティング構造（Handler層→Repository層、Gemini連携等が必要な場合はService層を挟む）を踏襲する。

| カテゴリ | エンドポイント | 概要 |
|:---|:---|:---|
| 食材管理 | `POST /api/ingredients/scan-receipt` | レシート画像から食材抽出 |
| 食材管理 | `GET /api/ingredients` | 食材一覧取得 |
| 食材管理 | `POST /api/ingredients` | 食材手動追加 |
| 食材管理 | `PUT /api/ingredients/{id}` | 食材更新 |
| 食材管理 | `DELETE /api/ingredients/{id}` | 食材削除 |
| サジェスト | `POST /api/menu/suggest` | メニューサジェスト生成 |
| サジェスト | `GET /api/menu/suggestions` | サジェスト一覧取得 |
| サジェスト | `GET /api/menu/suggestions/{id}` | サジェスト詳細（レシピ含む） |
| サジェスト | `POST /api/menu/suggestions/{id}/accept` | サジェスト採用→食事記録+食材控除 |
| サジェスト | `POST /api/menu/suggestions/{id}/dismiss` | サジェスト却下 |

### iOS アーキテクチャ

既存のMVVM + Repositoryパターンを踏襲し、2つの新規Featureモジュールを追加する。

| モジュール | 責務 |
|:---|:---|
| `Features/Pantry/` | 食材一覧、食材追加・編集、レシート撮影 |
| `Features/CookingSuggestion/` | サジェスト一覧、レシピ詳細、食事記録連携 |

### レート制限

メニューサジェストとレシート解析はGemini APIを呼び出すため、既存のレート制限ミドルウェア（`GeminiRateLimit`）の対象とする。

## 理由

### 食材を独立コレクションにした理由

- 食材は食事記録とは独立したライフサイクルを持つ（購入→消費→補充）
- 食材単位でのCRUD操作が必要
- 消費期限によるフィルタリングやソートが必要（複合インデックスが必要になる）
- `myMenu` とは性質が異なる（`myMenu` は「よく食べるメニューの登録」、`ingredients` は「現在の在庫」）

### Gemini APIを3つに分けた理由

- レシート解析: 画像入力 → 構造化データ出力（既存の食事画像解析と類似の処理パターン）
- メニュー提案: 複数のコンテキストデータを統合した推論（入力が多い）
- レシピ生成: 遅延実行可能（ユーザーが詳細を見た時点で生成）

1回のAPIコールで全てを処理するとプロンプトが肥大化し、出力品質が低下するリスクがある。また、レシピ生成を分離することでサジェスト一覧の応答速度を向上させる。

### `inputType: "suggestion"` を追加する理由

新しいコレクションを作るのではなく、既存の `analysisRequests` を拡張する:

- 食事記録の一覧表示、日次集計、カレンダー表示などの既存機能がそのまま使える
- `inputType` によるフィルタリングで入力元を判別可能
- `nutritionGoal` との整合性チェックや `weightRecords` との相関分析もそのまま機能する

### 代替案と却下理由

| 代替案 | 却下理由 |
|:---|:---|
| 外部レシピDBとの連携 | 外部依存が増加し、コスト・メンテナンスの負担が大きい |
| 食材のバーコードスキャン | MVP段階ではオーバーエンジニアリング |
| リアルタイム在庫同期 | 個人利用のため不要。ローカル状態で十分 |

## 結果

### 導入が必要なもの

バックエンド:
- `internal/handler/ingredient_handler.go` - 食材管理ハンドラ
- `internal/handler/menu_suggestion_handler.go` - サジェストハンドラ
- `internal/repository/ingredient_repository.go` - 食材リポジトリ
- `internal/repository/menu_suggestion_repository.go` - サジェストリポジトリ
- `pkg/gemini/receipt_parser.go` - レシート解析
- `pkg/gemini/menu_suggester.go` - メニュー提案
- `pkg/gemini/recipe_generator.go` - レシピ生成

iOS:
- `Features/Pantry/` - 食材管理画面
- `Features/CookingSuggestion/` - サジェスト画面
- `Core/Models/Ingredient.swift` - 食材モデル
- `Core/Models/MenuSuggestion.swift` - サジェストモデル
- `Core/Repositories/IngredientRepository.swift` - 食材リポジトリ
- `Core/Repositories/MenuSuggestionRepository.swift` - サジェストリポジトリ

Firestore:
- 新規コレクション用の複合インデックス追加
- `firestore.indexes.json` の更新

### 既存機能への影響

- `analysisRequests` に `inputType: "suggestion"` が追加される（`docs/CODEMAPS/data.md` の更新も必要）
- ルーティング設定に新規エンドポイントが追加される
- レート制限ミドルウェアの対象エンドポイントが増加する
- iOS のタブバー/ナビゲーションに食材管理・サジェストへの導線が追加される

### コスト影響

Gemini APIの呼び出し回数が増加する:
- レシート解析: 買い物のたびに1回（週2-3回程度）
- メニュー提案: 1日1-2回
- レシピ生成: メニュー詳細閲覧時（オンデマンド）

追加コスト見積もり（月間、通常利用）:
- 追加トークン: 入力約500K + 出力約200K
- 追加コスト: 無料枠超過時でも約$0.85/月

### 関連ドキュメント

- [サーバレスインフラ設計](./002-serverless-infrastructure.md)
- [自炊メニュー機能設計書](../specs/cooking-menu-feature.md)
- [データモデル](../CODEMAPS/data.md)
