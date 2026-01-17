# asken プロジェクトガイドライン

このドキュメントは、Claude Codeがこのプロジェクトで作業する際のガイドラインを定義します。

## プロジェクト概要

**asken**は、画像から食事内容を判定し、カロリーや栄養素を計算する個人用カロリー計算アプリケーションです。Gemini API（Gemini 3）を活用してMVP（Minimum Viable Product）を構築します。

### 主要機能

- 📸 **画像認識**: 食事の画像をアップロードして食材を自動判定
- 🔍 **カロリー検索**: 食品名からカロリーと栄養素を検索
- 📊 **栄養素計算**: タンパク質、脂質、炭水化物などの栄養素を計算
- 💾 **データ管理**: 食品データベース（PostgreSQL）とGemini APIを組み合わせた高精度な判定

### 技術スタック

- **フロントエンド**: Next.js (React + TypeScript)
- **バックエンド**: Golang
- **AI**: Gemini CLI（Gemini 3 API）via シェルコマンド
- **データベース**: PostgreSQL
- **ホスティング**: exe.dev（Ubuntu環境）
- **バージョン管理**: Git

### 対象ユーザー

個人利用（開発者のみ）- マルチユーザー対応は不要

---

## 詳細規約

詳細なコーディング規約、ベストプラクティス、セキュリティガイドライン等は `.claude/rules/` ディレクトリを参照してください。

- **フロントエンド**: `.claude/rules/frontend-nextjs.md`
- **バックエンド**: `.claude/rules/backend-golang.md`
- **データベース**: `.claude/rules/database.md`
- **Gemini API**: `.claude/rules/gemini-api.md`
- **テスト**: `.claude/rules/testing.md`
- **セキュリティ**: `.claude/rules/security.md`
- **Git運用**: `.claude/rules/git-workflow.md`
- **Claude Code指示**: `.claude/rules/claude-instructions.md`
