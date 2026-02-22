# ウチコミ コードマップ INDEX

最終更新: 2026-02-22

## 概要

ウチコミは格闘技の減量・体重コントロール支援iOSアプリケーション。
GoバックエンドとSwiftUI iOSアプリで構成され、Gemini APIによるAI食事分析・メニューサジェスト機能を持つ。

## コードマップ一覧

| ドキュメント | 内容 | 主な対象 |
|:---|:---|:---|
| [architecture.md](./architecture.md) | 全体アーキテクチャ | システム構成、データフロー、インフラ |
| [backend.md](./backend.md) | バックエンド構造 | Go API、ハンドラ、リポジトリ、Gemini連携 |
| [ios.md](./ios.md) | iOSアプリ構造 | SwiftUI画面、MVVM、認証 |
| [data.md](./data.md) | データモデル | Firestoreコレクション、スキーマ、インデックス |

## 主要機能エリア

| 機能 | バックエンド | iOS | データ |
|:---|:---|:---|:---|
| 食事分析 | analyze/status/history handler | Meals feature | analysisRequests |
| 体重管理 | weight handler | Weight feature | weightRecords/weightGoal |
| 栄養目標 | nutrition_goal handler | NutritionGoalSettingView | nutritionGoal |
| マイメニュー | my_menu handler | MyMenu feature | myMenu |
| 食材管理 | ingredient/scan_receipt handler | Pantry feature | ingredients |
| メニューサジェスト | menu_suggestion handler | CookingSuggestion feature | menuSuggestions |
| 認証 | auth/dev_auth middleware | Auth feature (UchikomiCore) | Firebase Auth |
| 通知 | - | Notification (ローカル) | UserDefaults |
