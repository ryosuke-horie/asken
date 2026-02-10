# プラン: WeightViewで初期ローディング状態をチェック

## Linear Issue
- Issue: EDG-702
- URL: https://linear.app/ryosuke-horie/issue/EDG-702

## 概要

食事タブから体重タブに移動した際、一瞬「縦の線」が表示される問題を修正する。
WeightViewで`isInitialLoading`をチェックして、初期ロード中は画面全体をローディング表示にする。

## 問題の原因

1. WeightViewでは`viewModel.isLoading`のみチェックしている
2. `isInitialLoading = true`で初期化されているが、WeightViewではチェックされていない
3. 初期ロード時に一瞬`isLoading = false`になるタイミングでScrollViewが表示される
4. 空のデータでWeightChartViewが描画され、何らかの線が表示される

## 実装計画

### WeightViewの変更

**条件分岐の修正:**

現在:
```swift
if viewModel.isLoading {
    // ProgressView
} else if let error = viewModel.errorMessage {
    // ErrorView
} else {
    // ScrollView
}
```

修正後:
```swift
if viewModel.isInitialLoading || viewModel.isLoading {
    // ProgressView
} else if let error = viewModel.errorMessage {
    // ErrorView
} else {
    // ScrollView
}
```

## 変更ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Features/Weight/WeightView.swift` | `isInitialLoading`チェックを追加 |

## エッジケース対応

| ケース | 対応 |
|:---|:---|
| 初期ロード中 | ProgressViewを表示 |
| 期間変更時のローディング | ProgressViewを表示 |
| エラー発生時 | ErrorViewを表示（優先） |

## 検証方法

1. アプリを起動し食事タブを選択
2. 体重タブに移動
3. 縦の線が表示されないことを確認
4. 期間セグメントの切り替えが正常に動作することを確認
