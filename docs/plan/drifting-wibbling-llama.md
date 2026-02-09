# EDG-707: 食事記録の単位変更の計算ロジックの修正

## Workflow Status
- [x] Phase 1: 準備
- [x] Phase 2: 計画
- [x] Phase 3: 実装
- [ ] Phase 4: レビューサイクル
- [ ] Phase 5: PR 作成
- [ ] Phase 6: ユーザー確認

**Current Phase**: Phase 4
**Branch**: edg-707
**Issue**: EDG-707

## 概要

現在、iOSアプリでは食事記録の量（単位と数値）が単一の文字列として扱われており、単位変更時の適切な再計算が行われていません。本プランでは、数値と単位を分離し、単位変更時にLLMで再計算する機能を実装します。

### 現状の問題

1. 単位と数値が分離されていない（`quantity: String` で "100g" や "1杯" を扱っている）
2. 単位変更時（例: 100g→1杯、1杯→1合）に栄養素が再計算されない
3. 数値入力に半角/全角のバリデーションがない

### ユーザーの要件

1. 単位と数値部分を分離したい
2. 数値部分には半角/全角のバリデーションを入れる
3. 数値のみ変更の場合（100g→200g）は比率で再計算
4. 単位が変わった場合（杯→合）はLLMで再計算

---

## 実装計画

### Phase 1: データモデルと単位列挙型

**新規ファイル**: `ios/Uchikomi/Features/Meals/Models/MeasurementUnit.swift`

```swift
enum MeasurementUnit: String, CaseIterable, Identifiable, Codable {
    case gram = "g"
    case cup = "杯"
    case serving = "人前"
    case piece = "個"
    case sheet = "枚"
    case stick = "本"
    case slice = "切れ"
    case meal = "食"
    case plate = "皿"
    case set = "膳"
    case block = "丁"
    case bundle = "束"
    case bag = "袋"
    case can = "缶"
    case go = "合"
    case ball = "玉"

    var id: String { rawValue }
    var displayName: String { rawValue }

    var inputType: QuantityInputType {
        switch self {
        case .gram: return .decimal
        default: return .integer
        }
    }
}

enum QuantityInputType {
    case integer    // 半角数字のみ
    case decimal    // 小数許容
}
```

**修正ファイル**: `ios/Uchikomi/Features/Meals/Models/QuantityParser.swift`

```swift
extension QuantityParser {
    /// 既存の文字列からMeasurementUnitを取得
    static func parseUnit(_ text: String) -> MeasurementUnit? {
        guard let parsed = parse(text) else { return nil }
        return MeasurementUnit(rawValue: parsed.unit)
    }

    /// 数値部分のみを文字列として抽出
    static func parseValue(_ text: String) -> String? {
        guard let parsed = parse(text) else { return nil }
        return parsed.value == floor(parsed.value)
            ? String(Int(parsed.value))
            : String(parsed.value)
    }
}
```

**新規ファイル**: `ios/Uchikomi/Features/Meals/Models/QuantityValidator.swift`

```swift
enum QuantityValidator {
    /// 半角数字（整数）のバリデーション
    static func isValidInteger(_ text: String) -> Bool {
        guard !text.isEmpty else { return true }
        return text.allSatisfy { $0.isNumber && $0.isASCII }
    }

    /// 半角数字（小数許容）のバリデーション
    static func isValidDecimal(_ text: String) -> Bool {
        guard !text.isEmpty else { return true }
        let decimalPattern = /^(\d+\.?\d*|\.\d+)$/
        return text.wholeMatch(of: decimalPattern) != nil
    }

    /// 全角数字を半角に変換
    static func normalizeFullWidth(_ text: String) -> String {
        let fullWidthToHalfWidthMap: [Character: Character] = [
            "０": "0", "１": "1", "２": "2", "３": "3", "４": "4",
            "５": "5", "６": "6", "７": "7", "８": "8", "９": "9",
            "．": "."
        ]
        return String(text.map { fullWidthToHalfWidthMap[$0] ?? $0 })
    }
}
```

---

### Phase 2: FoodEditItemの拡張

**修正ファイル**: `ios/Uchikomi/Features/Meals/Models/FoodEditItem.swift`

以下の変更を追加:

1. 新規プロパティ:
   - `var quantityValue: String` - 数値部分
   - `var quantityUnit: MeasurementUnit?` - 選択中の単位

2. 新規メソッド:
   - `func updateQuantityString()` - quantityValueとquantityUnitからquantityを生成
   - `var hasUnitChanged: Bool` - 単位変更検知

3. イニシャライザの拡張:
   - `convenience init(from nutritionInfo: NutritionInfo)` でパース結果を設定

```swift
// 追加する実装の概要
@Observable
final class FoodEditItem: Identifiable {
    // 既存プロパティ...
    var quantityValue: String = ""
    var quantityUnit: MeasurementUnit?

    var hasUnitChanged: Bool {
        guard let originalUnit = QuantityParser.parse(originalQuantity)?.unit,
              let current = quantityUnit else { return false }
        return originalUnit != current.rawValue
    }

    func updateQuantityString() {
        guard let unit = quantityUnit, !quantityValue.isEmpty else {
            quantity = ""
            return
        }
        quantity = "\(quantityValue)\(unit.rawValue)"
    }
}
```

---

### Phase 3: UIの変更

**修正ファイル**: `ios/Uchikomi/Features/Meals/Views/FoodItemEditRow.swift`

現在の単一TextFieldを、数値入力と単位選択の2要素に変更:

```swift
HStack(spacing: 8) {
    TextField("数値", text: $item.quantityValue)
        .textFieldStyle(.roundedBorder)
        .keyboardType(item.quantityUnit?.inputType == .decimal ? .decimalPad : .numberPad)
        .onChange(of: item.quantityValue) {
            item.quantityValue = QuantityValidator.normalizeFullWidth(item.quantityValue)
            item.updateQuantityString()
            item.recalculateNutrition()
        }

    Picker("単位", selection: $item.quantityUnit) {
        Text("選択").tag(nil as MeasurementUnit?)
        ForEach(MeasurementUnit.allCases) { unit in
            Text(unit.displayName).tag(unit as MeasurementUnit?)
        }
    }
    .pickerStyle(.menu)
    .onChange(of: item.quantityUnit) { _, _ in
        item.updateQuantityString()
        item.handleUnitChange()
    }
}

if item.hasUnitChanged {
    Text("保存後に栄養素が再計算されます")
        .font(.caption2)
        .foregroundStyle(.orange)
}
```

---

### Phase 4: ViewModelの拡張

**修正ファイル**: `ios/Uchikomi/Features/Meals/ViewModels/NutritionEditorViewModel.swift`

単位変更時のメッセージ表示を追加:

```swift
var hasAnyUnitChanged: Bool {
    foods.contains { $0.hasUnitChanged }
}

func save() async {
    // 既存の処理に加えて、単位変更がある場合はメッセージを表示
    if hasAnyUnitChanged {
        recalculatingMessage = "栄養素を再計算中です..."
    }
    // 既存の保存処理...
}
```

---

### Phase 5: テスト

**新規テスト**: `ios/UchikomiTests/Disabled/Features/Meals/MeasurementUnitTests.swift`

**新規テスト**: `ios/UchikomiTests/Disabled/Features/Meals/QuantityValidatorTests.swift`

**既存テスト更新**: `ios/UchikomiTests/Disabled/Features/Meals/FoodEditItemTests.swift`

- `hasUnitChanged` のテスト
- `quantityValue`/`quantityUnit` の初期化テスト

---

## 技術的な考慮事項

### 既存データの互換性

- 既存の `quantity: String` プロパティは残し、`quantityValue`/`quantityUnit` と同期
- `NutritionInfo` から初期化時にパースして設定
- 保存時は文字列に結合してバックエンドへ送信

### UI/UX

- 初期状態では単位が未選択（`nil`）を許容
- 既存データからの復元時にパース失敗した場合、フォールバック処理
- キーボードタイプを単位に応じて切り替え（整数/小数）

### エッジケース

- `quantityValue` が空の状態での単位変更
- 既存データの `quantity` がパースできない形式（"大盛り"など）のハンドリング

---

## 検証

1. 単体テストを実行
2. iOSシミュレータで動作確認
3. 数値変更時の比率計算を確認
4. 単位変更時の「保存後に再計算」メッセージを確認
5. Chrome DevTools MCP でバックエンドAPIの動作確認

---

## Critical Files

- `ios/Uchikomi/Features/Meals/Models/FoodEditItem.swift`
- `ios/Uchikomi/Features/Meals/Views/FoodItemEditRow.swift`
- `ios/Uchikomi/Features/Meals/Models/QuantityParser.swift`
- `ios/Uchikomi/Features/Meals/ViewModels/NutritionEditorViewModel.swift`
