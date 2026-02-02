# 全体アーキテクチャ

最終更新: 2026-02-02

## システム概要

ウチコミは格闘技の減量・体重コントロール支援アプリケーション。日々の記録（体重、食事、体調、トレーニング）とAI相談で減量を支援する。

## 技術スタック

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
                     ┌─────────┼─────────┐
                     ▼         ▼         ▼
              ┌──────────┐ ┌──────────┐ ┌──────────┐
              │ Firebase │ │ Firestore│ │  Cloud   │
              │   Auth   │ │          │ │ Storage  │
              └──────────┘ └──────────┘ └──────────┘
                               │
                               ▼
                        ┌──────────┐
                        │  Gemini  │
                        │   API    │
                        └──────────┘
```

## データフロー

### 食事画像分析フロー

```
1. 画像アップロード → Cloud Storage
2. Gemini API呼び出し → 画像分析
3. 結果をFirestoreに保存
4. iOSアプリで結果取得
```

### 認証フロー

```
1. ログイン/登録 → Firebase Auth
2. トークン取得
3. クライアント保存 (Keychain)
4. API呼び出し時にトークン検証
```

## ディレクトリ構造

```
utikomi/
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

- [iOS構造](./ios.md)
- [データモデル](./data.md)
