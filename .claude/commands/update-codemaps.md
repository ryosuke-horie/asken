# コードマップ更新

コードベースの構造を分析し、アーキテクチャドキュメントを更新:

1. 全ソースファイルのインポート、エクスポート、依存関係をスキャン
2. 以下の形式でトークン軽量なコードマップを生成:
   - codemaps/architecture.md - 全体アーキテクチャ
   - codemaps/backend.md - バックエンド構造
   - codemaps/frontend.md - フロントエンド構造
   - codemaps/data.md - データモデルとスキーマ

3. 前バージョンからの差分率を計算
4. 変更が30%を超える場合、更新前にユーザー承認を要求
5. 各コードマップに鮮度タイムスタンプを追加
6. レポートを.reports/codemap-diff.txtに保存

TypeScript/Node.jsを使用して分析。実装詳細ではなく高レベル構造に焦点。

## プロジェクト固有のコードマップ構造

バックエンド:
- Golang構造
- Handler → Service → Repository → Firestore
- 外部サービス連携（Gemini API）

フロントエンド:
- Next.js App Router構造
- Server Components / Client Components
- フック・状態管理の依存関係

共有:
- 型定義
- ユーティリティ
