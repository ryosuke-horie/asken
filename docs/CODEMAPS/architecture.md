# 全体アーキテクチャ

最終更新: 2026-02-08

## システム概要

ウチコミは格闘技の減量・体重コントロール支援アプリケーション。日々の記録（体重、食事、体調、トレーニング）とAI相談で減量を支援する。

## 技術スタック

| レイヤー | 技術 |
|:---|:---|
| iOSアプリ | Swift, SwiftUI |
| バックエンド | Golang (Cloud Run) |
| コンテナレジストリ | Artifact Registry |
| データベース | Firestore |
| ストレージ | Cloud Storage |
| シークレット | Secret Manager |
| 認証 | Firebase Auth |
| AI | Gemini API |
| CI/CD | GitHub Actions |
| CI/CD認証 | Workload Identity Federation |
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
│                     Go Backend (Cloud Run)                       │
├─────────────────────────────────────────────────────────────────┤
│  認証検証 (Firebase Auth)                                        │
│  ビジネスロジック                                                  │
│  データアクセス                                                   │
└──────────┬─────────────────┬─────────────────┬──────────────────┘
           │                 │                 │
           ▼                 ▼                 ▼
    ┌──────────┐      ┌──────────┐      ┌──────────┐
    │ Firestore│      │  Cloud   │      │  Gemini  │
    │          │      │ Storage  │      │   API    │
    └──────────┘      └──────────┘      └──────────┘
```

## データフロー

### 食事画像分析フロー

```
1. iOSアプリ → Go Backend: 画像アップロード
2. Go Backend → Cloud Storage: 画像保存
3. Go Backend → Gemini API: 画像分析リクエスト（ワーカーで非同期）
4. Go Backend → Firestore: 結果保存
5. iOSアプリ → Go Backend: ポーリングで結果取得
```

### 体重記録フロー

```
1. iOSアプリ → Go Backend: 体重データ送信 (POST /api/weight/records)
2. Go Backend → Firestore: 体重記録保存
3. iOSアプリ → Go Backend: 期間指定で一覧取得 (GET /api/weight/records)
4. iOSアプリ: Swift Chartsでグラフ表示
```

### 認証フロー

```
1. iOSアプリ → Firebase Auth: Google Sign-In / Apple Sign-In
2. Firebase Auth → iOSアプリ: FirebaseAuthUser (IDトークン取得可能)
3. iOSアプリ → Go Backend: API呼び出し (Authorization: Bearer {token})
4. Go Backend → Firebase Admin SDK: トークン検証
5. Go Backend: firebase_uid をContextに設定
6. Go Backend: リクエスト処理
```

### 開発環境の認証フロー

シミュレータではGoogle Sign-Inのパスキー認証が動作しないため、モック認証を使用:

```
1. iOSアプリ (シミュレータ): MockFirebaseAuthService使用
2. 「開発用ログイン」ボタンタップ
3. 固定トークン "dev-mock-token" を返す
4. Go Backend (APP_ENV=development): DevAuthMiddleware使用
5. トークン "dev-mock-token" を検証 → UID "dev-mock-user" を設定
```

## ディレクトリ構造

```
utikomi/
├── backend/           # Goバックエンド (Cloud Run)
│   ├── cmd/          # エントリーポイント
│   ├── internal/     # 内部パッケージ
│   └── pkg/          # 共有パッケージ
├── ios/               # iOSアプリ
│   ├── Uchikomi/     # メインアプリ
│   ├── UchikomiCore/ # コアフレームワーク
│   └── UchikomiTests/ # テスト
├── infrastructure/    # Terraformインフラ
│   ├── environments/ # 環境別設定 (dev, prod)
│   ├── modules/      # 再利用可能モジュール
│   └── scripts/      # セットアップスクリプト
└── docs/              # ドキュメント
    ├── CODEMAPS/     # コードマップ
    ├── adr/          # アーキテクチャ決定記録
    └── plan/         # 実装計画
```

## 外部依存関係

| サービス | 用途 |
|:---|:---|
| Cloud Run | サーバーレスコンテナ実行 |
| Artifact Registry | Dockerイメージ保存 |
| Firestore | NoSQLデータベース |
| Cloud Storage | 画像保存 |
| Secret Manager | APIキー等のシークレット管理 |
| Firebase Auth | ユーザー認証 |
| Gemini API | AI画像分析・栄養素計算 |
| Workload Identity Federation | GitHub Actions→GCP認証（キーレス） |

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
    ├── artifact-registry/  # Artifact Registryモジュール
    ├── cloud-run/          # Cloud Runモジュール
    ├── firestore/          # Firestoreモジュール
    ├── storage/            # Cloud Storageモジュール
    ├── firebase-auth/      # Firebase Authモジュール
    ├── github/             # GitHub secrets/variablesモジュール
    └── wif/                # Workload Identity Federationモジュール
```

### Terraformで管理するリソース

| リソース | 用途 |
|:---|:---|
| Cloud Run Service | バックエンドAPIホスティング |
| Artifact Registry | Dockerイメージ保存 |
| Firestore Database | データ永続化 |
| Cloud Storage Bucket | 画像保存 |
| Secret Manager | APIキー等のシークレット保存 |
| Firebase Auth | ユーザー認証 |
| Workload Identity Pool/Provider | GitHub Actions認証 |
| GitHub Secrets | CI/CD用シークレット |
| GitHub Variables | CI/CD用環境変数 |

## CI/CDパイプライン

```
┌─────────────────────────────────────────────────────────────────┐
│                     GitHub Actions                              │
├─────────────────────────────────────────────────────────────────┤
│  1. Push to main (backend/**)                                   │
│  2. WIF認証 (キーレス)                                           │
│  3. Docker Build (multi-stage)                                  │
│  4. Push to Artifact Registry                                   │
│  5. Deploy to Cloud Run                                         │
└─────────────────────────────────────────────────────────────────┘

ワークフロー: .github/workflows/deploy.yml
```

## 関連コードマップ

- [バックエンド構造](./backend.md)
- [iOS構造](./ios.md)
- [データモデル](./data.md)
