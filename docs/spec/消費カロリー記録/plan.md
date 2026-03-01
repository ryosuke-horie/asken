# Plan: 消費カロリー記録機能

## Linear Issue

- Issue: UTK-39
- タイトル: 消費カロリー記録機能を作成
- Spec: `docs/spec/消費カロリー記録/spec.md`

---

## 実装順序と依存関係

バックエンド → iOS の順で実装する。
バックエンドのAPIが完成してから iOS 側の実装・動作確認を行う。

---

## Phase 1: バックエンド

### 1-1. Firestoreインデックス追加

- [ ] `firestore.indexes.json` に `exerciseRecords` コレクションのインデックスを追加
  - フィールド: `recordedDate` ASC + `createdAt` ASC

### 1-2. データモデル・Repositoryインターフェース定義

- [ ] `backend/internal/repository/exercise_models.go` を新規作成
  - `ExerciseRecord` 構造体（id, exerciseName, durationMinutes, burnedCaloriesKcal, estimationMethod, recordedDate, createdAt, updatedAt）
  - `ExerciseRepository` インターフェース（Create, ListByDate, Delete）
  - バリデーション関数（exerciseName: 1〜100文字、durationMinutes: 5〜600）

### 1-3. Firestore Repository 実装

- [ ] `backend/internal/repository/exercise_repository_firestore.go` を新規作成
  - `Create`: `users/{userId}/exerciseRecords` にドキュメント作成
  - `ListByDate`: `recordedDate` でフィルタリングして一覧取得（`createdAt` 昇順）
  - `Delete`: ドキュメント削除（userID スコープ検証必須）
  - `GetLatestWeightForUser`: 消費カロリー計算用の最新体重取得（既存の WeightRepository から移譲も検討）
  - ユニットテスト作成（Firestoreエミュレータを使用）

### 1-4. Exercise サービス（MET計算 + Gemini補完）

- [ ] `backend/internal/service/exercise_service.go` を新規作成
  - MET値テーブルを定数として定義（spec.md の14種目）
  - `EstimateCalories(exerciseName, durationMinutes, weightKg)` 関数
    - プリセット種目に一致 → MET値計算: `MET × weightKg × (durationMinutes / 60.0)`
    - 一致なし → Gemini API で概算
  - 結果に `estimationMethod` ("met" | "gemini") を付与
  - ユニットテスト作成（MET計算の境界値、Gemini呼び出しのモック）

### 1-5. Gemini ExerciseEstimator

- [ ] `backend/pkg/gemini/exercise_estimator.go` を新規作成
  - 既存の `TextParser` と同様のパターンで実装
  - タイムアウト: 120秒
  - プロンプト: 種目名・時間・体重を渡し、消費カロリー（kcal）を数値で返させる
  - JSON スキーマで構造化出力を要求（`{ "burned_calories_kcal": 540.0 }`）
  - ユニットテスト作成（モッククライアントを使用）

### 1-6. Exercise ハンドラー

- [ ] `backend/internal/handler/exercise_handler.go` を新規作成
  - `HandleCreate`: `POST /api/exercise/records`
    - リクエスト: `exercise_name`, `duration_minutes`, `recorded_date`
    - ユーザーの最新体重を取得してカロリー計算
    - Firestore に保存してレスポンス返却
    - バリデーション・エラーハンドリング実装
  - `HandleList`: `GET /api/exercise/records?date=YYYY-MM-DD`
    - 日付指定で一覧取得、`total_burned_calories_kcal` を集計して返却
  - `HandleDelete`: `DELETE /api/exercise/records/{id}`
    - userID スコープで削除
  - ユニットテスト作成（各エンドポイントの正常系・異常系）

### 1-7. DailyMealsHandler の拡張

- [ ] `backend/internal/handler/daily_meals_handler.go` を変更
  - `ExerciseRepository` を依存として追加
  - `GetDailyMeals` 呼び出し時に同日の運動記録も取得（並列取得）
  - レスポンスに `total_burned_calories_kcal` を追加
  - ユニットテスト更新

### 1-8. ルーティング追加

- [ ] `backend/cmd/server/main.go` を変更
  - `ExerciseRepository` の初期化を `initRepositories` に追加
  - `ExerciseHandler` の初期化・依存注入を追加
  - `setupExerciseRoutes` 関数を追加してルーティング登録
    - `POST /api/exercise/records`
    - `GET /api/exercise/records`
    - `DELETE /api/exercise/records/{id}`

### 1-9. バックエンド動作確認

- [ ] `task lint` が通ること
- [ ] `task test` が通ること（カバレッジ80%以上）
- [ ] `task run` でサーバーを起動し、curl で各エンドポイントの動作確認

---

## Phase 2: iOS

### 2-1. データモデル定義

- [ ] `ios/Uchikomi/Core/Models/Exercise.swift` を新規作成
  - `ExerciseRecord` 構造体（Codable, Identifiable）
  - `ExerciseDailyResponse` 構造体

- [ ] `ios/Uchikomi/Core/Models/Meal.swift` を変更
  - `DailyMeals` に `totalBurnedCaloriesKcal: Double` を追加
  - デコード: `decodeIfPresent` でデフォルト0に対応（後方互換性）

### 2-2. APIエンドポイント追加

- [ ] `ios/Uchikomi/Core/Network/APIEndpoint.swift` を変更
  - `createExerciseRecord` （POST /api/exercise/records）
  - `exerciseRecords(date:)` （GET /api/exercise/records?date=...）
  - `deleteExerciseRecord(id:)` （DELETE /api/exercise/records/{id}）

### 2-3. Exercise Repository

- [ ] `ios/Uchikomi/Core/Repositories/ExerciseRepository.swift` を新規作成
  - `ExerciseRepositoryProtocol` プロトコル定義（`@mockable` タグ付き）
  - `ExerciseRepository` クラス実装
  - `createRecord(exerciseName:durationMinutes:recordedDate:)` → POST
  - `getRecords(date:)` → GET
  - `deleteRecord(id:)` → DELETE

### 2-4. ExerciseInputViewModel

- [ ] `ios/Uchikomi/Features/Meals/ViewModels/ExerciseInputViewModel.swift` を新規作成
  - `@Observable` クラスとして実装
  - 状態: `selectedPreset: ExercisePreset?`, `customExerciseName: String`, `durationMinutes: Int`, `records: [ExerciseRecord]`, `isLoading: Bool`, `errorMessage: String?`
  - プリセット種目リスト（カテゴリ別）
  - `addRecord()`: API呼び出し → 成功時にリストを更新
  - `deleteRecord(id:)`: API呼び出し → リストから除去
  - バリデーション（種目名空文字チェック、時間範囲チェック）

### 2-5. ExerciseInputView

- [ ] `ios/Uchikomi/Features/Meals/Views/ExerciseInputView.swift` を新規作成
  - シート形式の入力UI
  - セクション1: プリセット種目グリッド（カテゴリ別、タップで選択・ハイライト）
  - セクション2: 自由入力テキストフィールド（プリセット選択時は無効化）
  - セクション3: 時間入力（分）- ステッパーまたは数値キーパッド、初期値60分
  - 「追加」ボタン（バリデーション通過時のみ活性、ローディング中は非活性）
  - セクション4: 当日の記録一覧（種目名/時間/kcal、左スワイプ削除）
  - エラー表示

### 2-6. MealsView の変更

- [ ] `ios/Uchikomi/Features/Meals/MealsView.swift` を変更
  - `ForEach(MealType.allCases)` の後に `ExerciseSection` を追加
  - `ExerciseSection`: タイトル・合計消費カロリー・記録リスト表示
  - タップ時に `ExerciseInputSheet` をシートとして表示

### 2-7. MealsViewModel の変更

- [ ] `ios/Uchikomi/Features/Meals/MealsViewModel.swift` を変更
  - `DailyMeals` の `totalBurnedCaloriesKcal` を参照する computed property を追加（必要に応じて）
  - `loadMeals()` は既存のまま（`/api/meals/daily` に `total_burned_calories_kcal` が追加されるため変更不要）

### 2-8. NutritionSummaryCard の変更

- [ ] `ios/Uchikomi/Features/Meals/Views/NutritionSummaryCard.swift` を変更（ファイルパスを確認して対象ファイルに修正）
  - `totalBurnedCaloriesKcal` 引数を追加
  - 消費カロリー > 0 の場合のみ「消費カロリー」「正味カロリー」行を追加表示
  - 呼び出し元（MealsView）の引数を更新

---

## Phase 3: 動作確認とリリース準備

### 3-1. 結合動作確認

- [ ] バックエンドを起動し、iOSシミュレータ or 実機で動作確認
  - プリセット種目を選択 → 時間入力 → 追加 → 消費カロリーが表示されること
  - 自由入力（プリセット外の種目）→ Gemini 概算が動作すること
  - 記録の削除が正常に動作すること
  - NutritionSummaryCard に正味カロリーが表示されること
  - 消費カロリー0の場合、正味カロリー行が非表示であること

### 3-2. コードレビュー・品質確認

- [ ] `task lint` が通ること
- [ ] `task test` が通ること（バックエンド カバレッジ80%以上）
- [ ] セキュリティ確認
  - userID スコープが全エンドポイントで機能していること
  - バリデーションが全フィールドに適用されていること
  - Gemini への入力がサニタイズされていること

### 3-3. PR作成

- [ ] ブランチ: `edg-39` またはIssue番号に対応するブランチ名で作成
- [ ] PR description に `Closes UTK-39` を記載
- [ ] CI が全て通ること（lint, test）

---

## 懸念点・リスクと対処方針

### 懸念点1: WeightRepository からの体重取得

Exercise ハンドラーは消費カロリー計算のために最新体重が必要だが、現在 `WeightRecordRepository` は `WeightRecordHandler` が所有している。

対処: `ExerciseHandler` に `WeightRecordRepository` を追加の依存として渡す。既存の `initRepositories` で生成済みのインスタンスを共有する。

### 懸念点2: Gemini API のレート制限

自由入力時は Gemini API を呼び出すため、既存の食事分析と同じレート制限（0.2 rps）の影響を受ける可能性がある。

対処: 既存の `RateLimitMiddleware` に組み込み済みのため設定変更は不要。ユーザーへのエラーメッセージで「しばらく待ってから再試行」を案内する。

### 懸念点3: DailyMeals の後方互換性

既存の iOS クライアントが `totalBurnedCaloriesKcal` を知らない場合に備え、`decodeIfPresent` でデフォルト 0 とする。

### 懸念点4: NutritionSummaryCard のファイルパス

`NutritionSummaryCard` がどのファイルにあるか要確認（`MealsView.swift` 内に定義されている可能性がある）。実装前にファイルを確認してから変更箇所を特定する。

---

## 実装に関わるファイル一覧（変更順）

### バックエンド

```
firestore.indexes.json                                    変更
backend/internal/repository/exercise_models.go           新規
backend/internal/repository/exercise_repository_firestore.go  新規
backend/pkg/gemini/exercise_estimator.go                  新規
backend/internal/service/exercise_service.go              新規
backend/internal/handler/exercise_handler.go              新規
backend/internal/handler/daily_meals_handler.go           変更
backend/cmd/server/main.go                                変更
```

### iOS

```
ios/Uchikomi/Core/Models/Exercise.swift                   新規
ios/Uchikomi/Core/Models/Meal.swift                       変更
ios/Uchikomi/Core/Network/APIEndpoint.swift               変更
ios/Uchikomi/Core/Repositories/ExerciseRepository.swift   新規
ios/Uchikomi/Features/Meals/ViewModels/ExerciseInputViewModel.swift  新規
ios/Uchikomi/Features/Meals/Views/ExerciseInputView.swift 新規
ios/Uchikomi/Features/Meals/MealsView.swift               変更
ios/Uchikomi/Features/Meals/MealsViewModel.swift          変更
ios/.../NutritionSummaryCard (要ファイル確認)             変更
```
