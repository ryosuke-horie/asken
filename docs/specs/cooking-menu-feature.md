# 自炊メニュー機能 設計書

## 概要

冷蔵庫の食材を管理し、ユーザーの栄養目標・食事履歴・体重推移に基づいて最適な自炊メニューと調理方法をサジェストする機能。

関連:
- Issue: UTK-22
- ADR: [ADR-003](../adr/003-cooking-menu-architecture.md)

---

## 機能フロー全体像

```
[食材登録] → [食材管理] → [メニューサジェスト] → [レシピ確認] → [食事記録連動]
    │              │               │                                    │
    ├─ レシート撮影  ├─ 在庫確認      ├─ 栄養目標考慮                      ├─ 食事記録作成
    ├─ 手動入力     ├─ 数量編集      ├─ 食事履歴考慮                      └─ 食材在庫控除
    └─ 食事記録控除  └─ 削除         └─ 体重推移考慮
```

---

## 1. データモデル

### 1.1 ingredients コレクション

`users/{userId}/ingredients/{ingredientId}`

| フィールド | 型 | 必須 | 説明 |
|:---|:---|:---:|:---|
| id | string | Yes | 食材ID |
| name | string | Yes | 食材名（例: 鶏むね肉） |
| category | string | Yes | カテゴリ |
| quantity | number | Yes | 数量 |
| unit | string | Yes | 単位 |
| purchaseDate | timestamp | No | 購入日 |
| expiryDate | timestamp | No | 消費期限 |
| source | string | Yes | 入力元 (receipt / manual) |
| createdAt | timestamp | Yes | 作成日時 |
| updatedAt | timestamp | Yes | 更新日時 |

### カテゴリ一覧

| 値 | 表示名 |
|:---|:---|
| `meat` | 肉類 |
| `fish` | 魚介類 |
| `vegetable` | 野菜 |
| `fruit` | 果物 |
| `dairy` | 乳製品 |
| `grain` | 穀物 |
| `seasoning` | 調味料 |
| `beverage` | 飲料 |
| `other` | その他 |

### 単位

既存の `MeasurementUnit`（Go/iOS共通）を再利用する:
g, ml, 杯, 人前, 個, 枚, 本, 切れ, パック, 袋, 束, 丁, 缶, 合, 玉, 粒, 大さじ, 小さじ

### 1.2 menuSuggestions コレクション

`users/{userId}/menuSuggestions/{suggestionId}`

| フィールド | 型 | 必須 | 説明 |
|:---|:---|:---:|:---|
| id | string | Yes | サジェストID |
| title | string | Yes | メニュー名 |
| description | string | Yes | メニュー概要（1-2文） |
| reason | string | Yes | 提案理由（栄養バランス等） |
| ingredientsUsed | array | Yes | 使用食材リスト（下記参照） |
| recipe | string | No | 調理手順（遅延生成） |
| estimatedNutrition | map | Yes | 推定栄養素 |
| mealType | string | Yes | 推奨食事タイプ |
| status | string | Yes | 状態 |
| createdAt | timestamp | Yes | 作成日時 |
| updatedAt | timestamp | Yes | 更新日時 |

### ingredientsUsed 配列要素

```json
{
  "ingredientId": "食材ドキュメントID",
  "name": "鶏むね肉",
  "quantity": 200,
  "unit": "g"
}
```

### estimatedNutrition 構造

```json
{
  "calories": 450,
  "protein": 35.0,
  "fat": 12.0,
  "carbohydrates": 45.0
}
```

### status の状態遷移

```
suggested → accepted（食事記録作成、食材控除）
suggested → dismissed（却下、データ保持）
```

### 1.3 Firestore インデックス（追加分）

| コレクション | フィールド | 用途 |
|:---|:---|:---|
| ingredients | category ASC, name ASC | カテゴリ別食材一覧 |
| ingredients | expiryDate ASC | 消費期限順ソート |
| menuSuggestions | status ASC, createdAt DESC | ステータス別サジェスト一覧 |

### 1.4 analysisRequests の拡張

`inputType` に `"suggestion"` を追加する。

サジェスト採用時に作成される `analysisRequests` ドキュメント:

| フィールド | 値 |
|:---|:---|
| inputType | `"suggestion"` |
| status | `"completed"` |
| confirmed | `true` |
| result | サジェストの `estimatedNutrition` から生成 |
| mealType | サジェストの `mealType` |
| mealDate | 採用時の日付 |

---

## 2. API仕様

### 2.1 食材管理

#### POST /api/ingredients/scan-receipt

レシート画像から食材を抽出する。

リクエスト:
- Content-Type: `multipart/form-data`
- `image`: レシート画像（JPEG, PNG, HEIC, 最大10MB）

レスポンス:
```json
{
  "success": true,
  "data": {
    "ingredients": [
      {
        "name": "鶏むね肉",
        "category": "meat",
        "quantity": 500,
        "unit": "g",
        "source": "receipt"
      }
    ]
  }
}
```

この段階では食材はまだ保存されない。クライアントがレスポンスを確認・編集した後、個別に `POST /api/ingredients` で保存する。

#### GET /api/ingredients

食材一覧を取得する。

クエリパラメータ:

| パラメータ | 型 | 必須 | 説明 |
|:---|:---|:---:|:---|
| category | string | No | カテゴリでフィルタ |

レスポンス:
```json
{
  "success": true,
  "data": {
    "ingredients": [
      {
        "id": "xxx",
        "name": "鶏むね肉",
        "category": "meat",
        "quantity": 500,
        "unit": "g",
        "purchaseDate": "2026-02-18T00:00:00Z",
        "expiryDate": "2026-02-22T00:00:00Z",
        "source": "receipt",
        "createdAt": "2026-02-18T10:00:00Z",
        "updatedAt": "2026-02-18T10:00:00Z"
      }
    ]
  }
}
```

#### POST /api/ingredients

食材を手動追加する。

リクエスト:
```json
{
  "name": "鶏むね肉",
  "category": "meat",
  "quantity": 500,
  "unit": "g",
  "purchaseDate": "2026-02-18",
  "expiryDate": "2026-02-22",
  "source": "manual"
}
```

バリデーション:
- `name`: 1-100文字
- `category`: 上記カテゴリ一覧のいずれか
- `quantity`: 0より大きい数値
- `unit`: サポート対象の単位
- `source`: `receipt` または `manual`

#### PUT /api/ingredients/{id}

食材を更新する（数量変更、消費期限修正など）。

リクエスト: POST と同じフィールド（部分更新可）。

#### DELETE /api/ingredients/{id}

食材を削除する。

### 2.2 メニューサジェスト

#### POST /api/menu/suggest

メニューサジェストを生成する。

リクエスト:
```json
{
  "mealType": "dinner",
  "count": 3
}
```

| パラメータ | 型 | 必須 | 説明 |
|:---|:---|:---:|:---|
| mealType | string | Yes | 食事タイプ |
| count | number | No | 提案数（デフォルト: 3、最大: 5） |

処理フロー:
1. ユーザーの食材一覧を取得
2. 栄養目標（`nutritionGoal/current`）を取得
3. 直近7日間の食事履歴（`analysisRequests`）を取得
4. 直近30日間の体重推移（`weightRecords`）を取得
5. 上記をコンテキストとしてGemini APIでメニュー提案を生成
6. 結果を `menuSuggestions` に保存して返す

レスポンス:
```json
{
  "success": true,
  "data": {
    "suggestions": [
      {
        "id": "xxx",
        "title": "鶏むね肉のヘルシー照り焼き定食",
        "description": "高タンパク・低脂質の鶏むね肉を使った照り焼きと野菜の副菜",
        "reason": "本日のタンパク質摂取量が目標の60%のため、高タンパクメニューを提案",
        "ingredientsUsed": [...],
        "estimatedNutrition": {
          "calories": 520,
          "protein": 42.0,
          "fat": 10.0,
          "carbohydrates": 55.0
        },
        "mealType": "dinner",
        "status": "suggested",
        "createdAt": "2026-02-19T10:00:00Z"
      }
    ]
  }
}
```

#### GET /api/menu/suggestions

サジェスト一覧を取得する。

クエリパラメータ:

| パラメータ | 型 | 必須 | 説明 |
|:---|:---|:---:|:---|
| status | string | No | ステータスフィルタ（デフォルト: suggested） |
| limit | number | No | 取得件数（デフォルト: 10） |

#### GET /api/menu/suggestions/{id}

サジェスト詳細を取得する。`recipe` フィールドが未生成の場合、このタイミングでGemini APIを呼び出してレシピを生成し保存する（遅延生成）。

レスポンス（レシピ含む）:
```json
{
  "success": true,
  "data": {
    "id": "xxx",
    "title": "鶏むね肉のヘルシー照り焼き定食",
    "description": "...",
    "reason": "...",
    "ingredientsUsed": [...],
    "recipe": "1. 鶏むね肉を一口大に切る\n2. ...",
    "estimatedNutrition": {...},
    "mealType": "dinner",
    "status": "suggested"
  }
}
```

#### POST /api/menu/suggestions/{id}/accept

サジェストを採用する。以下を同時実行:

1. `menuSuggestions` のステータスを `accepted` に更新
2. `analysisRequests` に `inputType: "suggestion"` で食事記録を作成
3. `ingredients` から使用食材の数量を控除（控除後に0以下なら食材を削除）

レスポンス:
```json
{
  "success": true,
  "data": {
    "analysisRequestId": "xxx",
    "deductedIngredients": [
      {
        "ingredientId": "yyy",
        "name": "鶏むね肉",
        "deducted": 200,
        "remaining": 300
      }
    ]
  }
}
```

#### DELETE /api/menu/suggestions/{id}

サジェストを却下する（ステータスを `dismissed` に更新）。

---

## 3. Gemini API プロンプト設計

### 3.1 レシート解析プロンプト

入力: レシート画像

出力スキーマ:
```json
{
  "type": "object",
  "properties": {
    "items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "category": { "type": "string", "enum": ["meat", "fish", "vegetable", "fruit", "dairy", "grain", "seasoning", "beverage", "other"] },
          "quantity": { "type": "number" },
          "unit": { "type": "string" }
        },
        "required": ["name", "category", "quantity", "unit"]
      }
    }
  }
}
```

プロンプトの方針:
- レシートに記載された商品名から食材名を推定する
- パッケージ商品（例: 「若鶏もも 2P」）は食材名と数量に分解する
- 食材以外の商品（日用品、ポイント関連等）は除外する
- カテゴリは食材の種類から自動判別する
- 数量が不明な場合は一般的な1パックの量を推定する

### 3.2 メニュー提案プロンプト

入力コンテキスト:
- 利用可能な食材リスト（名前、数量、単位、消費期限）
- 栄養目標（カロリー、PFCバランス）
- 直近7日間の食事記録（メニュー名、栄養素）
- 直近30日間の体重推移
- 要求: 食事タイプ、提案数

プロンプトの方針:
- 消費期限が近い食材を優先的に使用する
- 直近の食事と被らないメニューを提案する
- 栄養目標の不足分を補うメニューを優先する
- 体重推移が増加傾向なら低カロリーメニューを、減少傾向なら適度なカロリーのメニューを提案
- 各提案に「なぜこのメニューを提案したか」の理由を付与する

出力スキーマ:
```json
{
  "type": "object",
  "properties": {
    "suggestions": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title": { "type": "string" },
          "description": { "type": "string" },
          "reason": { "type": "string" },
          "ingredients": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "name": { "type": "string" },
                "quantity": { "type": "number" },
                "unit": { "type": "string" }
              }
            }
          },
          "estimatedNutrition": {
            "type": "object",
            "properties": {
              "calories": { "type": "number" },
              "protein": { "type": "number" },
              "fat": { "type": "number" },
              "carbohydrates": { "type": "number" }
            }
          }
        }
      }
    }
  }
}
```

### 3.3 レシピ生成プロンプト

入力:
- メニュータイトル
- 使用食材リスト（名前、数量、単位）

プロンプトの方針:
- 手順は番号付きリストで記述
- 各手順は簡潔かつ具体的に
- 火加減や時間の目安を含める
- 調味料の分量も明記する

出力: プレーンテキスト（マークダウン形式の手順）

---

## 4. iOS 画面設計

### 4.1 食材管理（Pantry）

#### PantryListView

- 食材をカテゴリ別にグループ化して表示
- 各食材: 名前、数量+単位、消費期限（期限が近い場合は強調表示）
- アクション: 食材追加（手動 / レシート）、数量クイック編集、スワイプ削除
- 空状態: 食材登録を促すイラスト + ボタン

#### IngredientEditView

- 食材の追加・編集フォーム
- フィールド: 名前、カテゴリ（ピッカー）、数量+単位、購入日、消費期限
- バリデーション: 名前必須、数量は正の数

#### ReceiptScanView

- カメラUIでレシート撮影（既存の `CameraView` を再利用可能）
- 解析結果をリスト表示、各項目を編集可能
- 一括保存ボタン

### 4.2 メニューサジェスト（CookingSuggestion）

#### SuggestionRequestView

- 食事タイプ選択（朝食/昼食/夕食/間食）
- サジェスト生成ボタン
- ローディング中はアニメーション表示

#### SuggestionListView

- サジェスト候補をカード形式で表示
- 各カード: タイトル、概要、推定栄養素（カロリー、PFC）、使用食材数
- タップでレシピ詳細へ遷移

#### RecipeDetailView

- メニュータイトル、提案理由
- 使用食材リスト（在庫の何%を使用するか表示）
- 推定栄養素（目標との比較グラフ）
- 調理手順（遅延読み込み）
- 「このメニューで記録する」ボタン → 食事記録作成 + 食材控除

### 4.3 ナビゲーション

タブバーに「食材」タブを追加するか、既存の食事記録画面からサジェスト機能への導線を設ける。

導線案:
```
タブバー
├── 食事（既存）
│     └── サジェストボタン → SuggestionRequestView
├── 食材（新規）→ PantryListView
├── 体重（既存）
└── 設定（既存）
```

---

## 5. 食事記録連動フロー

### 5.1 サジェスト採用時

```
1. ユーザーが RecipeDetailView で「このメニューで記録する」をタップ
2. POST /api/menu/suggestions/{id}/accept を呼び出し
3. バックエンドで以下をトランザクション実行:
   a. menuSuggestions のステータスを accepted に更新
   b. analysisRequests に新規ドキュメント作成 (inputType: "suggestion")
   c. 各使用食材の ingredients.quantity を控除
   d. 数量が0以下になった食材は削除
4. レスポンスを受けてiOS側の状態を更新
```

### 5.2 食事記録からの食材控除（将来拡張）

通常の食事記録（画像/テキスト入力）時にも、分析結果の食品名と在庫食材を照合して自動控除する機能を将来的に追加可能。この設計書の初期スコープには含めないが、データモデルは拡張可能な設計にしておく。

---

## 6. エラーハンドリング

| シナリオ | 対応 |
|:---|:---|
| レシート画像が食品レシートでない | Gemini の出力が空配列 → 「食品が検出されませんでした」表示 |
| Gemini API タイムアウト | 既存のタイムアウト設定（120秒）でエラー返却 |
| レート制限超過 | 既存のレート制限ミドルウェアで 429 返却 |
| 食材の在庫不足（サジェスト採用時） | 実際の在庫と照合し、不足分を警告表示（採用は可能） |
| サジェスト生成に十分な食材がない | 最低限のメニューを提案 + 「食材が少ないため限定的な提案です」表示 |

---

## 7. テスト方針

| 対象 | テスト方法 |
|:---|:---|
| Repository層（Firestore操作） | Firestoreエミュレータで統合テスト |
| Handler層（API） | テーブル駆動ユニットテスト |
| Gemini連携 | プロンプトのレスポンスをモックしたユニットテスト |
| 食材控除ロジック | ユニットテスト（境界値: 0になるケース、不足ケース） |
| E2E | デプロイ環境でのAPI結合テスト |
| iOS | 手動テスト（ios-testing-policyに従う） |
