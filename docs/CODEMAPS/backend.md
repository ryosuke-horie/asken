# バックエンドアーキテクチャ

最終更新: 2026-01-30
フレームワーク: Golang (標準ライブラリ)
エントリーポイント: backend/cmd/server/main.go

## ディレクトリ構造

```
backend/
├── cmd/
│   ├── server/main.go      # HTTPサーバーエントリーポイント
│   └── seed/main.go        # データシードエントリーポイント
├── internal/
│   ├── handler/            # HTTPハンドラ
│   ├── middleware/         # ミドルウェア
│   ├── repository/         # データアクセス
│   ├── service/            # ビジネスロジック
│   ├── seeder/             # シードデータ
│   └── worker/             # バックグラウンドワーカー
├── pkg/
│   ├── database/           # DB接続
│   └── gemini/             # Gemini CLI連携
└── database/
    └── migrations/         # SQLマイグレーション
```

## レイヤードアーキテクチャ

```
Handler → Service → Repository → PostgreSQL
           ↓
        Gemini CLI
```

## APIエンドポイント

### 認証 (認証不要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/auth/register | AuthHandler | ユーザー登録 |
| POST | /api/auth/login | AuthHandler | ログイン |

### 食事分析 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/analyze | AnalyzeHandler | 分析リクエスト作成 |
| GET | /api/analyze/{id} | StatusHandler | 分析ステータス取得 |
| POST | /api/upload-image | AnalyzeHandler | 画像アップロード |
| GET | /api/meals/daily | DailyMealsHandler | 日次食事取得 |
| POST | /api/meals/skip | SkipMealHandler | 食事スキップ記録 |
| POST | /api/meals/from-mylist | MylistHandler | マイリストから記録 |

### 履歴 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/history | HistoryHandler | 履歴一覧 |
| GET | /api/history/{id} | HistoryHandler | 履歴詳細 |
| DELETE | /api/history/{id} | HistoryDeleteHandler | 履歴削除 |

### 体重管理 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/weight-records | WeightHandler | 体重記録作成 |
| GET | /api/weight-records | WeightHandler | 体重記録取得 |
| GET | /api/weight-goal | WeightHandler | 目標体重取得 |
| PUT | /api/weight-goal | WeightHandler | 目標体重更新 |

### マイリスト (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/mylist | MylistHandler | 一覧取得 |
| POST | /api/mylist | MylistHandler | 新規作成 |
| GET | /api/mylist/{id} | MylistHandler | 詳細取得 |
| PUT | /api/mylist/{id} | MylistHandler | 更新 |
| DELETE | /api/mylist/{id} | MylistHandler | 削除 |
| PUT | /api/mylist/reorder | MylistHandler | 並び替え |
| POST | /api/mylist/analyze | MylistHandler | AI分析 |

### 体調記録 (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| POST | /api/condition-records | ConditionHandler | 体調記録作成 |
| GET | /api/condition-records | ConditionHandler | 体調記録取得 |

### トレーニング (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/training/locations | TrainingHandler | 場所一覧 |
| POST | /api/training/locations | TrainingHandler | 場所作成 |
| PUT | /api/training/locations/{id} | TrainingHandler | 場所更新 |
| DELETE | /api/training/locations/{id} | TrainingHandler | 場所削除 |
| GET | /api/training/locations/{id}/equipment | TrainingHandler | 器具一覧 |
| POST | /api/training/locations/{id}/equipment | TrainingHandler | 器具作成 |
| PUT | /api/training/equipment/{id} | TrainingHandler | 器具更新 |
| DELETE | /api/training/equipment/{id} | TrainingHandler | 器具削除 |
| GET | /api/training/records | TrainingHandler | 記録一覧 |
| POST | /api/training/records | TrainingHandler | 記録作成 |
| PUT | /api/training/records/{id} | TrainingHandler | 記録更新 |
| DELETE | /api/training/records/{id} | TrainingHandler | 記録削除 |
| GET | /api/training/menus | TrainingHandler | メニュー一覧 |
| POST | /api/training/menus | TrainingHandler | メニュー作成 |
| DELETE | /api/training/menus/{id} | TrainingHandler | メニュー削除 |
| POST | /api/training/suggest-menu | TrainingHandler | AIメニュー提案 |
| POST | /api/training/normalize-equipment | TrainingHandler | 器具名正規化 |

### プロフィール (認証必要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/profile | ProfileHandler | プロフィール取得 |
| PUT | /api/profile | ProfileHandler | プロフィール更新 |

### 画像配信 (認証不要)

| メソッド | パス | ハンドラ | 用途 |
|:---|:---|:---|:---|
| GET | /api/images/{uuid} | ImageHandler | 画像ファイル配信 |

## 主要モジュール

### Handlers (internal/handler/)

| ファイル | 責務 |
|:---|:---|
| auth_handler.go | 認証（登録、ログイン） |
| analyze_handler.go | 食事分析リクエスト |
| status_handler.go | 分析ステータス確認 |
| daily_meals_handler.go | 日次食事データ |
| history_handler.go | 履歴一覧・詳細 |
| history_delete_handler.go | 履歴削除 |
| weight_handler.go | 体重記録・目標 |
| mylist_handler.go | よく食べるもの管理 |
| skip_meal_handler.go | 食事スキップ |
| condition_handler.go | 体調記録 |
| training_handler.go | トレーニング管理 |
| profile_handler.go | プロフィール管理 |
| image_handler.go | 画像配信 |

### Repositories (internal/repository/)

| ファイル | 責務 |
|:---|:---|
| analysis_repository.go | 分析リクエスト・結果 |
| user_repository.go | ユーザー |
| weight_repository.go | 体重記録・目標 |
| mylist_repository.go | マイリスト |
| condition_repository.go | 体調記録 |
| training_repository.go | トレーニング |
| profile_repository.go | プロフィール |

### Services (internal/service/)

| ファイル | 責務 |
|:---|:---|
| auth_service.go | JWT生成・検証、パスワードハッシュ |
| food_service.go | 食品分析ロジック |

### Gemini連携 (pkg/gemini/)

| ファイル | 責務 |
|:---|:---|
| client.go | Gemini CLI実行基盤 |
| classifier.go | 画像から食品分類 |
| text_parser.go | テキストから食品抽出 |
| nutrition.go | 栄養素計算 |
| training_menu.go | トレーニングメニュー提案 |
| equipment_normalizer.go | 器具名正規化 |

### Worker (internal/worker/)

| ファイル | 責務 |
|:---|:---|
| analysis_worker.go | 非同期分析処理 (5秒間隔ポーリング) |

## 依存関係図

```
cmd/server/main.go
├── internal/handler/*
│   ├── internal/service/*
│   │   └── pkg/gemini/*
│   └── internal/repository/*
│       └── pkg/database/postgres.go
├── internal/middleware/auth.go
│   └── internal/service/auth_service.go
└── internal/worker/analysis_worker.go
    ├── internal/service/food_service.go
    └── internal/repository/analysis_repository.go
```

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
