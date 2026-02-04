# プラン: 量変更時の栄養素再計算機能

## Linear Issue
- Issue: EDG-597
- URL: https://linear.app/ryosuke-horie/issue/EDG-597

## 概要
食事記録でメニューの量を変更した時に、栄養素（カロリー、タンパク質、脂質、炭水化物）を自動で再計算する機能を実装する。

## 要件
- 量の変更を検知して即座に栄養素を再計算（iOS側で計算）
- 対応する量の表記:
  - グラム表記: "100g" → "150g"（比率1.5倍）
  - 数量表記: "1杯" → "2杯"（比率2倍）
- 対応外: "大盛り" などの曖昧な表現

## 実装計画

### Phase 1: 量パーサーの実装（iOS）

新規ファイル: `ios/Uchikomi/Features/Meals/Utils/QuantityParser.swift`

```swift
struct QuantityParser {
    struct ParsedQuantity {
        let value: Double
        let unit: String  // "g", "グラム", "杯", "人前" など
    }

    static func parse(_ text: String) -> ParsedQuantity?
}
```

対応パターン:
- "100g", "100G", "100グラム", "100 g" → value: 100, unit: "g"
- "1杯", "2杯", "1.5杯" → value: 1/2/1.5, unit: "杯"
- "1人前", "2人前" → value: 1/2, unit: "人前"

### Phase 2: FoodEditItem の拡張（iOS）

ファイル: `ios/Uchikomi/Features/Meals/Models/FoodEditItem.swift`

変更内容:
1. 元の量（`originalQuantity`）を保持するプロパティ追加
2. 比率計算メソッド追加

```swift
@Observable
final class FoodEditItem: Identifiable {
    // 既存
    var quantity: String
    var calories: Double
    ...

    // 追加
    private(set) var originalQuantity: String
    private(set) var originalCalories: Double
    private(set) var originalProtein: Double
    private(set) var originalFat: Double
    private(set) var originalCarbohydrates: Double

    func recalculateNutrition()
}
```

### Phase 3: 再計算ロジックの実装（iOS）

ファイル: `ios/Uchikomi/Features/Meals/ViewModels/NutritionEditorViewModel.swift`

変更内容:
- 量の変更を監視（Combine/onChange）
- 変更検知時に `FoodEditItem.recalculateNutrition()` を呼び出し

### Phase 4: UI更新（iOS）

ファイル: `ios/Uchikomi/Features/Meals/Views/FoodItemEditRow.swift`

変更内容:
- 量のTextField変更時に再計算をトリガー
- 栄養素表示をリアルタイム更新（現在は参考値として読み取り専用だが、計算後の値を表示）

## 技術的な考慮事項

### 1. 元の値の保持
- `FoodEditItem` 初期化時に元の量と栄養素を保存
- 再計算時は常に元の値をベースに比率計算（累積誤差防止）

### 2. パース失敗時の挙動
- 元の量または新しい量がパースできない場合は再計算しない
- 栄養素は元の値を維持

### 3. 精度
- 栄養素は小数点以下1桁で丸める

## 変更対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Features/Meals/Utils/QuantityParser.swift` | 新規作成 |
| `ios/Uchikomi/Features/Meals/Models/FoodEditItem.swift` | 元の値保持、再計算メソッド |
| `ios/Uchikomi/Features/Meals/ViewModels/NutritionEditorViewModel.swift` | 変更監視 |
| `ios/Uchikomi/Features/Meals/Views/FoodItemEditRow.swift` | onChange追加 |

## テスト計画

### ユニットテスト
- `QuantityParser` のパーステスト（各パターン）
- `FoodEditItem.recalculateNutrition()` の計算テスト
- 比率計算の精度テスト

### 手動テスト
1. 食事記録を作成（例: 白米100g）
2. 編集画面を開く
3. 量を "150g" に変更
4. 栄養素が1.5倍に更新されることを確認
5. "1杯" → "2杯" の変更でも2倍になることを確認

## バックエンド変更
なし（iOS側のみの変更）
