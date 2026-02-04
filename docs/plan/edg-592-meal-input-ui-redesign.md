# プラン: 食事入力UIの再設計

## Linear Issue
- Issue: EDG-592
- URL: https://linear.app/ryosuke-horie/issue/EDG-592

## 概要

現在の「モード選択 → 入力」の2ステップUIを廃止し、最初から入力フィールドが表示されるUIに変更する。

## ユーザー要望

1. 最初から入力フィールドが表示されている（モード選択なし）
2. 1行 = 1メニュー（食品名 + 量）で複数行入力できる
3. 画像選択ボタンも同じ画面にある
4. 量は自由テキスト（「100g」「1個」「お茶碗1杯」など）
5. 画像認識結果 + 手動追加を合わせて使いたい（併用可能）

---

## 新しいUI設計

```
+------------------------------------------+
| [閉じる]     昼食を記録                  |
+------------------------------------------+
|                                          |
|  --- 登録済みの記録 ---                  |
|  [既存の食事カード...]                   |
|                                          |
|  --- 新しく記録する ---                  |
|                                          |
|  +--------------------------------------+|
|  | 食材                            [🗑] ||
|  | 料理名: [鶏むね肉            ]       ||
|  | 量:    [100g                 ]       ||
|  +--------------------------------------+|
|  +--------------------------------------+|
|  | 食材                            [🗑] ||
|  | 料理名: [ご飯                ]       ||
|  | 量:    [1杯                  ]       ||
|  +--------------------------------------+|
|                                          |
|  [+ 食品を追加]                          |
|                                          |
|  --- または画像から入力 ---              |
|  +--------------------------------------+|
|  | [写真を選択] [カメラ]                ||
|  | (選択した画像プレビュー)             ||
|  +--------------------------------------+|
|                                          |
|  [分析する]                              |
|                                          |
+------------------------------------------+
```

---

## 技術的制約

- バックエンドAPIはテキストと画像の同時送信をサポートしていない
- テキスト入力は`input_text`フィールドに複数食品をカンマ区切りで記載する形式

---

## 実装計画

### Step 1: MealInputViewModel の修正

1. `manualFoods: [FoodEditItem]`プロパティを追加（初期値は空の1行）
2. `addManualFood()`, `removeManualFood(_:)`を実装
3. 手動入力をテキスト形式に変換する関数を追加

```swift
// 変換例: "鶏むね肉 100g, ご飯 1杯"
private func buildInputText(from foods: [FoodEditItem]) -> String {
    foods
        .filter { !$0.name.isEmpty }
        .map { food in
            food.quantity.isEmpty ? food.name : "\(food.name) \(food.quantity)"
        }
        .joined(separator: ", ")
}
```

### Step 2: MealInputView のUI変更

1. `InputMode` enumを廃止
2. `InputModeSelectionSection`を廃止
3. `TextInputSection`を`ManualFoodInputSection`に置き換え
   - 既存の`FoodItemEditRow`を再利用
   - 「+ 食品を追加」ボタン
4. `ImageSelectionSection`は維持（同じ画面に配置）
5. 分析ボタンを1つに統合

### Step 3: 分析ロジックの統合

「分析する」ボタン押下時の動作:

| 画像 | 手動入力 | 処理 |
|:--:|:--:|:--|
| あり | - | 画像分析API → NutritionEditor |
| なし | あり | テキスト分析API → NutritionEditor |
| なし | なし | エラー表示 |

画像と手動入力の併用:
- 画像分析後、NutritionEditorで手動入力分を追加可能
- 手動入力した食品は`manualFoods`として保持し、NutritionEditorに渡す

### Step 4: テストの追加

- 手動入力の追加/削除
- テキスト変換ロジック
- 空入力バリデーション

---

## 変更対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Features/Meals/MealInputViewModel.swift` | `manualFoods`プロパティ追加、テキスト変換ロジック |
| `ios/Uchikomi/Features/Meals/MealInputView.swift` | モード選択廃止、複数行入力UI実装 |
| `ios/UchikomiTests/Features/Meals/MealInputViewModelTests.swift` | 新しいテストケース追加 |

## 再利用するコンポーネント

| ファイル | 用途 |
|:---|:---|
| `FoodEditItem.swift` | 食品データモデル（name, quantity） |
| `FoodItemEditRow.swift` | 食品入力行UI |

---

## 検証方法

1. `task ios:lint` でリントエラーがないことを確認
2. `xcodebuild build` でビルドが成功することを確認
3. ユーザーによる手動検証:
   - 最初から入力フィールドが表示されること
   - 食品を複数行追加・削除できること
   - 手動入力のみで分析が成功すること
   - 画像選択のみで分析が成功すること
   - 分析結果がNutritionEditorに表示されること
