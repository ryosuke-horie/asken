# Research: 消費カロリー記録機能

## Linear Issue

- Issue: UTK-39
- タイトル: 消費カロリー記録機能を作成
- 概要: topの食事記録の間食の下に追加する。マシントレやジムでの練習などを入力することで食事と同じように消費カロリーを概算する。

---

## 1. 既存システムの構造

### 1.1 食事記録画面（MealsView）の現状

`ios/Uchikomi/Features/Meals/MealsView.swift` において、以下の順序でUIが構成されている：

1. `NutritionSummaryCard` - 日次栄養素合計サマリー（カロリー・PFC表示）
2. `ForEach(MealType.allCases)` - 食事タイプ別セクション（朝食/昼食/夕食/間食）
   - 各セクションは `MealTypeSection` コンポーネントで描画

「間食の下に追加」とはこの `ForEach` の直後に新しいセクションを追加することを意味する。

### 1.2 既存の MealType 定義

```swift
// MealType.allCases の順序 → 朝食, 昼食, 夕食, 間食
```

間食 (`snack`) が現状の最後の食事タイプ。新機能はその下のセクションとして追加する。

### 1.3 DailyMeals のデータ構造

`GET /api/meals/daily` が返す `DailyMealsResponse`：

```go
type DailyMealsResponse struct {
    Date       string                                `json:"date"`
    Meals      map[string][]repository.HistoryDetail `json:"meals"`
    DailyTotal repository.DailyTotal                 `json:"daily_total"`
}

type DailyTotal struct {
    TotalCalories       float64
    TotalProtein        float64
    TotalFat            float64
    TotalCarbohydrates  float64
    TotalMicronutrients map[string]float64
}
```

現状の `DailyTotal` は摂取カロリーのみで、消費カロリーは含まれていない。

### 1.4 既存の Firestore コレクション構造

```
users/{userId}/
  ├── analysisRequests/{requestId}   # 食事分析・記録
  ├── weightRecords/{recordId}       # 体重記録
  ├── weightGoal/current             # 目標体重
  ├── nutritionGoal/current          # 栄養目標
  ├── myMenu/{menuId}                # マイメニュー
  ├── ingredients/{ingredientId}     # 食材
  └── menuSuggestions/{suggestionId} # AIメニューサジェスト
```

`analysisRequests` コレクションの主要フィールド：
- `status`: pending/processing/completed/failed
- `inputType`: image/text/mylist/skipped/suggestion
- `mealType`: breakfast/lunch/dinner/snack
- `mealDate`: timestamp
- `confirmed`: boolean
- `result`: { foods[], totalCalories, totalProtein, totalFat, totalCarbohydrates }

### 1.5 Gemini API クライアントの現状

`pkg/gemini/` に以下のクライアントが存在する：

| クライアント | 役割 |
|:---|:---|
| `Classifier` | 食事画像から食材を識別 |
| `TextParser` | テキスト入力から食材を抽出 |
| `NutritionCalculator` | 食材から栄養素を計算 |
| `ReceiptParser` | レシート解析 |
| `MenuSuggester` | メニュー提案 |

消費カロリー計算用の Gemini クライアントは未実装。

---

## 2. 技術的実現方法の選択肢

### 選択肢A: 既存 `analysisRequests` コレクションを拡張

新しい `inputType = "exercise"` を追加し、食事記録と同一コレクションに記録する。

- `mealType` は使用しない（または "exercise" などを追加）
- `result.totalCalories` を負値で記録し消費カロリーを表現
- または `result.burnedCalories` フィールドを追加

メリット:
- 既存の日次集計ロジック (`GetDailyMeals`) を再利用可能
- `DailyTotal` に消費カロリーを加算するだけで対応可能

デメリット:
- 食事記録と運動記録の意味が混在する
- `foods` 配列など食事専用フィールドが空になり、コレクションの意味が曖昧になる
- `mealType` の扱いが不明確

### 選択肢B: 新規 `exerciseRecords` コレクションを作成

運動記録専用のコレクションを独立して作成する。

```
users/{userId}/
  └── exerciseRecords/{recordId}
        ├── exerciseName: string      # 種目名 (例: "マシントレ", "スパー")
        ├── durationMinutes: int      # 実施時間（分）
        ├── burnedCaloriesKcal: float64  # 消費カロリー（Gemini概算）
        ├── note: string              # メモ
        ├── recordedDate: string      # 実施日 (YYYY-MM-DD)
        ├── createdAt: timestamp
        └── updatedAt: timestamp
```

メリット:
- 食事記録との明確な分離
- 将来の拡張（種目別統計、ワークアウト履歴など）に対応しやすい
- 専用のバリデーション・ロジックが適用できる

デメリット:
- 新規コレクション・Firestore インデックス・リポジトリ・ハンドラー・サービスが必要
- 日次集計 API (`/api/meals/daily`) に運動データを統合するための改修が必要

### 推奨アプローチ

選択肢B（新規コレクション）を推奨。

理由：
- 食事記録と運動記録はドメインが異なる（摂取 vs 消費）
- 既存の `analysisRequests` コレクションは Gemini 非同期ワーカーとの連携を前提とした複雑な状態管理（pending/processing/completed）を持ち、即時記録の運動記録には不適切
- 将来的なトレーニング記録機能の拡張に備えて独立させておく方が自然

---

## 3. 消費カロリーの概算方法

### 3.1 Gemini API を使う方法

ユーザーが「マシントレ 1時間」「スパー 2ラウンド」などを入力すると、Gemini API がカロリーを概算する。

既存の `TextParser` や `NutritionCalculator` と同様のパターンで `ExerciseCalorieEstimator` を実装する。

プロンプト設計の方針：
- 体重情報があれば考慮する（消費カロリーは体重依存）
- 格闘技特有の種目（スパーリング、打ち込み、走り込みなど）を理解させる
- 一般的なMETS値を参考にした計算

メリット：
- ユーザーが自由なテキストで入力できる（食事記録と同じUX）
- 格闘技特有の練習種目も対応可能

デメリット：
- Gemini APIへの依存（コスト・レイテンシ・レート制限）
- 概算精度の保証が困難
- レート制限 (GeminiRateLimit=0.2 rps) の影響を受ける

### 3.2 固定の計算式（MET値ベース）

MET（Metabolic Equivalent of Task）値を使って計算する：
```
消費カロリー = MET値 × 体重(kg) × 時間(h)
```

代表的な MET 値（格闘技関連）：
- 一般的な筋力トレーニング: MET 3-6
- ジム（マシントレ）: MET 4-5
- スパーリング: MET 6-8
- ミット打ち: MET 6-7
- ランニング（一般）: MET 7-9

メリット：
- Gemini API 不要でコスト・レイテンシゼロ
- 計算の透明性・予測可能性が高い

デメリット：
- 固定のMET値テーブルが必要
- ユーザーが自由記述できず、プルダウン等から選択する UX になる
- 格闘技特有の複雑な練習を正確に分類できない

### 推奨アプローチ

Gemini API を使う方法を推奨。

理由：
- 課題の説明「食事と同じように消費カロリーを概算する」がGemini APIを使った自由入力UXを示唆している
- 既存の食事記録と同じテキスト入力フローで実装できるため、UXの一貫性が保てる
- 格闘技特有の練習記述（「スパー3ラウンド」「打ち込み30分」等）に柔軟に対応できる
- ただし、レート制限対策として食事分析と同じ制限（GeminiRateLimit=0.2 rps）を適用する

計算式でも対応可能かもしれないですね。
柔術 90分とかキックボクシング 60分とか自転車30分とかの入力が考えられます

---

## 4. 既存コードとの統合ポイント

### 4.1 日次食事API (`GET /api/meals/daily`) への統合

`DailyMealsResponse` に `TotalBurnedCalories` を追加し、消費カロリー記録一覧も返す：

```go
type DailyMealsResponse struct {
    Date            string                                `json:"date"`
    Meals           map[string][]repository.HistoryDetail `json:"meals"`
    DailyTotal      repository.DailyTotal                 `json:"daily_total"`
    ExerciseRecords []repository.ExerciseRecord           `json:"exercise_records"` // 新規追加
}
```

または専用エンドポイント `GET /api/exercise/records?date=YYYY-MM-DD` を作成する方法もある。

### 4.2 iOS側の NutritionSummaryCard への反映

現状の表示：
- 摂取カロリー total
- PFC

追加検討：
- 消費カロリー表示（-xxx kcal）
- 正味カロリー表示（摂取 - 消費）

---

## 5. 既存の体重情報との連携

消費カロリーの Gemini 概算精度を上げるため、ユーザーの体重情報を取得してプロンプトに組み込むことができる。

現状のデータフロー（`MealsViewModel.loadMeals`）：
1. 食事データ取得（並列）
2. 直近7日の体重データ取得（並列）
3. 栄養目標取得（体重を渡す）

同様のパターンで、運動記録作成時に最新の体重をコンテキストとして渡す設計が可能。

---

## 6. セキュリティ・バリデーション要件

既存の体重記録 (`WeightRecord`) のバリデーションパターンを参考：

- 数値の NaN/Inf チェック
- 範囲チェック（消費カロリー: 0〜9999 kcal/日など合理的な範囲）
- メモフィールド: 200文字以内
- 日時バリデーション: RFC3339 形式、未来日時の禁止（5分まで許容）
- テキスト入力: XSS対策（Goのhtml/templateやサニタイズ）
- ユーザースコープの徹底（`userID` によるデータ分離）

---

## 7. Firestore インデックス要件

新規 `exerciseRecords` コレクションに必要なインデックス：

```json
{
  "collectionGroup": "exerciseRecords",
  "queryScope": "COLLECTION",
  "fields": [
    { "fieldPath": "recordedDate", "order": "ASCENDING" }
  ]
}
```

日次取得クエリ (`recordedDate == "YYYY-MM-DD"`) はコレクションが小さい場合は単一フィールドで対応可能。期間取得が必要になった場合は `recordedDate` の範囲クエリインデックスを追加する。

---

## 8. 関連する外部知識・ベストプラクティス

### 8.1 格闘技の消費カロリー概算（参考）

多分毎回の練習内容を細かく記録できる人は１割にも満たないはず

| 種目 | 強度 | MET | 体重60kgで1時間の消費カロリー目安 |
|:---|:---|:---|:---|
| マシントレ（筋トレ） | 中 | 4-5 | 240-300 kcal |
| ミット打ち / 打ち込み | 高 | 6-8 | 360-480 kcal |
| スパーリング | 高 | 7-9 | 420-540 kcal |
| ランニング（一般） | 中〜高 | 7-9 | 420-540 kcal |
| 縄跳び | 高 | 8-12 | 480-720 kcal |
| ウォームアップ / ストレッチ | 低 | 2-3 | 120-180 kcal |

Gemini はこれらの値を参考に概算する。

### 8.2 既存の Gemini API 活用パターン

`pkg/gemini/text_parser.go` のパターンに倣い、`ExerciseCalorieEstimator` を実装する：
- JSON スキーマを使って構造化された出力を要求
- タイムアウトを 120 秒に設定
- レート制限: `GeminiRateLimit=0.2` rps を遵守

---

## 9. 未解決の論点

以下の点は仕様策定時にユーザーに確認が必要：

1. 消費カロリーを NutritionSummaryCard に反映させるか（摂取カロリーとの差し引き表示）
2. 日次集計 API に統合するか、専用エンドポイントを作るか
3. 運動記録の CRUD は全て必要か（削除のみ？更新も？）
4. Gemini で概算した値をユーザーが手動修正できるようにするか
5. 体重情報を Gemini プロンプトに渡すか（精度向上のため）

---

## 10. 関連ファイル一覧

### バックエンド（Go）

| ファイル | 関連性 |
|:---|:---|
| `backend/cmd/server/main.go` | ルーティング追加対象 |
| `backend/internal/handler/daily_meals_handler.go` | 拡張候補 |
| `backend/internal/handler/weight_record_handler.go` | 実装パターンの参考 |
| `backend/internal/repository/analysis_models.go` | AnalysisRepository 参考 |
| `backend/internal/repository/weight_models.go` | WeightRecord 参考 |
| `backend/pkg/gemini/text_parser.go` | Gemini クライアント実装パターン |
| `firestore.indexes.json` | インデックス追加対象 |

### iOS（Swift）

| ファイル | 関連性 |
|:---|:---|
| `ios/Uchikomi/Features/Meals/MealsView.swift` | 消費カロリーセクション追加対象 |
| `ios/Uchikomi/Features/Meals/MealsViewModel.swift` | 運動データ取得ロジック追加対象 |
| `ios/Uchikomi/Core/Models/Meal.swift` | DailyMeals モデル拡張候補 |
