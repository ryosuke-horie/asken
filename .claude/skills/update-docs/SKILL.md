---
name: update-docs
description: 正確な情報源からドキュメント全体を同期・更新。Taskfile.yml、.env.example、ソースコード構造を元に各ドキュメントの整合性を確認。
---

# ドキュメント更新スキル

正確な情報源（source of truth）からドキュメントを同期します。

## 対象ドキュメント一覧

| ドキュメント | 情報源 | 確認内容 |
|:---|:---|:---|
| README.md | ソースコード全体 | 主要機能、技術スタック、セットアップ手順 |
| CLAUDE.md | .claude/rules/, .claude/agents/ | ルール一覧テーブルの整合性 |
| docs/CONTRIB.md | Taskfile.yml, .env.example | コマンド一覧、環境セットアップ、テスト手順 |
| docs/RUNBOOK.md | infrastructure/, Taskfile.yml | デプロイ手順、トラブルシューティング |
| ios/README.md | ios/project.yml, ios/ | ビルド手順、依存関係、セットアップ |
| infrastructure/README.md | infrastructure/ | インフラ構成、デプロイ手順 |
| docs/setup/GOOGLE_SIGNIN_SETUP.md | ios/, backend/ | Google Sign-In設定手順 |
| .claude/rules/08-development-commands.md | Taskfile.yml | Taskfileコマンド一覧の同期 |

対象外（別スキルまたは更新不要）:
- docs/CODEMAPS/ → `update-codemaps` スキルが担当
- docs/adr/ → 履歴記録のため更新対象外
- docs/plan/ → 実装計画の一時ドキュメント

## 実行手順

### 1. 情報源の読み取り

- Taskfile.ymlのタスク定義を読む
- 全.env.exampleファイルを読む
- .claude/rules/ と .claude/agents/ のファイル一覧を取得
- ios/project.yml を読む（iOS関連の場合）

### 2. 各ドキュメントの同期

docs/CONTRIB.md:
- 利用可能なTaskfileコマンド
- 環境セットアップ手順
- テスト実行方法

docs/RUNBOOK.md:
- デプロイ手順
- よくある問題と修正
- ロールバック手順

README.md:
- 主要機能リスト
- 技術スタック
- クイックスタート手順

CLAUDE.md:
- rules/一覧テーブルとファイルの整合性
- agents/一覧テーブルとファイルの整合性

.claude/rules/08-development-commands.md:
- Taskfile.ymlのタスクと一致しているか
- コマンド説明が正確か

ios/README.md:
- ビルド手順
- 依存関係情報

infrastructure/README.md:
- GCPリソース構成
- デプロイ手順

### 3. 古いドキュメントの検出

- 90日以上更新されていないドキュメントを検出
- 手動レビュー用にリスト化

### 4. 差分サマリーを表示

変更があったファイルと変更内容の要約を表示。
