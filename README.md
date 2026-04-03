# ウチコミ - 格闘技向け体重管理アプリ

柔術・キックボクシングなど格闘技の体重コントロールを支援するiOSアプリ。
食事写真をAIが自動分析し、カロリー・PFCバランスを記録。体重推移の可視化や栄養目標の管理で、計量に向けた減量をサポートする。

## 主な機能

- 食事画像のAI栄養素分析（Gemini API）
- 体重記録・推移グラフ
- 栄養目標設定（PFCバランス）
- 運動記録・消費カロリー推定
- 食事・体重リマインダー通知

## 技術スタック

| レイヤー | 技術 |
|:---|:---|
| iOS | Swift / SwiftUI |
| バックエンド | Go 1.25 |
| AI | Gemini API |
| データベース | Cloud Firestore |
| ストレージ | Cloud Storage |
| インフラ | Cloud Run / Terraform |
| 認証 | Firebase Authentication（Google Sign-In） |
| CI/品質管理 | Lefthook / golangci-lint / SwiftLint |

## アーキテクチャ

```mermaid
graph TB
    subgraph Client
        iOS[iOS App<br/>Swift / SwiftUI]
    end

    subgraph Google Cloud
        CR[Cloud Run<br/>Go API Server]
        FS[Cloud Firestore]
        CS[Cloud Storage]
        FA[Firebase Auth]
    end

    subgraph External
        GM[Gemini API]
    end

    iOS -->|REST API| CR
    iOS -->|認証| FA
    CR -->|データ永続化| FS
    CR -->|画像保存| CS
    CR -->|栄養素分析| GM
```

## ディレクトリ構成

```
utikomi/
├── backend/             # Go バックエンド
│   ├── cmd/server/      # エントリーポイント
│   ├── internal/        # handler / service / repository / middleware
│   └── pkg/             # gemini / database / storage
├── ios/                 # iOS アプリ（SwiftUI）
│   ├── Uchikomi/        # Features / Core / Resources
│   └── UchikomiCore/    # コアフレームワーク
├── infrastructure/      # Terraform（GCP）
└── docs/                # ドキュメント・ADR
```