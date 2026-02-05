# ウチコミ プロジェクトガイドライン

このドキュメントは、Claude Codeがこのプロジェクトで作業する際のガイドラインを定義します。

## プロジェクト概要

ウチコミは、柔術/キックボクシングなど格闘技の減量・体重コントロールを支援する個人用アプリケーションです。日々の記録（体重、食事、体調、疲労度）とAI相談で合意形成しながら進めるツールを目指します。Gemini API（Gemini 3）を活用して構築します。

### 主要機能

- ユーザー認証: メールアドレス/パスワードによるログイン認証
- 画像認識: 食事の画像をアップロードして食材を自動判定
- カロリー検索: 食品名からカロリーと栄養素を検索
- 栄養素計算: タンパク質、脂質、炭水化物などの栄養素を計算
- AIエージェント: 減量サポートを提供

### 技術スタック

- iOS: Swift / SwiftUI
- バックエンド: Golang
- AI: Gemini API（将来的にLangChain等でAIエージェント自作予定）

### 対象ユーザー

個人利用（開発者のみ）

---

## 詳細規約

詳細なコーディング規約、ベストプラクティス、セキュリティガイドライン等は `.claude/rules/` ディレクトリを参照してください。

### 基本ルール

| ファイル | 内容 |
|:---|:---|
| `00-common.md` | 共通ルール |
| `01-testing-details.md` | テスト詳細（古典派スタイル、モック基準） |
| `02-git-workflow-details.md` | Git/GitHubワークフロー詳細（Linear連携） |
| `03-coding-style.md` | コーディングスタイル（KISS, DRY, YAGNI） |
| `04-security-details.md` | セキュリティ詳細（コミット前チェック） |
| `05-performance.md` | パフォーマンス最適化（モデル選択、コンテキスト管理） |
| `06-patterns.md` | 共通パターン（Result型、APIレスポンス形式） |
| `07-dependencies.md` | 依存関係管理（dependabot設定） |
| `08-development-commands.md` | 開発コマンド（Taskfile） |
| `09-documentation.md` | ドキュメント更新ルール |

### 技術別ルール

| ファイル | 内容 |
|:---|:---|
| `10-ios-testing.md` | iOSテストガイドライン（Swift Testing, Mockolo） |
| `11-backend-golang.md` | バックエンド開発規約（Go） |
| `12-gemini-api.md` | Gemini API連携のベストプラクティス |
| `13-firestore.md` | Firestore規約（インデックス管理） |
| `14-testing-tdd.md` | テスト駆動開発（TDD） |
| `15-security-practices.md` | セキュリティガイドライン |
| `16-git-workflow-general.md` | Git運用とコードレビュー |
| `17-ios-build-troubleshooting.md` | iOSビルドトラブル対応（SPMキャッシュ、Firebase SDK） |

### Claude Code拡張

| ファイル | 内容 |
|:---|:---|
| `30-claude-instructions.md` | Claude Codeへの指示 |
| `31-hooks.md` | フックシステム（PreToolUse, PostToolUse, Stop） |
| `32-agents.md` | エージェントオーケストレーション |
| `33-plan-mode.md` | Plan Modeルール（プラン保存、Linear連携） |

---

## 関連ドキュメント

| ドキュメント | 内容 |
|:---|:---|
| [docs/CONTRIB.md](./docs/CONTRIB.md) | 開発者ガイド（環境セットアップ、コマンド一覧） |
| [docs/RUNBOOK.md](./docs/RUNBOOK.md) | 運用手順書 |
| [docs/CODEMAPS/](./docs/CODEMAPS/) | コードマップ |
| [docs/adr/](./docs/adr/) | アーキテクチャ決定記録 |

