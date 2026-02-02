# ウチコミ プロジェクトガイドライン

このドキュメントは、Claude Codeがこのプロジェクトで作業する際のガイドラインを定義します。

## プロジェクト概要

ウチコミは、柔術/キックボクシングなど格闘技の減量・体重コントロールを支援する個人用アプリケーションです。日々の記録（体重、食事、体調、疲労度）とAI相談で合意形成しながら進めるツールを目指します。Gemini API（Gemini 3）を活用してMVP（Minimum Viable Product）を構築します。

### 主要機能

- 🔐 ユーザー認証: メールアドレス/パスワードによるログイン認証
- 📸 画像認識: 食事の画像をアップロードして食材を自動判定
- 🔍 カロリー検索: 食品名からカロリーと栄養素を検索
- 📊 栄養素計算: タンパク質、脂質、炭水化物などの栄養素を計算
- 💾 データ管理: Gemini APIを活用した食事分析

### 技術スタック

- iOS: Swift / SwiftUI
- バックエンド: Golang
- AI: Gemini CLI（Gemini 3 API）via シェルコマンド
- ホスティング: exe.dev（Ubuntu環境）
- バージョン管理: Git

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

### 技術別ルール

| ファイル | 内容 |
|:---|:---|
| `10-ios-testing.md` | iOSテストガイドライン（Swift Testing, Mockolo） |
| `backend-golang.md` | バックエンド開発規約（Go） |
| `database.md` | データベース規約 |
| `gemini-api.md` | Gemini API連携のベストプラクティス |
| `testing.md` | テスト駆動開発（TDD） |
| `security.md` | セキュリティガイドライン |
| `git-workflow.md` | Git運用とコードレビュー |
| `claude-instructions.md` | Claude Codeへの指示 |

### Claude Code拡張

| ファイル | 内容 |
|:---|:---|
| `40-hooks.md` | フックシステム（PreToolUse, PostToolUse, Stop） |
| `41-agents.md` | エージェントオーケストレーション |
