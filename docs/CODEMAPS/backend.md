# バックエンドアーキテクチャ

最終更新: 2026-02-03
フレームワーク: Golang (標準ライブラリ)
エントリーポイント: backend/cmd/server/main.go
デプロイ先: Cloud Run (asia-northeast1)

## ディレクトリ構造

```
backend/
├── Dockerfile              # マルチステージビルド (distroless/nonroot)
├── .dockerignore           # Docker除外設定
├── cmd/
│   └── server/main.go      # HTTPサーバーエントリーポイント
├── internal/
│   ├── handler/            # HTTPハンドラ
│   ├── middleware/         # ミドルウェア（認証）
│   ├── repository/         # データアクセス (Firestore)
│   ├── service/            # ビジネスロジック
│   └── worker/             # バックグラウンドワーカー
└── pkg/
    ├── database/           # Firestore接続
    └── gemini/             # Gemini CLI連携
```

## レイヤードアーキテクチャ

```
Handler → Service → Repository → Firestore
           ↓
        Gemini CLI

Middleware (Authenticator)
├── AuthMiddleware      # Firebase Auth (本番)
└── DevAuthMiddleware   # モック認証 (開発)
```

## 認証アーキテクチャ

Firebase Authenticationを使用。iOSアプリから送信されたIDトークンをバックエンドで検証。

```
iOSアプリ → Firebase Auth → IDトークン取得
                              ↓
Go Backend ← Authorization: Bearer {token}
    ↓
Firebase Admin SDK → トークン検証 → UID取得
    ↓
Context に firebase_uid を設定
```

### 開発環境

`APP_ENV=development`の場合、DevAuthMiddlewareが有効:
- トークン`dev-mock-token`で固定UID`dev-mock-user`として認証
- Firebase Admin SDKを初期化しない

## APIエンドポイント

### 食事分析 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/analyze | AnalyzeHandler | 分析リクエスト作成 |
| GET | /api/analyze/{id} | StatusHandler | 分析ステータス取得 |
| POST | /api/upload-image | AnalyzeHandler | 画像アップロード |
| GET | /api/meals/daily | DailyMealsHandler | 日次食事取得 |
| POST | /api/meals/skip | SkipMealHandler | 食事スキップ記録 |

### 履歴 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/history | HistoryHandler | 履歴一覧 |
| GET | /api/history/{id} | HistoryHandler | 履歴詳細 |
| PUT | /api/history/{id} | HistoryHandler | 履歴更新（食材編集） |
| DELETE | /api/history/{id} | HistoryDeleteHandler | 履歴削除 |

### 画像配信 (認証不要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/images/{uuid} | ImageHandler | 画像ファイル配信 |

## 主要モジュール

### Handlers (internal/handler/)

| ファイル | 責務 |
|:---|:---|
| analyze_handler.go | 食事分析リクエスト |
| status_handler.go | 分析ステータス確認 |
| daily_meals_handler.go | 日次食事データ |
| history_handler.go | 履歴一覧・詳細・更新 |
| history_delete_handler.go | 履歴削除 |
| skip_meal_handler.go | 食事スキップ |
| image_handler.go | 画像配信 |

### Repositories (internal/repository/)

| ファイル | 責務 |
|:---|:---|
| analysis_models.go | 分析関連の型定義・インターフェース |
| analysis_repository_firestore.go | 分析リクエスト・結果（Firestore実装） |

**注意**: 体重、体調、トレーニング、マイリスト、プロフィールのリポジトリは未実装です。

### Middleware (internal/middleware/)

| ファイル | 責務 |
|:---|:---|
| auth.go | Firebase Auth認証ミドルウェア、Authenticatorインターフェース |
| dev_auth.go | 開発用モック認証ミドルウェア |

### Services (internal/service/)

| ファイル | 責務 |
|:---|:---|
| firebase_auth_service.go | Firebase Admin SDKラッパー |
| food_service.go | 食品分析ロジック |

### Gemini連携 (pkg/gemini/)

| ファイル | 責務 |
|:---|:---|
| client.go | Gemini CLI実行基盤 |
| classifier.go | 画像から食品分類 |
| text_parser.go | テキストから食品抽出 |
| nutrition.go | 栄養素計算 |
| training_menu.go | トレーニングメニュー提案（未使用） |
| equipment_normalizer.go | 器具名正規化（未使用） |

### Worker (internal/worker/)

| ファイル | 責務 |
|:---|:---|
| analysis_worker.go | 非同期分析処理 (5秒間隔ポーリング) |

### Database (pkg/database/)

| ファイル | 責務 |
|:---|:---|
| firestore.go | Firestoreクライアント初期化 |

## 依存関係図

```
cmd/server/main.go
├── internal/middleware/
│   ├── auth.go (Authenticator interface)
│   │   └── AuthMiddleware (Firebase本番)
│   │       └── internal/service/firebase_auth_service.go
│   └── dev_auth.go
│       └── DevAuthMiddleware (開発モック)
├── internal/handler/*
│   ├── internal/service/*
│   │   └── pkg/gemini/*
│   └── internal/repository/*
│       └── pkg/database/firestore.go
└── internal/worker/analysis_worker.go
    ├── internal/service/food_service.go
    └── internal/repository/analysis_repository_firestore.go
```

## コンテナ化

### Dockerfile (マルチステージビルド)

```
Stage 1: builder (golang:1.25-alpine)
  - 依存関係のダウンロード
  - 静的バイナリのビルド (CGO_ENABLED=0)
  - LDFlags: -w -s (デバッグ情報除去)

Stage 2: runtime (distroless/static-debian12:nonroot)
  - 最小イメージ（約50MB以下）
  - 非rootユーザーで実行
  - ポート8080公開
```

### ヘルスチェック

| タイプ | パス | 設定 |
|:---|:---|:---|
| startup | /api/health | 10秒後開始、10秒間隔、5秒タイムアウト、3回失敗で再起動 |
| liveness | /api/health | 30秒間隔、5秒タイムアウト、3回失敗で再起動 |

### 環境変数

| 変数 | 説明 | 設定元 |
|:---|:---|:---|
| GOOGLE_APPLICATION_CREDENTIALS | Firebase/Firestore認証 | Secret Manager |
| APP_ENV | 環境 (development/production) | Cloud Run環境変数 |
| ALLOWED_ORIGINS | CORSオリジン | Cloud Run環境変数 |

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
