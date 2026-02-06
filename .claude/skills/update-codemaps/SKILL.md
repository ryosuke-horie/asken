---
name: update-codemaps
description: コードベースの構造を分析し、アーキテクチャドキュメント（docs/CODEMAPS/）を更新。
---

# コードマップ更新スキル

コードベースの構造を分析し、アーキテクチャドキュメントを更新します。

## 実行手順

1. 全ソースファイルのインポート、エクスポート、依存関係をスキャン
2. 以下の形式でトークン軽量なコードマップを生成:
   - codemaps/architecture.md - 全体アーキテクチャ
   - codemaps/backend.md - バックエンド構造
   - codemaps/ios.md - iOSアプリ構造
   - codemaps/data.md - データモデルとスキーマ

3. 前バージョンからの差分率を計算
4. 変更が30%を超える場合、更新前にユーザー承認を要求
5. 各コードマップに鮮度タイムスタンプを追加
6. レポートを.reports/codemap-diff.txtに保存

実装詳細ではなく高レベル構造に焦点。

## プロジェクト固有のコードマップ構造

バックエンド（Go）:
- Handler → Service → Repository → Firestore
- 外部サービス連携（Gemini HTTP API）

iOS（SwiftUI）:
- MVVM + Repository
- View → ViewModel → Repository → APIClient
- 共通コンポーネント・拡張
