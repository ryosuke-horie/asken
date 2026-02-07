# ウチコミ プロジェクトガイドライン

このドキュメントは、Claude Codeがこのプロジェクトで作業する際のガイドラインを定義します。

## 言語

- エージェントメモリ（MEMORY.md）を含む全ての出力は日本語で記述すること

## 詳細規約

詳細なコーディング規約、ベストプラクティス、セキュリティガイドライン等は `.claude/rules/` ディレクトリを参照してください。

### 基本ルール

| ファイル | 内容 |
|:---|:---|
| `00-common.md` | 共通ルール |
| `01-testing-details.md` | テスト詳細（古典派スタイル、モック基準） |
| `02-git-workflow-details.md` | Git/GitHubワークフロー詳細（Linear連携、PRレビュー、マージ戦略） |
| `03-coding-style.md` | コーディングスタイル（KISS, DRY, YAGNI） |
| `04-security-details.md` | セキュリティ詳細（コミット前チェック） |
| `05-performance.md` | パフォーマンス最適化（モデル選択、コンテキスト管理） |
| `06-patterns.md` | 共通パターン（APIレスポンス形式、Repositoryパターン） |
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
| `17-ios-build-troubleshooting.md` | iOSビルドトラブル対応（SPMキャッシュ、Firebase SDK） |

### Claude Code拡張

| ファイル | 内容 |
|:---|:---|
| `31-hooks.md` | フックシステム（PreToolUse, PostToolUse, Stop） |
| `33-plan-mode.md` | Plan Modeルール（プラン保存、Linear連携） |

### 即座に使用すべきエージェント

ユーザーからの明示的な指示なしでプロアクティブに使用すること:

| トリガー | エージェント |
|:---|:---|
| 複雑な機能リクエスト | planner |
| コード作成/修正後 | code-reviewer |
| バグ修正または新機能 | tdd-guide |

利用可能なエージェント一覧は `.claude/agents/` を参照。

---

## 関連ドキュメント

| ドキュメント | 内容 |
|:---|:---|
| [docs/CONTRIB.md](./docs/CONTRIB.md) | 開発者ガイド（環境セットアップ、コマンド一覧） |
| [docs/RUNBOOK.md](./docs/RUNBOOK.md) | 運用手順書 |
| [docs/CODEMAPS/](./docs/CODEMAPS/) | コードマップ |
| [docs/adr/](./docs/adr/) | アーキテクチャ決定記録 |

