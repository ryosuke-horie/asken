# 開発者ガイド（CONTRIB）

最終更新: 2026-02-02

このドキュメントは、ウチコミプロジェクトの開発ワークフロー、利用可能なコマンド、環境セットアップ、テスト手順を説明します。

## 前提条件

| ツール | バージョン | 用途 |
|:---|:---|:---|
| mise | 最新 | ツール・環境変数管理 |
| Go | 1.25以上 | バックエンド開発（miseで自動インストール） |
| Terraform | 1.10以上 | インフラ管理（miseで自動インストール） |
| Docker / Docker Compose | 最新 | PostgreSQL起動 |
| Task | 3.x | タスクランナー |
| golangci-lint | 最新 | Goリント |
| Xcode | 16以上 | iOS開発 |
| Mockolo | 最新 | iOSモック生成 |
| 1Password CLI | 最新 | シークレット管理（オプション） |

## 環境セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/ryosuke-horie/utikomi.git
cd utikomi
```

### 2. miseのセットアップ

このプロジェクトではmiseを使用してツールと環境変数を管理しています。

```bash
# 初回のみ信頼設定
mise trust
mise install
```

miseが自動的に以下を管理します:
- **ツールバージョン**: Go 1.25.6, Terraform 1.10
- **GCP環境変数**: `GCP_PROJECT_ID`, `GCP_REGION`, `GCP_ZONE`
- **gcloud構成の自動切り替え**: `CLOUDSDK_ACTIVE_CONFIG_NAME=utikomi-dev`
- **Terraform変数**: `TF_VAR_*` プレフィックスで自動設定

### 3. 依存関係のインストール

```bash
# Go依存関係
task setup
```

### 4. 環境変数の設定

#### バックエンド

`backend/.env`を作成（`backend/.env.example`を参考）:

| 環境変数 | 説明 | 例 |
|:---|:---|:---|
| DATABASE_URL | PostgreSQL接続文字列 | `postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable` |

#### インフラ管理（Terraform）

Terraform用の環境変数は`.mise.toml`で管理されています。
1Password CLIを使用している場合、シークレットは自動的に注入されます。

詳細は[infrastructure/README.md](../infrastructure/README.md)を参照してください。

## 開発ワークフロー

### 日常の開発サイクル

```bash
# 1. DBを起動
task db-up

# 2. テストデータを投入（必要に応じて）
task db-seed

# 3. バックエンドサーバーを起動
task run
```

### コード変更後の確認

```bash
# バックエンド
task lint
task test

# iOS
task ios:test
```

### 開発終了時

```bash
task db-down
```

## 利用可能なコマンド（Taskfile）

### 一般

| コマンド | 説明 |
|:---|:---|
| `task --list` | 利用可能なコマンド一覧を表示 |
| `task help` | 利用可能なコマンド一覧を表示 |

### バックエンド

| コマンド | 説明 |
|:---|:---|
| `task setup` | Go依存関係をダウンロード |
| `task clean` | ビルド成果物を削除 |
| `task test` | Goテストを実行 |
| `task test:coverage` | Goカバレッジ計測付きテスト |
| `task lint` | golangci-lintを実行 |
| `task build` | Goバイナリをビルド |
| `task run` | バックエンドサーバーを起動 |

### データベース

| コマンド | 説明 |
|:---|:---|
| `task db-up` | PostgreSQLを起動 |
| `task db-down` | PostgreSQLを停止 |
| `task db-seed` | テストデータを投入 |
| `task db-clean` | DBをリセット（ボリューム削除+再起動） |

### iOS

| コマンド | 説明 |
|:---|:---|
| `task ios:generate-mocks` | Mockoloでモックを生成 |
| `task ios:test` | iOSのテストを実行 |
| `task ios:test:coverage` | iOSカバレッジ計測付きテスト |
| `task ios:lint` | SwiftLintを実行 |
| `task ios:format` | SwiftFormatを実行（コード整形） |
| `task ios:format-check` | SwiftFormatチェック（CI用） |

### デプロイ

| コマンド | 説明 |
|:---|:---|
| `task deploy` | 本番デプロイ（本番サーバーでのみ実行） |

## テスト手順

### バックエンドテスト

```bash
# 全テストを実行
task test

# 詳細出力
cd backend && go test ./... -v

# カバレッジ
cd backend && go test ./... -cover
```

### iOSテスト

```bash
# ユニットテスト
task ios:test

# モック再生成（Protocolを変更した場合）
task ios:generate-mocks
```

テストユーザー:
- メール: `test@example.com`
- パスワード: `Pass0123`

## Git運用

### ブランチ命名規則

```
edg-<LinearタスクID>

例: edg-305, edg-421
```

### コミットメッセージ形式

Conventional Commits形式:

```
<type>: <description>

例:
feat: 食事画像アップロード機能を追加
fix: 栄養素計算のバリデーションエラーを修正
refactor: AuthManagerのリファクタリング
```

### 作業フロー

```bash
# 1. ブランチを作成
git checkout -b edg-305

# 2. 作業を実施

# 3. コミット
git add .
git commit -m "feat: 機能の説明"

# 4. プッシュ
git push -u origin edg-305

# 5. PRを作成
gh pr create --title "タイトル" --body "Fixes EDG-305"
```

## インフラ管理（Terraform）

### 概要

GCPリソースはTerraformで管理しています。

| リソース | 用途 |
|:---|:---|
| Firestore | データベース |
| Cloud Storage | 画像保存 |
| Firebase Auth | ユーザー認証 |
| GitHub Secrets/Variables | CI/CD用シークレット |

### セットアップ

詳細は[infrastructure/README.md](../infrastructure/README.md)を参照してください。

```bash
cd infrastructure/environments/dev

# 初期化
terraform init

# プラン確認
terraform plan

# 適用
terraform apply
```

## 関連ドキュメント

- [README.md](../README.md) - プロジェクト概要
- [DEPLOY.md](../DEPLOY.md) - デプロイ手順
- [RUNBOOK.md](./RUNBOOK.md) - 運用手順書
- [infrastructure/README.md](../infrastructure/README.md) - Terraformインフラ管理
- [ios/README.md](../ios/README.md) - iOSアプリ開発ガイド
- [.claude/rules/](../.claude/rules/) - 詳細規約
