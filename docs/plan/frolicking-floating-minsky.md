# プラン: 食事記録「食べなかった」機能（iOS側統合）

## Linear Issue
- Issue: EDG-687
- URL: https://linear.app/ryosuke-horie/issue/EDG-687

## 概要

食事記録で「食べなかった」を記録できるようにする。
バックエンドの `POST /api/meals/skip` は実装済み。iOS側のUI・API統合が未実装のため、これを追加する。

## 現状

- バックエンド: `POST /api/meals/skip` 完全実装済み（リクエスト: `{meal_type, meal_date}` / レスポンス: `{id}` / 201 Created）
- iOS Model: `InputType.skipped` は `Meal.swift:40` に定義済み
- iOS API/UI: 未実装

## 実装計画

### Step 1: APIEndpoint追加
- ファイル: `ios/Uchikomi/Core/Network/APIEndpoint.swift`
- `static let skipMeal` を Meals Endpoints セクション（L41の`analyze`の前）に追加
- path: `"meals/skip"`, method: `.post`, requiresAuth: `true`

### Step 2: Repository層の拡張
- ファイル: `ios/Uchikomi/Core/Repositories/MealRepository.swift`

追加内容:
1. `SkipMealRequest` 構造体（`TextAnalyzeRequest` L36-48 と同パターン）
   - `mealType: String` / `mealDate: String` + CodingKeys
2. `MealRepositoryProtocol` に `skipMeal(mealType:mealDate:) async throws` を追加（戻り値なし）
   - IDはクライアント側で不要なので `requestWithoutResponse` を使用（`deleteHistory` L112-114 と同パターン）
3. `MealRepository` に実装を追加
   - `SkipMealRequest` を作成し `apiClient.requestWithoutResponse(endpoint:body:)` を呼ぶ
   - `dateFormatter.string(from:)` で日付をフォーマット

### Step 3: Mock再生成
```bash
task ios:generate-mocks
```

### Step 4: Model層にヘルパー追加
- ファイル: `ios/Uchikomi/Core/Models/Meal.swift`
- `MealsByType` に `isSkipped(for:)` メソッドを追加
  ```swift
  func isSkipped(for type: MealType) -> Bool {
      meals(for: type).contains { $0.inputType == .skipped }
  }
  ```

### Step 5: MealsViewModel にskipMealメソッド追加
- ファイル: `ios/Uchikomi/Features/Meals/MealsViewModel.swift`
- `isSkipping: Bool` プロパティ追加
- `skipMeal(mealType:)` メソッド追加（`deleteHistory` L65-79 と同パターン）
  - API呼び出し -> `loadMeals()` でデータリフレッシュ -> エラーハンドリング

### Step 6: MealsView のUI変更
- ファイル: `ios/Uchikomi/Features/Meals/MealsView.swift`

MealTypeSection の変更:
1. `onSkipped: () -> Void` コールバックを追加
2. `isSkipped` 判定用 computed property を追加:
   ```swift
   private var isSkipped: Bool {
       meals.contains { $0.inputType == .skipped }
   }
   ```
3. 表示を3状態に分岐（判定順序: isSkipped -> isEmpty -> else）:
   - `isSkipped == true`: 「食べませんでした」テキスト表示（kcal表示なし、アイコン: `moon.zzz`）
   - `meals.isEmpty`: 「記録なし」テキスト + 「食べなかった」ボタン
   - else（通常記録あり）: 現在どおり食品リスト表示
4. 「食べなかった」ボタンは独立した `Button` として配置（ヘッダーの`onTapped`とは分離）
5. ヘッダー部分のタップ（入力画面への遷移）はどの状態でも有効
   - skipped状態でもタップで入力画面を開ける（通常記録の追加でskippedは自動解除される）

MealsView の ForEach:
- `onSkipped:` に `viewModel.skipMeal(mealType:)` を接続

### Step 7: MealInputView でのskipped表示対応
- ファイル: `ios/Uchikomi/Features/Meals/MealInputView.swift`

ExistingMealsSection の変更:
- `existingMeals` を skipped と通常記録に分離して表示
- skipped レコード: `ExistingMealCard` の代わりに「食べませんでした」簡易表示 + 「取り消す」ボタン
- 通常レコード: 既存の `ExistingMealCard` で表示
- 「取り消す」は既存の `deletingMeal` フローで `deleteHistory` を呼ぶ
- 注: バックエンドの仕様上、skipped記録と通常記録は同一食事タイプ内で混在しない（skipped作成時に既存記録は全削除される）

### Step 8: テスト追加
- ファイル: `ios/UchikomiTests/Disabled/MealsViewModelTests.swift` に追加
  - 既存テストと同じファイル・同じパターンを維持

テストケース:
1. `skipMeal成功時に食事データがリロードされるべき`
2. `skipMeal失敗時にエラーメッセージが設定されるべき`
3. `skipMeal中はisSkippingがtrueになるべき`
4. `MealsByType.isSkipped` の正常判定テスト

## 変更対象ファイル一覧

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Core/Network/APIEndpoint.swift` | `skipMeal` エンドポイント追加 |
| `ios/Uchikomi/Core/Repositories/MealRepository.swift` | Protocol + 実装 + Request構造体 |
| `ios/Uchikomi/Core/Models/Meal.swift` | `MealsByType.isSkipped(for:)` 追加 |
| `ios/Uchikomi/Features/Meals/MealsViewModel.swift` | `skipMeal()` + `isSkipping` 追加 |
| `ios/Uchikomi/Features/Meals/MealsView.swift` | MealTypeSection の3状態表示対応 |
| `ios/Uchikomi/Features/Meals/MealInputView.swift` | skipped記録の専用表示 |
| `ios/UchikomiTests/Generated/MockGenerated.swift` | 自動再生成 |
| `ios/UchikomiTests/Disabled/MealsViewModelTests.swift` | テストケース追加 |

## 技術的な考慮事項

- バックエンドの `CreateSkippedMeal` は同日同食事タイプの既存記録を全削除する。MealTypeSectionの「食べなかった」ボタンは `meals.isEmpty` の場合のみ表示するため、既存記録の上書き問題は発生しない
- MealInputViewから既存記録がある状態で新規入力すると、バックエンドが `deleteSkippedRecords` で自動的にskipped記録を削除するため、skippedの取り消しは自然に行われる
- `APIClient` は 200-299 を正常扱いするため、バックエンドの 201 Created は問題なく処理される
- `requestWithoutResponse` を使用するため、レスポンスモデルの定義は不要

## テスト計画

1. `task ios:generate-mocks` でMock再生成
2. 既存テスト + 新規テストケースの実行: `task ios:test`
3. バックエンドテスト: `task test`（変更なしだが確認）
4. 手動確認（Chrome DevTools MCP）:
   - 記録なし状態で「食べなかった」ボタンが表示されること
   - タップ後に「食べませんでした」表示に切り替わること
   - 入力画面からskipped記録を取り消せること
   - 通常記録を追加するとskippedが自動解除されること
