# 全体アーキテクチャ

最終更新: 2026-02-02

## システム概要

ウチコミは格闘技の減量・体重コントロール支援アプリケーション。日々の記録（体重、食事、体調、トレーニング）とAI相談で減量を支援する。

## 技術スタック

| レイヤー | 技術 |
|:---|:---|
| iOSアプリ | Swift, SwiftUI |
| バックエンド | Golang (Cloud Functions) |
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
│                   Go Backend (Cloud Functions)                  │
├─────────────────────────────────────────────────────────────────┤
│  認証検証 (Firebase Auth)                                        │
│  ビジネスロジック                                                 │
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
3. Go Backend → Gemini API: 画像分析リクエスト
4. Go Backend → Firestore: 結果保存
5. iOSアプリ → Go Backend: 結果取得
```

### 認証フロー

```
1. iOSアプリ → Firebase Auth: ログイン/登録
2. Firebase Auth → iOSアプリ: IDトークン発行
3. iOSアプリ: トークンをKeychain保存
4. iOSアプリ → Go Backend: API呼び出し (Authorization: Bearer {token})
5. Go Backend → Firebase Auth: トークン検証
6. Go Backend: リクエスト処理
```

## ディレクトリ構造

```
utikomi/
├── backend/           # Goバックエンド (Cloud Functions)
│   ├── cmd/          # エントリーポイント
│   ├── internal/     # 内部パッケージ
│   └── pkg/          # 共有パッケージ
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
