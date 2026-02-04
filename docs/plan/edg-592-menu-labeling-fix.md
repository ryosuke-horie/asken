# プラン: 料理分解問題とUI表記の修正

## Linear Issue
- Issue: EDG-592
- URL: https://linear.app/ryosuke-horie/issue/EDG-592

## 概要

現在の問題:
1. テキスト入力時にLLMが料理を食材に分解してしまう（例: ラーメン → 麺、スープ、チャーシュー...）
2. UIの「食材」という表記が実態に合っていない

期待する動作:
- 「醤油ラーメン」は「醤油ラーメン」としてそのまま扱う
- 「幕の内弁当」も分解せず1つの料理として扱う
- UI表記を「メニュー」に統一

---

## 修正計画

### 1. バックエンド: プロンプト修正

**ファイル:** `backend/pkg/gemini/text_parser.go` (36-53行)

現在のプロンプト:
```
重要なルール:
- 料理名（例: チキンカツ定食、親子丼）の場合は、構成する食材に分解してください
```

修正後のプロンプト:
```
重要なルール:
- 料理名（例: チキンカツ定食、親子丼、ラーメン）はそのまま1つの料理として扱ってください
- 食材に分解しないでください（例: ラーメン → 麺、スープ、チャーシューに分解しない）
- 量が明記されていない場合は一般的な1食分の量を推定してください
- 「大盛り」「おかわり」などの表現は量に反映してください
- 個数（2個、3杯など）は適切な量に変換してください
- 日本語の料理名を使用してください
```

### 2. iOS: UI表記の変更

| ファイル | 行 | 現在 | 修正後 |
|:---|:---|:---|:---|
| `FoodItemEditRow.swift` | 12 | `Text("食材")` | `Text("メニュー")` |
| `MealInputView.swift` | 392 | `Label("食品を追加", ...)` | `Label("メニューを追加", ...)` |
| `MealInputView.swift` | 465 | `Text("検出された食品")` | `Text("検出されたメニュー")` |
| `NutritionEditorView.swift` | 55 | `Label("食材を追加", ...)` | `Label("メニューを追加", ...)` |

### 3. iOS: プレースホルダーテキストの変更

**ファイル:** `FoodItemEditRow.swift` (22-25行)

現在:
```swift
TextField("料理名（例：鶏むね肉、ご飯）", text: $item.name)
TextField("量（例：100g、1杯、大盛り）", text: $item.quantity)
```

修正後:
```swift
TextField("メニュー名（例：醤油ラーメン、カレーライス）", text: $item.name)
TextField("量（例：1杯、1人前、大盛り）", text: $item.quantity)
```

---

## 修正対象ファイル一覧

| ファイル | 変更内容 |
|:---|:---|
| `backend/pkg/gemini/text_parser.go` | プロンプト修正（分解しない指示） |
| `ios/Uchikomi/Features/Meals/Views/FoodItemEditRow.swift` | 「食材」→「メニュー」、プレースホルダー変更 |
| `ios/Uchikomi/Features/Meals/MealInputView.swift` | 「食品を追加」→「メニューを追加」、「検出された食品」→「検出されたメニュー」 |
| `ios/Uchikomi/Features/Meals/Views/NutritionEditorView.swift` | 「食材を追加」→「メニューを追加」 |

---

## 検証方法

1. バックエンドを再起動
2. iOSアプリをビルド
3. 以下のテストケースで動作確認:
   - 「ラーメン」と入力 → 「ラーメン」として1つの料理で返される（分解されない）
   - 「醤油ラーメン 1杯」と入力 → そのまま返される
   - 「幕の内弁当」と入力 → 分解されず1つの料理として返される
   - UI上の表記が「メニュー」になっていることを確認
