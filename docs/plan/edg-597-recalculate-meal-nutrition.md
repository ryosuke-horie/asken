# プラン: 食事記録の栄養素再計算機能

## Linear Issue

- Issue: EDG-597
- URL: https://linear.app/ryosuke-horie/issue/EDG-597

## 概要

食事記録編集時に、杯数変更の場合は算術計算、メニュー名変更の場合はLLM（Gemini API）を使って栄養素を再計算する機能を実装する。

## 要件

1. **杯数変更のみの場合**
   - LLMを使わず単純な算術計算で対応
   - 例: 1杯 → 2杯なら、全ての栄養素を2倍
   - 高速なレスポンスを維持（UX重視）

2. **メニュー自体を変更した場合**
   - LLMで標準的な栄養素・カロリーを取得
   - 例: 「家系ラーメン」→「味噌ラーメン」に変更

3. **データ保存**
   - 杯数（serving_count）をバックエンドに保存する
   - 杯数に応じて計算された栄養素も保存する
   - 履歴読み込み時に杯数を復元し、基準値を逆算する

## 実装アプローチ

**iOS/バックエンド分散型**を採用

- 杯数変更: iOS側で即時計算（ネットワーク不要）
- メニュー名変更: バックエンドAPI経由でLLM呼び出し

### 処理フロー

```
[iOS側]
ユーザーがメニューを編集
  ├─ 杯数のみ変更 → iOS側で算術計算 → 即時UI更新
  └─ メニュー名変更 → バックエンドAPI呼び出し

[バックエンド側]
POST /api/nutrition/estimate
  ├─ リクエスト: { "food_name": "味噌ラーメン", "quantity": 1 }
  └─ レスポンス: { 栄養素情報 }
```

## 技術的な決定事項

### 杯数の管理

新規`quantity`フィールドを追加する（`estimatedAmount`のパースは複雑なため避ける）

- `FoodEditItem`に`quantity: Int`を追加
- `baseNutrition`（1単位あたりの栄養素）を保持
- 計算式: `新しい栄養素 = baseNutrition × quantity`

### API設計

```
POST /api/nutrition/estimate

Request:
{
  "food_name": "味噌ラーメン",
  "quantity": 1
}

Response:
{
  "name": "味噌ラーメン",
  "estimated_amount": "1杯",
  "calories_kcal": 500,
  "protein_g": 20,
  "fat_g": 15,
  "carbohydrates_g": 60
}
```

## 実装ステップ

### フェーズ1: バックエンドAPI追加

1. `POST /api/nutrition/estimate`エンドポイント作成
2. NutritionCalculatorに単品検索メソッド追加
3. ルーティング追加

### フェーズ2: iOS側モデル拡張

1. FoodEditItemに`quantity`, `baseNutrition`プロパティ追加
2. 杯数変更の計算ロジック追加

### フェーズ3: iOS側リポジトリ拡張

1. APIエンドポイント追加
2. MealRepositoryProtocolにメソッド追加

### フェーズ4: iOS側ViewModel拡張

1. NutritionEditorViewModelに再計算ロジック追加
   - `onQuantityChanged` - 即時計算
   - `onNameChanged` - API呼び出し

### フェーズ5: iOS側View更新

1. 杯数入力UI追加
2. ローディング状態の表示

## スコープ外（後続タスク）

- クライアント側キャッシュ
- 詳細なエラーハンドリング
- オフライン対応

## テスト計画

### バックエンド

- NutritionEstimateHandler: リクエストバリデーション、レスポンス形式
- EstimateSingleFood: 正常系、空文字

### iOS

- FoodEditItem: 杯数変更計算の正確性
- NutritionEditorViewModel: 変更検出、API呼び出しタイミング

## 成功基準

- 杯数変更時、即時にUI反映（ネットワーク呼び出しなし）
- メニュー名変更時、LLMから栄養素を取得して表示
- 保存時、杯数と計算後の栄養素がバックエンドに保存される
- 履歴読み込み時、杯数が復元され、さらに杯数を変更できる
