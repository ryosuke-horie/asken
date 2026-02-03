---
name: ios-app-review
description: iOS App Reviewガイドラインに基づき、ios/ と backend/ の実装・設定・メタデータをコードレビューして、クリティカルな違反はLinearのUtikomiプロジェクトに日本語Issue作成し、確認・検討事項はコメントで整理する時に使う。
---

# iOS App Review

## Overview
App Store審査リスクを公式ガイドラインと実装証拠で評価し、重大事項はLinearに起票する。

## Workflow
1. 対象の確認: `ios/` と `backend/` を対象に、設定ファイル・認証・課金・データ収集の実装を確認する。
2. シグナル収集: `scripts/scan_review_signals.py` を実行し、出力を読む。
3. ガイドライン照合: `references/guidelines.md` と `references/privacy-manifest.md` を参照し、該当条項と証拠を紐づける。
4. 重要度判定: CRITICAL か REVIEW に分類する。
5. 出力: CRITICAL は Linear Issue 作成、REVIEW はコメントで提示する。

## Signal Collection
- 実行コマンド: `python3 skills/ios-app-review/scripts/scan_review_signals.py --out reports/ios-review-signals.json`
- 追加で確認するファイル: `ios/**/Info.plist`, `ios/**/PrivacyInfo.xcprivacy`, `ios/**/*.entitlements`, StoreKit利用箇所, 認証SDK, 広告/計測SDK, WebView実装, バックエンドの課金・削除・データ収集フロー

## Triage Criteria
CRITICAL:
- 公開API以外の使用、UIWebViewなど非推奨/禁止APIの利用疑い
- アプリ外課金/デジタルコンテンツ課金の回避疑い
- サードパーティログイン提供時に Sign in with Apple が無い
- Privacy manifest が無い/不正
- 収集データの同意やプライバシーポリシーが欠けている疑い

REVIEW:
- ガイドライン適用の解釈が必要
- サーバー側フロー次第で違反/適合が分かれる
- 証拠不足で追加確認が必要

## Linear Issue Template (CRITICAL only)
Title: `[App Review][CRITICAL] <短い要約>`

Body:
- ガイドライン: `2.5.1` など
- 影響: 申請リジェクトの可能性
- 根拠: `path:line` の証拠
- 背景: 実装や設定の説明
- 修正案: 実装・設定の改善案

## Output Format
- Summary: 件数と重要ポイント
- Critical Issues: 作成したLinear Issueの一覧
- Review/Check Items: コメントでの指摘一覧
- Needs Info: 追加確認が必要な事項

## Resources
- `references/guidelines.md`
- `references/privacy-manifest.md`
- `scripts/scan_review_signals.py`
