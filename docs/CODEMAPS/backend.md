# バックエンドアーキテクチャ

最終更新: 2026-03-01
フレームワーク: Golang (標準ライブラリ)
エントリーポイント: backend/cmd/server/main.go
デプロイ先: Cloud Run (asia-northeast1)

## ディレクトリ構造

```
backend/
├── Dockerfile              # マルチステージビルド (distroless/nonroot)
├── .dockerignore           # Docker除外設定
├── cmd/
│   ├── server/main.go      # HTTPサーバーエントリーポイント
│   └── ops/                # 運用スクリプト
│       └── 20250219/       # 微量栄養素バックフィルスクリプト
├── internal/
│   ├── handler/            # HTTPハンドラ
│   ├── middleware/         # ミドルウェア（認証、レート制限、セキュリティ）
│   ├── repository/         # データアクセス (Firestore)
│   ├── service/            # ビジネスロジック
│   ├── worker/             # バックグラウンドワーカー
│   ├── util/               # ユーティリティ（タイムゾーン等）
│   └── testutil/           # テストユーティリティ
├── pkg/
│   ├── database/           # Firestore接続
│   ├── gemini/             # Gemini HTTP API連携
│   └── storage/            # Cloud Storage接続
└── e2e/                    # E2Eテスト
```

## レイヤードアーキテクチャ

```
Handler → Service → Repository → Firestore
           ↓
        Gemini HTTP API (分類・栄養計算・レシート解析・メニューサジェスト)

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
- ビルドタグ`!production`で制御（`production`タグ付きビルドではDevAuthMiddlewareが無効化される）

## APIエンドポイント

### ヘルスチェック (認証不要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/health | HealthHandler | ヘルスチェック |

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

### 体重 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/weight/records | WeightRecordHandler | 体重記録一覧 |
| POST | /api/weight/records | WeightRecordHandler | 体重記録作成 |
| GET | /api/weight/records/{id} | WeightRecordHandler | 体重記録取得 |
| PUT | /api/weight/records/{id} | WeightRecordHandler | 体重記録更新 |
| DELETE | /api/weight/records/{id} | WeightRecordHandler | 体重記録削除 |
| GET | /api/weight/goal | WeightGoalHandler | 目標体重取得 |
| PUT | /api/weight/goal | WeightGoalHandler | 目標体重設定 |

### 栄養目標 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/nutrition/goal | NutritionGoalHandler | 栄養目標取得 |
| PUT | /api/nutrition/goal | NutritionGoalHandler | 栄養目標設定 |

### マイメニュー (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/my-menu | MyMenuHandler | マイメニュー一覧 |
| POST | /api/my-menu | MyMenuHandler | マイメニュー作成 |
| GET | /api/my-menu/{id} | MyMenuHandler | マイメニュー詳細 |
| PUT | /api/my-menu/{id} | MyMenuHandler | マイメニュー更新 |
| DELETE | /api/my-menu/{id} | MyMenuHandler | マイメニュー削除 |
| POST | /api/my-menu/{id}/record | MyMenuHandler | マイメニューから食事記録 |

### 食材管理 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/ingredients | IngredientHandler | 食材一覧 |
| POST | /api/ingredients | IngredientHandler | 食材作成 |
| PUT | /api/ingredients/{id} | IngredientHandler | 食材更新 |
| DELETE | /api/ingredients/{id} | IngredientHandler | 食材削除 |
| POST | /api/ingredients/scan-receipt | ScanReceiptHandler | レシート読取 |

### メニューサジェスト (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/menu/suggest | MenuSuggestionHandler | サジェストリクエスト |
| GET | /api/menu/suggestions | MenuSuggestionHandler | サジェスト一覧 |
| GET | /api/menu/suggestions/{id} | MenuSuggestionHandler | サジェスト詳細 |
| POST | /api/menu/suggestions/{id}/accept | MenuSuggestionHandler | サジェスト採用 |
| POST | /api/menu/suggestions/{id}/dismiss | MenuSuggestionHandler | サジェスト却下 |

### 消費カロリー記録 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/exercise/records | ExerciseHandler | 運動記録作成（Gemini APIまたはMET値で消費カロリー推定） |
| DELETE | /api/exercise/records/{id} | ExerciseHandler | 運動記録削除 |
| GET | /api/exercise/daily | ExerciseHandler | 日次運動一覧・消費カロリー合計取得 |

### 画像配信 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/images/{uuid} | ImageHandler | 画像ファイル配信 |

## 主要モジュール

### Handlers (internal/handler/)

| ファイル | 責務 |
|:---|:---|
| health_handler.go | ヘルスチェック |
| analyze_handler.go | 食事分析リクエスト・画像アップロード |
| status_handler.go | 分析ステータス確認 |
| daily_meals_handler.go | 日次食事データ |
| history_handler.go | 履歴一覧・詳細・更新（NutritionRecalculator依存、非同期再計算） |
| history_delete_handler.go | 履歴削除 |
| skip_meal_handler.go | 食事スキップ |
| image_handler.go | 画像配信 |
| weight_record_handler.go | 体重記録CRUD |
| weight_goal_handler.go | 目標体重取得・設定 |
| nutrition_goal_handler.go | 栄養目標取得・設定 |
| my_menu_handler.go | マイメニューCRUD・食事記録 |
| ingredient_handler.go | 食材CRUD |
| scan_receipt_handler.go | レシート読取（Gemini API経由） |
| menu_suggestion_handler.go | メニューサジェスト（提案・一覧・詳細・採用・却下） |
| exercise_handler.go | 消費カロリー記録CRUD・日次取得 |

### Repositories (internal/repository/)

| ファイル | 責務 |
|:---|:---|
| analysis_models.go | 分析関連の型定義・インターフェース |
| analysis_repository_firestore.go | 分析リクエスト・結果（Firestore実装） |
| storage_repository.go | 画像ストレージ操作（Cloud Storage） |
| weight_models.go | 体重関連の型定義・インターフェース |
| weight_repository_firestore.go | 体重記録・目標（Firestore実装） |
| nutrition_goal_models.go | 栄養目標の型定義・インターフェース・PFC計算 |
| nutrition_goal_repository_firestore.go | 栄養目標（Firestore実装） |
| my_menu_repository.go | マイメニュー型定義・インターフェース（`totalMicronutrients` を含む） |
| my_menu_repository_firestore.go | マイメニュー（Firestore実装、微量栄養素合算を保存） |
| ingredient_models.go | 食材型定義・カテゴリ・インターフェース |
| ingredient_repository_firestore.go | 食材CRUD（Firestore実装） |
| menu_suggestion_models.go | メニューサジェスト型定義・インターフェース |
| menu_suggestion_repository_firestore.go | メニューサジェストCRUD（Firestore実装） |
| exercise_models.go | 運動記録型定義・インターフェース・バリデーション |
| exercise_repository_firestore.go | 運動記録CRUD（Firestore実装） |

### Middleware (internal/middleware/)

| ファイル | 責務 |
|:---|:---|
| auth.go | Firebase Auth認証ミドルウェア、Authenticatorインターフェース、IsDevMode() |
| dev_auth.go | 開発用モック認証ミドルウェア（`!production`ビルドタグ） |
| dev_auth_prod.go | 本番用DevAuthMiddlewareスタブ（`production`ビルドタグ、常に拒否） |
| rate_limit.go | レート制限ミドルウェア（Gemini API用） |
| rate_limit_config.go | レート制限設定 |
| security_headers.go | セキュリティヘッダー設定ミドルウェア |

### Services (internal/service/)

| ファイル | 責務 |
|:---|:---|
| firebase_auth_service.go | Firebase Admin SDKラッパー |
| food_service.go | 食品分析ロジック |
| exercise_service.go | 運動記録作成・消費カロリー推定ロジック（METテーブルまたはGemini API） |

### Utility (internal/util/)

| ファイル | 責務 |
|:---|:---|
| log.go | ログ用文字列切り詰めユーティリティ |
| timezone.go | タイムゾーンユーティリティ |

### Gemini連携 (pkg/gemini/)

| ファイル | 責務 |
|:---|:---|
| http_client.go | Gemini HTTP API クライアント |
| client.go | HTTPClient ラッパー（後方互換性） |
| classifier.go | 画像から食品分類 |
| text_parser.go | テキストから食品抽出 |
| nutrition.go | 栄養素計算 |
| micronutrients.go | 微量栄養素メタデータ（鉄、カルシウム、亜鉛、食物繊維、ビタミンA/B1/B2/B6/B12/C/D/E） |
| schema.go | Gemini APIレスポンスのJSONスキーマ定義 |
| receipt_parser.go | レシート画像解析 |
| menu_suggester.go | メニューサジェスト生成 |
| exercise_estimator.go | Gemini APIによる消費カロリー推定（プリセット外の運動種目に使用） |
| mock_http_client.go | テスト用HTTPクライアントモック |

### Worker (internal/worker/)

| ファイル | 責務 |
|:---|:---|
| analysis_worker.go | 非同期分析処理 (5秒間隔ポーリング) |

### Storage (pkg/storage/)

| ファイル | 責務 |
|:---|:---|
| client.go | Cloud Storageクライアント初期化 |

### Database (pkg/database/)

| ファイル | 責務 |
|:---|:---|
| firestore.go | Firestoreクライアント初期化 |

### Operations (cmd/ops/)

| ディレクトリ | 責務 |
|:---|:---|
| 20250219/ | 既存MyMenuデータへの微量栄養素バックフィルスクリプト |

### E2Eテスト (e2e/)

| ファイル | 責務 |
|:---|:---|
| e2e_test.go | E2Eテストスイートセットアップ |
| health_test.go | ヘルスチェックE2Eテスト |
| analyze_test.go | 分析フローE2Eテスト |
| auth.go | テスト用認証ヘルパー |
| helpers.go | テスト用共通ヘルパー |
| cleanup.go | テストデータクリーンアップ |
| history_test.go | 履歴API E2Eテスト |
| meals_test.go | 食事API E2Eテスト |
| weight_test.go | 体重API E2Eテスト（記録CRUD、目標設定） |
| image_test.go | 画像配信API E2Eテスト |
| ingredients_test.go | 食材管理API E2Eテスト |
| menu_test.go | メニューサジェストAPI E2Eテスト |

## 依存関係図

```
cmd/server/main.go
├── internal/middleware/
│   ├── auth.go (Authenticator interface)
│   │   └── AuthMiddleware (Firebase本番)
│   │       └── internal/service/firebase_auth_service.go
│   ├── dev_auth.go (!production ビルドタグ)
│   │   └── DevAuthMiddleware (開発モック)
│   ├── dev_auth_prod.go (production ビルドタグ)
│   │   └── DevAuthMiddleware (常に拒否)
│   └── security_headers.go
│       └── SecurityHeaders (全ルート適用)
├── internal/handler/*
│   ├── internal/service/*
│   │   └── pkg/gemini/*
│   ├── internal/repository/*
│   │   └── pkg/database/firestore.go
│   ├── history_handler.go
│   │   └── NutritionRecalculator (pkg/gemini/nutrition.go)
│   ├── ingredient_handler.go
│   │   └── repository.IngredientRepository
│   ├── scan_receipt_handler.go
│   │   └── pkg/gemini/receipt_parser.go
│   ├── menu_suggestion_handler.go
│   │   ├── repository.MenuSuggestionRepository
│   │   ├── repository.IngredientRepository
│   │   ├── repository.NutritionGoalRepository
│   │   ├── repository.WeightRecordRepository
│   │   ├── repository.AnalysisRepository
│   │   └── pkg/gemini/menu_suggester.go
│   └── exercise_handler.go
│       └── service.ExerciseService
│           ├── repository.ExerciseRepository
│           └── pkg/gemini/exercise_estimator.go
├── internal/worker/analysis_worker.go
│   ├── internal/service/food_service.go
│   └── internal/repository/analysis_repository_firestore.go
└── internal/repository/
    ├── storage_repository.go
    │   └── pkg/storage/client.go
    ├── weight_repository_firestore.go
    │   └── pkg/database/firestore.go
    ├── nutrition_goal_repository_firestore.go
    │   └── pkg/database/firestore.go
    ├── my_menu_repository_firestore.go
    │   └── pkg/database/firestore.go
    ├── ingredient_repository_firestore.go
    │   └── pkg/database/firestore.go
    ├── menu_suggestion_repository_firestore.go
    │   └── pkg/database/firestore.go
    └── exercise_repository_firestore.go
        └── pkg/database/firestore.go
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
| GOOGLE_APPLICATION_CREDENTIALS | Firebase/Firestore認証 | ローカル: 手動設定 / Cloud Run: サービスアカウント |
| GEMINI_API_KEY | Gemini API キー | ローカル: 手動設定 / Cloud Run: Secret Manager |
| GCS_BUCKET_NAME | Cloud Storageバケット名 | ローカル: 手動設定 / Cloud Run: 環境変数 |
| GCP_PROJECT_ID | GCPプロジェクトID | Cloud Run環境変数 |
| APP_ENV | 環境 (development/production) | Cloud Run環境変数 |
| ALLOWED_ORIGINS | CORSオリジン（カンマ区切り） | Cloud Run環境変数 |

Secret Manager管理のシークレット:

| シークレット名 | 説明 |
|:---|:---|
| gemini-api-key | Gemini API キー（Google AI Studio発行） |

## 運用スクリプト

| スクリプト | 役割 |
|:---|:---|
| `tools/deploy/deploy-dev.sh` | Cloud Runへの手動デプロイ |
| `tools/e2e/run-backend-e2e-dev.sh` | 開発環境のバックエンドE2Eテスト実行（デプロイと分離） |

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
