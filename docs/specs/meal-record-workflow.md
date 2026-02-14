# 食事記録ワークフロー仕様書

## 概要

食事記録機能は、ユーザーが食事の画像またはテキストを入力し、Gemini APIを使って栄養素を自動推定・保存するワークフローである。

## ワークフロー全体像

```
[1. 入力] → [2. 分類] → [3. 栄養素計算] → [4. 結果表示] → [5. 編集・保存] → [6. 再計算（条件付き）]
```

---

## 1. 入力フェーズ

### 1.1 画像入力

| 項目 | 仕様 |
|:---|:---|
| エンドポイント | `POST /api/analyze` (multipart/form-data) |
| 最大ファイルサイズ | 10MB |
| 対応形式 | JPEG, PNG, HEIC |
| 必須パラメータ | `image`, `meal_type` |
| 任意パラメータ | `meal_date` (YYYY-MM-DD, デフォルト: 今日) |

### 1.2 テキスト入力

| 項目 | 仕様 |
|:---|:---|
| エンドポイント | `POST /api/analyze` (application/json) |
| リクエストボディ上限 | 4KB |
| テキスト最大長 | 1000文字 |
| 必須パラメータ | `text`, `meal_type` |
| 任意パラメータ | `meal_date` (YYYY-MM-DD, デフォルト: 今日) |

### 1.3 meal_type の有効値

| 値 | 表示名 |
|:---|:---|
| `breakfast` | 朝食 |
| `lunch` | 昼食 |
| `dinner` | 夕食 |
| `snack` | 間食 |

---

## 2. 分類フェーズ (Classifier / TextParser)

### 2.1 処理内容

- 画像入力: `Classifier.ClassifyFoodsFromData` で料理名と推定量を抽出
- テキスト入力: `TextParser.ParseTextToFoods` でテキストから料理名と推定量を抽出

### 2.2 Gemini API responseSchema による出力制約

Gemini APIの `responseSchema` を使用して、出力フォーマットをAPIレベルで強制する。

#### スキーマ定義

```json
{
  "type": "ARRAY",
  "items": {
    "type": "OBJECT",
    "properties": {
      "name": { "type": "STRING" },
      "quantity_value": { "type": "NUMBER" },
      "quantity_unit": {
        "type": "STRING",
        "enum": ["g", "ml", "杯", "人前", "個", "枚", "本", "切れ",
                 "食", "皿", "膳", "丁", "束", "袋", "缶", "合", "玉", "粒"]
      }
    },
    "required": ["name", "quantity_value", "quantity_unit"]
  }
}
```

### 2.3 サポート単位一覧

| 単位 | 用途例 |
|:---|:---|
| `g` | 重量（ご飯 200g, 鶏肉 150g） |
| `ml` | 容量（味噌汁 200ml, 牛乳 200ml） |
| `杯` | 飲食物（ラーメン 1杯, コーヒー 1杯） |
| `人前` | 料理（パスタ 1人前） |
| `個` | 個体（おにぎり 2個, りんご 1個） |
| `枚` | 薄いもの（食パン 2枚, 餃子の皮 10枚） |
| `本` | 細長いもの（バナナ 1本, ソーセージ 2本） |
| `切れ` | 切り分けたもの（刺身 8切れ） |
| `食` | 1食分（定食 1食） |
| `皿` | 皿盛り（カレー 1皿, サラダ 1皿） |
| `膳` | ご飯（ご飯 1膳） |
| `丁` | 豆腐（豆腐 1丁） |
| `束` | 束ねるもの（そうめん 1束） |
| `袋` | 袋入り（ポテトチップス 1袋） |
| `缶` | 缶入り（ビール 1缶） |
| `合` | 米（米 2合） |
| `玉` | 丸いもの（レタス 1玉, うどん 1玉） |
| `粒` | 粒状のもの（薬 1粒, 大豆 10粒） |

### 2.4 estimated_amount の構築

Gemini APIの `quantity_value` (数値) と `quantity_unit` (単位) から `estimated_amount` 文字列を構築する。

- 整数値の場合: `"{整数}{単位}"` (例: `1杯`, `200g`)
- 小数値の場合: `"{小数点1位}{単位}"` (例: `1.5杯`, `0.5個`)

### 2.5 出力

`FoodItem` 構造体の配列:
```go
type FoodItem struct {
    Name            string `json:"name"`
    EstimatedAmount string `json:"estimated_amount"`
}
```

---

## 3. 栄養素計算フェーズ (NutritionCalculator)

### 3.1 処理内容

分類フェーズの `FoodItem` リストを入力として、Gemini APIで栄養素を推定する。

### 3.2 responseSchema

分類フェーズと同様のスキーマに加え、栄養素フィールドを含む:

```json
{
  "type": "ARRAY",
  "items": {
    "type": "OBJECT",
    "properties": {
      "name": { "type": "STRING" },
      "quantity_value": { "type": "NUMBER" },
      "quantity_unit": { "type": "STRING", "enum": [...] },
      "calories_kcal": { "type": "NUMBER" },
      "protein_g": { "type": "NUMBER" },
      "fat_g": { "type": "NUMBER" },
      "carbohydrates_g": { "type": "NUMBER" }
    },
    "required": ["name", "quantity_value", "quantity_unit",
                  "calories_kcal", "protein_g", "fat_g", "carbohydrates_g"]
  }
}
```

### 3.3 出力

`NutritionInfo` 構造体の配列:
```go
type NutritionInfo struct {
    Name            string  `json:"name"`
    EstimatedAmount string  `json:"estimated_amount"`
    Calories        float64 `json:"calories_kcal"`
    Protein         float64 `json:"protein_g"`
    Fat             float64 `json:"fat_g"`
    Carbohydrates   float64 `json:"carbohydrates_g"`
}
```

---

## 4. 結果表示フェーズ (iOS)

### 4.1 FoodEditItem の初期化

`NutritionInfo` から `FoodEditItem` を生成する際:

```swift
quantityValue = QuantityParser.parseValue(quantity) ?? quantity
quantityUnit = QuantityParser.parseUnit(quantity)
```

- `estimated_amount` はバックエンドで `{数値}{単位}` フォーマットが保証されている
- `QuantityParser` は18種の単位を全てパース可能
- パース失敗時（レガシーデータ等）: `quantityValue` = 元の文字列, `quantityUnit` = nil

### 4.2 iOS QuantityParser がサポートするフォーマット

| パターン | 例 | パース結果 |
|:---|:---|:---|
| `{数値}g` | `150g` | value=150, unit="g" |
| `{数値}G` | `150G` | value=150, unit="g" |
| `{数値}グラム` | `150グラム` | value=150, unit="g" |
| `{数値}ml` | `200ml` | value=200, unit="ml" |
| `{数値}ML` | `200ML` | value=200, unit="ml" |
| `{数値}mL` | `200mL` | value=200, unit="ml" |
| `{数値}ミリリットル` | `200ミリリットル` | value=200, unit="ml" |
| `{数値}{日本語単位}` | `1杯` | value=1, unit="杯" |

---

## 5. 編集・保存フェーズ

### 5.1 canSave バリデーション (iOS)

保存ボタンの有効/無効を制御する条件:

```swift
var canSave: Bool {
    !foods.isEmpty && foods.allSatisfy { food in
        !food.name.isEmpty &&
            !food.quantityValue.isEmpty
    }
}
```

- 食品リストが1件以上
- 全ての食品の名前が非空
- 全ての食品の数量値が非空

### 5.2 品目の追加・削除

| 操作 | 処理 |
|:---|:---|
| 品目追加 | `addFood()` で空の `FoodEditItem` を追加 |
| 品目削除 | `removeFood()` で指定品目を配列から除去 |

### 5.3 保存 API

| 項目 | 仕様 |
|:---|:---|
| エンドポイント | `PUT /api/history/{id}` |
| リクエスト形式 | `UpdateHistoryRequest` |

#### UpdateHistoryRequest バリデーション

| 項目 | ルール |
|:---|:---|
| `foods` 配列 | 最小1件、最大50件 |
| `foods[].name` | 必須、空文字不可、スペースのみ不可 |
| `foods[].estimated_amount` | 必須、空文字不可、スペースのみ不可 |
| `foods[].calories_kcal` | 0以上 |
| `foods[].protein_g` | 0以上 |
| `foods[].fat_g` | 0以上 |
| `foods[].carbohydrates_g` | 0以上 |

### 5.4 保存時の処理

1. `confirmed: true` に更新（履歴一覧に表示される）
2. 合計値（カロリー、タンパク質、脂質、炭水化物）を再計算
3. レスポンスとして更新後の `HistoryDetail` を返却

---

## 6. 非同期再計算フェーズ

### 6.1 トリガー条件

以下の全てを満たす場合に非同期再計算が実行される:

1. 食材の名前が変更された（`detectNameChanges` で検知）
2. 食材数が変更前と同じ（追加・削除時はスキップ）

### 6.2 処理フロー

1. 現在の食材リストを `FoodItem` に変換
2. Gemini APIで全食材の栄養素を一括再計算
3. 鮮度チェック: 再計算中にユーザーが再保存していないか確認
4. 鮮度チェックOKの場合のみFirestoreに結果を保存

### 6.3 リトライ

| 対象 | リトライ回数 | 間隔 |
|:---|:---|:---|
| Gemini API | 1回（計2回まで） | 2秒 |
| Firestore保存 | 1回（計2回まで） | 1秒 |

### 6.4 リトライ不要なエラー

- `context.Canceled`
- `context.DeadlineExceeded`
- `repository.ErrNotFound`

---

## 7. 削除

| 項目 | 仕様 |
|:---|:---|
| エンドポイント | `DELETE /api/history/{id}` |
| レスポンス | 204 No Content |
| 関連処理 | Cloud Storage上の画像も削除（画像入力の場合のみ） |

個別品目の削除APIは存在しない。品目を削除する場合は、GETで取得→クライアント側で除外→PUTで更新。

---

## 8. データモデル

### Firestore: analysisRequests

| フィールド | 型 | 説明 |
|:---|:---|:---|
| `userID` | string | Firebase UID |
| `status` | string | pending / processing / completed / failed |
| `confirmed` | bool | ユーザーが保存済みか |
| `inputType` | string | image / text / mylist / skipped |
| `mealType` | string | breakfast / lunch / dinner / snack |
| `mealDate` | string | YYYY-MM-DD |
| `result.foods` | array | NutritionInfo配列 |
| `result.total_calories` | number | 合計カロリー |
| `result.total_protein` | number | 合計タンパク質 |
| `result.total_fat` | number | 合計脂質 |
| `result.total_carbohydrates` | number | 合計炭水化物 |
