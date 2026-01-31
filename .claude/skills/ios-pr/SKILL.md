---
name: ios-pr
description: iOSアプリのPR作成前にユニットテストを実行。テスト失敗時はPR作成をブロック。
model: sonnet
allowed-tools: Bash, Read, Glob, Write
---

# iOS PR作成スキル

## 実行手順

1. 変更確認: `ios/` に変更があるか確認
2. ユニットテスト実行（Swift Testing + スナップショットテスト）
3. テスト成功時のみPR作成

## コマンド

```bash
# ユニットテスト（Taskfileを使用）
task ios:test
```

## 結果

- 成功: PR作成に進む
- 失敗: 失敗したテストを報告し、PR作成をブロック

## 備考

- UIテスト（XCUITest）は現在未実装
- スナップショットテストはユニットテストに含まれる
