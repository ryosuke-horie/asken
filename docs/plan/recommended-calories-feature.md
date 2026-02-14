# プラン: 目標カロリー推奨値計算機能

## 概要
ユーザーの属性（性別・年齢・身長・活動レベル）から推奨カロリーを計算して、NutritionGoalSettingViewで表示する機能を追加。

## 要件
- 性別・年齢・身長・活動レベルを入力できるUI
- 推奨カロリーを計算して表示
- 計算式は改訂Harris-Benedict式を使用

## 実装計画

### 1. ユーザー属性モデル
- `Gender` / `ActivityLevel` 列挙型を作成
  - 性別: male/female
  - 活動レベル: sedentary/lightlyActive/moderatelyActive/veryActive/athlete

### 2. 推奨カロリー計算ロジック
- 改訂Harris-Benedict式
  - 男性: 88.362 + 13.397 x 体重(kg) + 4.799 x 身長(cm) - 5.677 x 年齢
  - 女性: 447.593 + 9.247 x 体重(kg) + 3.098 x 身長(cm) - 4.330 x 年齢
- 活動レベルによる補正
  - sedentary: x 1.2
  - lightlyActive: x 1.375
  - moderatelyActive: x 1.55
  - veryActive: x 1.725
  - athlete: x 1.9

### 3. UI追加
- `NutritionGoalSettingView` に推奨値計算セクションを追加
- 計算結果の「採用」ボタンで目標カロリーに反映

### 4. データ永続化（将来的）
- ユーザー属性をFirestoreに保存

## 技術的な考慮事項
- 現在のPR #187に追加実装
- iOSのみの実装でバックエンド変更なし
- 計算ロジックはiOS側で実装
