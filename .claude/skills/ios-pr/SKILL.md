---
name: ios-pr
description: |
  iOSアプリのPR作成前にUIテストとユニットテストを実行。
  テスト失敗時はPR作成をブロック。
  ios/ディレクトリに変更がある場合に使用。
allowed-tools: Bash(xcodebuild *), Bash(gh *), Read, Glob
---

# iOS PR作成スキル

## 実行手順

1. **変更確認**: `ios/` に変更があるか確認
2. **ユニットテスト実行**
3. **UIテスト実行**
4. **テスト成功時のみPR作成**

## コマンド

```bash
# ユニットテスト
cd ios && xcodebuild test -scheme Uchikomi -destination 'platform=iOS Simulator,name=iPhone 15'

# UIテスト
cd ios && xcodebuild test -scheme UchikomiUITests -destination 'platform=iOS Simulator,name=iPhone 15'
```

## 結果

- **成功**: PR作成に進む
- **失敗**: 失敗したテストを報告し、PR作成をブロック
