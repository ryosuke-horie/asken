# 全体アーキテクチャ

最終更新: 2026-02-02

## システム概要

ウチコミは格闘技の減量・体重コントロール支援アプリケーション。日々の記録（体重、食事、体調、トレーニング）とAI相談で減量を支援する。

## 技術スタック

### 現行環境（exe.dev VM）

| レイヤー | 技術 |
|:---|:---|
| iOSアプリ | Swift, SwiftUI |
| バックエンド | Golang (標準ライブラリ) |
| データベース | PostgreSQL |
| AI | Gemini CLI (gemini-3-flash-preview) |
| ホスティング | exe.dev (Ubuntu) |

### 新アーキテクチャ（GCPサーバレス）- 移行中

| レイヤー | 技術 |
|:---|:---|
| iOSアプリ | Swift, SwiftUI |
| データベース | Firestore |
| ストレージ | Cloud Storage |
| 認証 | Firebase Auth |
| AI | Gemini API |
| インフラ管理 | Terraform |

## システム構成図

```
┌─────────────────────────────────────────────────────────────────┐
│                         クライアント                              │
├─────────────────────────────────────────────────────────────────┤
│                          iOS App                                │
│                         (SwiftUI)                               │
└──────────────────────────────┬──────────────────────────────────┘
                               │
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
5. iOSアプリ (ポーリング) → 結果取得
```

### 認証フロー

```
1. ログイン/登録 → AuthHandler → UserRepository
2. JWT生成 → AuthService
3. クライアント保存 (Keychain)
4. API呼び出し → Authorization: Bearer {token}
5. AuthMiddleware → トークン検証 → ハンドラー実行
```

## ディレクトリ構造

```
utikomi/
├── backend/           # Golangバックエンド
│   ├── cmd/          # エントリーポイント
│   ├── internal/     # 内部パッケージ
│   ├── pkg/          # 共有パッケージ
│   └── database/     # マイグレーション
├── ios/               # iOSアプリ
│   ├── Uchikomi/     # メインアプリ
│   └── UchikomiTests/ # テスト
├── infrastructure/    # Terraformインフラ
│   ├── environments/ # 環境別設定 (dev, prod)
│   ├── modules/      # 再利用可能モジュール
│   └── scripts/      # セットアップスクリプト
└── docs/              # ドキュメント
    └── CODEMAPS/     # コードマップ
```

## 外部依存関係

### 現行環境

| サービス | 用途 |
|:---|:---|
| Gemini CLI | 食事画像分析、テキストからの食品抽出、栄養素計算 |
| PostgreSQL | データ永続化 |

### 新アーキテクチャ（GCP）

| サービス | 用途 |
|:---|:---|
| Firestore | NoSQLデータベース |
| Cloud Storage | 画像保存 |
| Firebase Auth | ユーザー認証 |
| Gemini API | AI画像分析・栄養素計算 |

## インフラ構成（Terraform）

```
infrastructure/
├── environments/
│   └── dev/              # dev環境
│       ├── main.tf       # モジュール呼び出し
│       ├── providers.tf  # プロバイダー設定
│       ├── backend.tf    # GCSバックエンド
│       ├── variables.tf  # 変数定義
│       └── outputs.tf    # 出力定義
└── modules/
    ├── firestore/        # Firestoreモジュール
    ├── storage/          # Cloud Storageモジュール
    ├── firebase-auth/    # Firebase Authモジュール
    └── github/           # GitHub secrets/variablesモジュール
```

### Terraformで管理するリソース

| リソース | 用途 |
|:---|:---|
| Firestore Database | データ永続化 |
| Cloud Storage Bucket | 画像保存 |
| Firebase Auth | ユーザー認証 |
| GitHub Secrets | CI/CD用シークレット |
| GitHub Variables | CI/CD用環境変数 |

## 関連コードマップ

- [バックエンド構造](./backend.md)
- [iOS構造](./ios.md)
- [データモデル](./data.md)
