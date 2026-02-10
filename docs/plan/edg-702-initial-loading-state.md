# プラン: 体重記録画面の初期ローディング状態追加

## Linear Issue
- Issue: EDG-702
- URL: https://linear.app/ryosuke-horie/issue/EDG-702

## 概要

食事タブから体重タブに移動した際、一瞬「縦の線」（RuleMark）が単独で表示される問題を修正する。
初期ローディング状態を追加して、最初のデータ読み込み完了までチャート全体を非表示にする。

## 問題の原因

1. WeightViewModel初期化時: `isLoading = false`, `chartRecords = []`
2. タブ移動時に`.task`が実行されるが、最初の描画タイミングではデータが空
3. `goal`が先に設定されると、空のチャートに`RuleMark`（目標線）だけが描画される

## 実装計画

### 1. WeightViewModelの変更

**isInitialLoadingプロパティの追加:**
- 初期値は `true`
- 最初の `loadData()` 完了後に `false` にする
- 期間変更時の `loadChartData()` では変更しない

### 2. WeightChartViewの変更

**isInitialLoading引数の追加:**
- `isInitialLoading` が `true` の場合はプレースホルダー（ProgressView）を表示
- 従来の `isLoading` は期間変更時のローディングに使用

### 3. WeightViewの変更

**WeightChartViewへの引数追加:**
- `isInitialLoading: viewModel.isInitialLoading` を渡す

## 変更ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Features/Weight/WeightViewModel.swift` | `isInitialLoading` プロパティ追加、ローディングメソッド修正 |
| `ios/Uchikomi/Features/Weight/Views/WeightChartView.swift` | `isInitialLoading` 引数追加、表示ロジック修正 |
| `ios/Uchikomi/Features/Weight/WeightView.swift` | `isInitialLoading` 引数を渡す |

## エッジケース対応

| ケース | 対応 |
|:---|:---|
| エラー発生時 | `isInitialLoading` は `false` にならず、エラービュー表示 |
| 期間変更時 | `isInitialLoading` は変更せず、`isLoading` のみ使用 |
| データが空の場合 | `isInitialLoading` が `false` で「データがありません」表示 |

## 検証方法

1. アプリを起動し食事タブを選択
2. 体重タブに移動
3. 縦の線（目標線）が一瞬表示されないことを確認
4. 期間セグメントの切り替えが正常に動作することを確認
