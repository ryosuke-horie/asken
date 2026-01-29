# 全体アーキテクチャ

最終更新: 2026-01-30

## システム概要

ウチコミは格闘技の減量・体重コントロール支援アプリケーション。日々の記録（体重、食事、体調、トレーニング）とAI相談で減量を支援する。

## 技術スタック

| レイヤー | 技術 |
|:---|:---|
| フロントエンド | Next.js 15 (App Router), TypeScript, SWR |
| バックエンド | Golang (標準ライブラリ) |
| iOSアプリ | Swift, SwiftUI, WidgetKit |
| データベース | PostgreSQL |
| AI | Gemini CLI (gemini-3-flash-preview) |
| ホスティング | exe.dev (Ubuntu) |

## システム構成図

```
┌─────────────────────────────────────────────────────────────────┐
│                         クライアント                              │
├─────────────────────┬─────────────────────┬─────────────────────┤
│    Next.js Web      │      iOS App        │    iOS Widget       │
│    (localhost:3000) │   (SwiftUI)         │   (WidgetKit)       │
└──────────┬──────────┴──────────┬──────────┴──────────┬──────────┘
           │                     │                     │
           └──────────────┬──────┴─────────────────────┘
                          │ HTTP/REST
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    バックエンド (Go :8080)                        │
├─────────────────────────────────────────────────────────────────┤
│  Middleware (CORS, Auth)                                        │
│    ├── Auth認証 (JWT)                                           │
│    └── CORSヘッダー                                              │
├─────────────────────────────────────────────────────────────────┤
│  Handler層                                                      │
│    ├── AuthHandler      (認証)                                  │
│    ├── AnalyzeHandler   (食事分析)                              │
│    ├── DailyMealsHandler(日次食事)                              │
│    ├── WeightHandler    (体重管理)                              │
│    ├── ConditionHandler (体調記録)                              │
│    ├── TrainingHandler  (トレーニング)                          │
│    ├── MylistHandler    (マイリスト)                            │
│    └── ProfileHandler   (プロフィール)                          │
├─────────────────────────────────────────────────────────────────┤
│  Service層                                                      │
│    ├── AuthService      (JWT生成/検証)                          │
│    └── FoodService      (食品分析ロジック)                       │
├─────────────────────────────────────────────────────────────────┤
│  Repository層                                                   │
│    ├── AnalysisRepository                                       │
│    ├── UserRepository                                           │
│    ├── WeightRepository                                         │
│    ├── ConditionRepository                                      │
│    ├── TrainingRepository                                       │
│    ├── MylistRepository                                         │
│    └── ProfileRepository                                        │
├─────────────────────────────────────────────────────────────────┤
│  Worker                                                         │
│    └── AnalysisWorker   (非同期分析処理)                         │
└──────────┬──────────────────────────────────────────────────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
┌─────────┐  ┌─────────┐
│PostgreSQL│  │Gemini CLI│
│ (DB)     │  │ (AI)     │
└─────────┘  └─────────┘
```

## データフロー

### 食事画像分析フロー

```
1. 画像アップロード → AnalyzeHandler → AnalysisRepository (pending保存)
2. AnalysisWorker (ポーリング) → pending取得
3. FoodService → Gemini CLI実行 → 栄養素計算
4. AnalysisRepository (completed更新)
5. フロントエンド (ポーリング) → 結果取得
```

### 認証フロー

```
1. ログイン/登録 → AuthHandler → UserRepository
2. JWT生成 → AuthService
3. クライアント保存 (localStorage/Keychain)
4. API呼び出し → Authorization: Bearer {token}
5. AuthMiddleware → トークン検証 → ハンドラー実行
```

## ディレクトリ構造

```
main/
├── frontend/          # Next.jsフロントエンド
│   ├── app/          # App Router
│   ├── components/   # UIコンポーネント
│   ├── hooks/        # カスタムフック
│   ├── contexts/     # Reactコンテキスト
│   ├── lib/          # ユーティリティ
│   ├── types/        # 型定義
│   └── e2e/          # E2Eテスト (Playwright)
├── backend/           # Golangバックエンド
│   ├── cmd/          # エントリーポイント
│   ├── internal/     # 内部パッケージ
│   ├── pkg/          # 共有パッケージ
│   └── database/     # マイグレーション
├── ios/               # iOSアプリ
│   ├── Uchikomi/     # メインアプリ
│   ├── UchikomiWidget/ # ウィジェット
│   └── UchikomiTests/ # テスト
└── docs/              # ドキュメント
    └── CODEMAPS/     # コードマップ
```

## 外部依存関係

| サービス | 用途 |
|:---|:---|
| Gemini CLI | 食事画像分析、テキストからの食品抽出、栄養素計算 |
| PostgreSQL | データ永続化 |

## 関連コードマップ

- [バックエンド構造](./backend.md)
- [フロントエンド構造](./frontend.md)
- [iOS構造](./ios.md)
- [データモデル](./data.md)
