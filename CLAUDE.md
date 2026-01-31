# ウチコミ プロジェクトガイドライン

このドキュメントは、Claude Codeがこのプロジェクトで作業する際のガイドラインを定義します。

## プロジェクト概要

**ウチコミ**は、柔術/キックボクシングなど格闘技の減量・体重コントロールを支援する個人用アプリケーションです。日々の記録（体重、食事、体調、疲労度）とAI相談で合意形成しながら進めるツールを目指します。Gemini API（Gemini 3）を活用してMVP（Minimum Viable Product）を構築します。

### 主要機能

- 🔐 **ユーザー認証**: メールアドレス/パスワードによるログイン認証
- 📸 **画像認識**: 食事の画像をアップロードして食材を自動判定
- 🔍 **カロリー検索**: 食品名からカロリーと栄養素を検索
- 📊 **栄養素計算**: タンパク質、脂質、炭水化物などの栄養素を計算
- 💾 **データ管理**: 食品データベース（PostgreSQL）とGemini APIを組み合わせた高精度な判定

### 技術スタック

- **iOS**: Swift / SwiftUI
- **バックエンド**: Golang
- **AI**: Gemini CLI（Gemini 3 API）via シェルコマンド
- **データベース**: PostgreSQL
- **ホスティング**: exe.dev（Ubuntu環境）
- **バージョン管理**: Git

### 対象ユーザー

個人利用（開発者のみ）

---

## 詳細規約

詳細なコーディング規約、ベストプラクティス、セキュリティガイドライン等は `.claude/rules/` ディレクトリを参照してください。

- **iOS**: `.claude/rules/10-ios-testing.md`
- **バックエンド**: `.claude/rules/backend-golang.md`
- **データベース**: `.claude/rules/database.md`
- **Gemini API**: `.claude/rules/gemini-api.md`
- **テスト**: `.claude/rules/testing.md`
- **セキュリティ**: `.claude/rules/security.md`
- **Git運用**: `.claude/rules/git-workflow.md`
- **Claude Code指示**: `.claude/rules/claude-instructions.md`
