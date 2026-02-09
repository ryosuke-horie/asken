---
name: ios-pr
description: iOSアプリのPR作成前のチェック。現在iOSテストは一時無効化されているため、スキップされます。
model: sonnet
allowed-tools: Bash, Read, Glob, Write
---

# iOS PR作成スキル

> `注意: iOSテストは一時無効化されています`
>
> macOS/Xcode バージョン問題により、iOS テストを一時的に無効化しています。
> 詳細は `.claude/rules/ios-testing-policy.md` を参照してください。

## 実行手順

1. 変更確認: `ios/` に変更があるか確認
2. iOSテストはスキップ（無効化中）

## 結果

- 常に成功とみなし、PR作成に進む
