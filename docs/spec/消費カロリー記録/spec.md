# Spec: 消費カロリー記録機能

## Linear Issue

- Issue: UTK-39
- タイトル: 消費カロリー記録機能を作成

---

## 1. 機能の定義と境界

### 何を作るか

MealsView（食事記録画面）の間食セクションの下に「消費カロリー」セクションを追加する。
ユーザーが種目と実施時間（分）を選ぶだけで、その日の消費カロリーを記録・概算できる。

### 機能の範囲

- 日次の消費カロリー記録の作成・削除
- 種目のプリセット選択 + 自由入力の併用
- ハイブリッド概算（プリセット種目はMET計算、未知の種目はGemini APIで補完）
- NutritionSummaryCard に消費カロリーと正味カロリーを表示
- 消費カロリーを含めた日次集計のAPI統合

---

## 2. 入力・インターフェース定義

### 2.1 ユーザー入力フィールド

| フィールド | 型 | 必須 | 説明 |
|:---|:---|:---|:---|
| 種目名 | string | 必須 | プリセットから選択 or 自由テキスト入力 |
| 実施時間 | int (分) | 必須 | 5〜600分の範囲 |

メモフィールドは持たない。シンプルに「種目を選んで時間を入力するだけ」のUXとする。

### 2.2 プリセット種目リスト

バックエンドにMET値テーブルとして保持する。

| カテゴリ | 種目名 | MET値 |
|:---|:---|:---|
| 格闘技 | 柔術 | 6.0 |
| 格闘技 | グラップリング | 6.0 |
| 格闘技 | レスリング | 7.0 |
| 格闘技 | キックボクシング | 7.0 |
| 格闘技 | ボクシング | 7.0 |
| 格闘技 | ムエタイ | 7.0 |
| 格闘技 | 総合格闘技（MMA） | 7.5 |
| トレーニング | ウェイトトレーニング | 5.0 |
| トレーニング | マシントレーニング | 4.5 |
| トレーニング | 体幹トレーニング（床） | 3.5 |
| 有酸素 | ランニング | 8.0 |
| 有酸素 | 自転車 | 6.0 |
| 有酸素 | 水泳 | 7.0 |
| その他 | ウォームアップ / ストレッチ | 2.5 |

プリセットに存在しない種目を自由入力した場合は Gemini API で概算する。

### 2.3 消費カロリー計算式

```
消費カロリー(kcal) = MET値 × 体重(kg) × 時間(h)
```

体重は当日の最新体重記録を使用する。体重記録がない場合はデフォルト値 65 kg を使用する。

Gemini API での概算時は、バックエンドが体重情報をプロンプトに含めて問い合わせを行い、kcal 値を返させる。

---

## 3. APIインターフェース定義

### 3.1 新規エンドポイント

#### POST /api/exercise/records

運動記録を作成し、消費カロリーを概算して返す。

リクエスト：

```json
{
  "exercise_name": "柔術",
  "duration_minutes": 90,
  "recorded_date": "2026-02-28"
}
```

レスポンス（201 Created）：

```json
{
  "id": "uuid",
  "exercise_name": "柔術",
  "duration_minutes": 90,
  "burned_calories_kcal": 540.0,
  "estimation_method": "met",
  "recorded_date": "2026-02-28",
  "created_at": "2026-02-28T10:00:00Z"
}
```

`estimation_method`: `"met"` または `"gemini"` を返す。

#### GET /api/exercise/records?date=YYYY-MM-DD

指定日の運動記録一覧を取得する。

レスポンス（200 OK）：

```json
{
  "date": "2026-02-28",
  "records": [
    {
      "id": "uuid",
      "exercise_name": "柔術",
      "duration_minutes": 90,
      "burned_calories_kcal": 540.0,
      "estimation_method": "met",
      "recorded_date": "2026-02-28",
      "created_at": "2026-02-28T10:00:00Z"
    }
  ],
  "total_burned_calories_kcal": 540.0
}
```

#### DELETE /api/exercise/records/{id}

指定IDの運動記録を削除する。

レスポンス（204 No Content）

### 3.2 既存エンドポイントの拡張

#### GET /api/meals/daily （拡張）

`DailyMealsResponse` に運動記録の集計値を追加する。

```json
{
  "date": "2026-02-28",
  "meals": { ... },
  "daily_total": {
    "total_calories": 2000.0,
    "total_protein": 120.0,
    "total_fat": 60.0,
    "total_carbohydrates": 200.0
  },
  "total_burned_calories_kcal": 540.0
}
```

iOS側は `MealsViewModel.loadMeals()` の1回のAPIコールで消費カロリーも取得できる。

---

## 4. データモデル定義

### 4.1 Firestore コレクション

```
users/{userId}/exerciseRecords/{recordId}
  ├── id: string (UUID v4)
  ├── exerciseName: string
  ├── durationMinutes: int
  ├── burnedCaloriesKcal: float64
  ├── estimationMethod: string ("met" | "gemini")
  ├── recordedDate: string (YYYY-MM-DD)
  ├── createdAt: timestamp
  └── updatedAt: timestamp
```

### 4.2 Firestoreインデックス（追加）

```json
{
  "collectionGroup": "exerciseRecords",
  "queryScope": "COLLECTION",
  "fields": [
    { "fieldPath": "recordedDate", "order": "ASCENDING" },
    { "fieldPath": "createdAt", "order": "ASCENDING" }
  ]
}
```

### 4.3 iOS モデル

#### ExerciseRecord

```swift
struct ExerciseRecord: Codable, Identifiable {
    let id: String
    let exerciseName: String
    let durationMinutes: Int
    let burnedCaloriesKcal: Double
    let estimationMethod: String
    let recordedDate: String
    let createdAt: String
}
```

#### ExerciseDailyResponse

```swift
struct ExerciseDailyResponse: Codable {
    let date: String
    let records: [ExerciseRecord]
    let totalBurnedCaloriesKcal: Double
}
```

#### DailyMeals（拡張）

```swift
struct DailyMeals: Codable {
    let date: String
    let meals: MealsByType
    let dailyTotal: DailyTotal
    let totalBurnedCaloriesKcal: Double  // 追加
}
```

---

## 5. UIの仕様

### 5.1 MealsView の変更

間食（snack）セクションの下に `ExerciseSection` を追加する。

```
[NutritionSummaryCard] ← 正味カロリーを追加表示
[朝食セクション]
[昼食セクション]
[夕食セクション]
[間食セクション]
[消費カロリーセクション] ← 新規追加
```

`ExerciseSection` の表示内容：
- タイトル: 「消費カロリー」 + 編集アイコン
- 右上に合計消費カロリー（記録がある場合）
- 各運動記録: `種目名 / xx分 → xxx kcal`
- 記録なし時: 「記録なし」

タップ時: `ExerciseInputSheet` をシートとして表示する。

### 5.2 ExerciseInputSheet の仕様

#### プリセット選択エリア

- カテゴリ別（格闘技・トレーニング・有酸素・その他）に種目を表示
- タップで選択状態になる（選択中はハイライト）

#### 自由入力エリア

- プリセット以外の種目名を入力するテキストフィールド
- プリセット選択時はこのフィールドはクリアされ非アクティブになる
- 自由入力時はプリセット選択が解除される

#### 時間入力

- 分単位の数値入力（ステッパーまたは数値キーパッド）
- 初期値: 60分

#### 追加ボタン

- タップ → `POST /api/exercise/records` を呼び出す
- ローディング中はボタンを非活性化
- 成功時: シートを閉じて MealsView を更新
- エラー時: エラーメッセージをシート内に表示

#### 記録一覧と削除

- `ExerciseInputSheet` 内に当日の記録一覧を表示
- スワイプ削除（左スワイプ → 削除ボタン → 確認なしで即削除）

### 5.3 NutritionSummaryCard の変更

消費カロリーがある場合（> 0）に以下を追加表示する：

```
摂取カロリー: 2000 kcal
消費カロリー: -540 kcal
─────────────────
正味カロリー: 1460 kcal
```

消費カロリーが 0 の場合は現状の表示（摂取カロリーのみ）を維持する。

---

## 6. バリデーション仕様

### バックエンド

| フィールド | バリデーション |
|:---|:---|
| exercise_name | 1〜100文字、空文字不可 |
| duration_minutes | 5〜600の整数 |
| recorded_date | YYYY-MM-DD形式、未来日付不可（当日まで）|
| burned_calories_kcal | 0より大きい値 (計算結果のサニティチェック) |

### iOS

| フィールド | バリデーション |
|:---|:---|
| 種目名（自由入力） | 空文字不可 |
| 時間 | 5〜600分 |

---

## 7. エラーハンドリング

| ケース | バックエンド | iOS表示 |
|:---|:---|:---|
| Gemini API 失敗 | 500 エラーを返す | 「消費カロリーの概算に失敗しました。再試行してください」 |
| レート制限超過 | 429 Too Many Requests | 「しばらくしてからお試しください」 |
| バリデーションエラー | 400 Bad Request + メッセージ | APIエラーメッセージをそのまま表示 |
| 削除対象が見つからない | 404 Not Found | 「記録が見つかりません」（UIは既に削除済みとして扱う） |

---

## 8. 非機能要件

| 項目 | 要件 |
|:---|:---|
| レスポンス（MET計算時） | 500ms以内 |
| レスポンス（Gemini概算時） | 5秒以内 |
| Geminiレート制限 | 既存の `GeminiRateLimit=0.2 rps` に従う |
| データ分離 | userID でスコープし、他ユーザーのデータにはアクセス不可 |
| ログ | バックエンドで推定方法（MET/Gemini）をログ出力 |

---

## 9. 対象外とするスコープ（やらないこと）

- 運動記録の更新機能（削除して再登録で対応）
- 運動記録の履歴閲覧（過去日付の参照は今回の範囲外）
- 週次・月次の消費カロリー集計・グラフ
- カスタムプリセット種目の登録・編集
- Apple ヘルスケアとの連携
- 消費カロリーの目標設定

---

## 10. 実装コンポーネント一覧

### バックエンド（Go）

| ファイル | 変更種別 | 概要 |
|:---|:---|:---|
| `pkg/gemini/exercise_estimator.go` | 新規 | 運動種目の消費カロリーを Gemini で概算するクライアント |
| `internal/service/exercise_service.go` | 新規 | MET計算 + Gemini補完のハイブリッド概算ロジック |
| `internal/repository/exercise_models.go` | 新規 | ExerciseRecord モデル・Repository インターフェース |
| `internal/repository/exercise_repository_firestore.go` | 新規 | Firestore 実装 |
| `internal/handler/exercise_handler.go` | 新規 | REST ハンドラー（POST/GET/DELETE） |
| `internal/handler/daily_meals_handler.go` | 変更 | `total_burned_calories_kcal` を追加返却 |
| `cmd/server/main.go` | 変更 | ルーティング追加・ハンドラー初期化追加 |
| `firestore.indexes.json` | 変更 | `exerciseRecords` インデックス追加 |

### iOS（Swift）

| ファイル | 変更種別 | 概要 |
|:---|:---|:---|
| `Core/Models/Exercise.swift` | 新規 | ExerciseRecord・ExerciseDailyResponse モデル |
| `Core/Models/Meal.swift` | 変更 | DailyMeals に `totalBurnedCaloriesKcal` 追加 |
| `Core/Network/APIEndpoint.swift` | 変更 | 運動記録エンドポイント追加 |
| `Core/Repositories/ExerciseRepository.swift` | 新規 | API通信 Repository |
| `Features/Meals/MealsView.swift` | 変更 | ExerciseSection 追加 |
| `Features/Meals/MealsViewModel.swift` | 変更 | 消費カロリーの読み込み対応 |
| `Features/Meals/Views/ExerciseInputView.swift` | 新規 | 運動記録入力シート（プリセット + 自由入力 + 記録一覧） |
| `Features/Meals/Views/NutritionSummaryCard.swift` | 変更 | 消費カロリー・正味カロリー表示追加 |
| `Features/Meals/ViewModels/ExerciseInputViewModel.swift` | 新規 | 入力シートの状態管理・API呼び出し |
