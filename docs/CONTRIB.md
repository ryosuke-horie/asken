# 開発者ガイド（CONTRIB）

最終更新: 2026-02-15

このドキュメントは、ウチコミプロジェクトの開発ワークフロー、利用可能なコマンド、環境セットアップ、テスト手順を説明します。

## 前提条件

| ツール | バージョン | 用途 |
|:---|:---|:---|
| mise | 最新 | ツール・環境変数管理 |
| Go | 1.25以上 | バックエンド開発（miseで自動インストール） |
| Terraform | 1.10以上 | インフラ管理（miseで自動インストール） |
| Docker / Docker Compose | 最新 | ローカル開発用 |
| Task | 3.x | タスクランナー |
| Lefthook | 2.x | Gitフック（lint/test/format） |
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

# Git hooksをインストール
task hooks:install
```

miseが自動的に以下を管理します:
- ツールバージョン: Go 1.25.6, Terraform 1.10, Lefthook 2.1.1
- GCP環境変数: `GCP_PROJECT_ID`, `GCP_REGION`, `GCP_ZONE`
- gcloud構成の自動切り替え: `CLOUDSDK_ACTIVE_CONFIG_NAME=utikomi-dev`
- Terraform変数: `TF_VAR_*` プレフィックスで自動設定

### 3. 依存関係のインストール

```bash
# Go依存関係
task setup
```

### 4. 環境変数の設定

#### バックエンド

`backend/.env`を作成（`backend/.env.example`を参照）:

| 環境変数 | 必須 | 説明 | 例 |
|:---|:---|:---|:---|
| GOOGLE_APPLICATION_CREDENTIALS | 必須 | Firebase/Firestoreサービスアカウント鍵のパス | `/path/to/sa-key.json` |
| GEMINI_API_KEY | 必須 | Gemini API キー | `AIzaSy...` |
| GCS_BUCKET_NAME | 必須 | Cloud Storage バケット名（画像保存用） | `utikomi-dev-images` |
| ALLOWED_ORIGINS | オプション | CORS許可オリジン（カンマ区切り） | - |
| APP_ENV | オプション | 環境（`development`でモック認証有効） | `development` |

#### インフラ管理（Terraform）

Terraform用の環境変数は`.mise.toml`で管理されています。
1Password CLIを使用している場合、シークレットは自動的に注入されます。

| 環境変数 | 説明 | 管理方法 |
|:---|:---|:---|
| TF_VAR_gcp_project_id | GCPプロジェクトID | .mise.toml（固定値） |
| TF_VAR_gcp_region | GCPリージョン | .mise.toml（固定値） |
| TF_VAR_github_repository | GitHubリポジトリ | .mise.toml（固定値） |
| TF_VAR_github_token | GitHub PAT | 1Password経由 |
| TF_VAR_gcp_sa_key | GCPサービスアカウントキー | 1Password経由 |
| TF_VAR_gemini_api_key | Gemini APIキー | 1Password経由 |

詳細は[infrastructure/README.md](../infrastructure/README.md)を参照してください。

### 5. 開発用モック認証の設定

iOSシミュレータではGoogle Sign-Inのパスキー認証が動作しないため、開発用のモック認証が自動的に有効になります。

#### iOS側（自動）

シミュレータでDEBUGビルドを実行すると、自動的にモック認証が使用されます:
- 「開発用ログイン」ボタンが表示される
- Google/Apple Sign-Inボタンは非表示

#### バックエンド側

`backend/.env`に以下を追加:

```bash
APP_ENV=development
```

これにより:
- Firebase認証の代わりにモック認証が使用される
- トークン`dev-mock-token`で固定UID`dev-mock-user`として認証される
- 本番環境では`APP_ENV`を設定しないこと
- DevAuthMiddlewareはビルドタグ(`!production`)で制御されており、`production`タグ付きビルドでは自動的に無効化される

## 開発ワークフロー

### 日常の開発サイクル

```bash
# バックエンドサーバーを起動
task run
```

注意: データベースはFirestoreを使用しています。ローカル開発時にFirestoreエミュレータを使用する場合:

```bash
# Firestoreエミュレータを起動
firebase emulators:start --only firestore

# 別ターミナルでバックエンドを起動（エミュレータ接続）
FIRESTORE_EMULATOR_HOST=localhost:8080 task run
```

### コード変更後の確認

```bash
# バックエンド
task format
task lint
task test

# iOS
task ios:format-check
task ios:lint
```

iOSテストは現在一時停止中です。必要時のみ手動で `task ios:test` を実行してください。

通常はLefthookにより `pre-commit` / `pre-push` で自動実行されます。
手動実行したい場合:

```bash
task hooks:run
```

### 開発終了時

Firestoreエミュレータを使用している場合は、エミュレータを停止します。

## 利用可能なコマンド（Taskfile）

### 一般

| コマンド | 説明 |
|:---|:---|
| `task --list` | 利用可能なコマンド一覧を表示 |
| `task help` | 利用可能なコマンド一覧を表示 |
| `task hooks:install` | Lefthookフックをインストール |
| `task hooks:run` | pre-commitフックを手動実行 |

### バックエンド

| コマンド | 説明 |
|:---|:---|
| `task setup` | Go依存関係をダウンロード |
| `task clean` | ビルド成果物を削除 |
| `task test` | Goテストを実行 |
| `task test:coverage` | Goカバレッジ計測付きテスト |
| `task lint` | golangci-lintを実行（デッドコード検知を含む） |
| `task format` | Goコードを整形 |
| `task build` | Goバイナリをビルド |
| `task run` | バックエンドサーバーを起動 |

### データベース（Firestore）

| コマンド | 説明 |
|:---|:---|
| `firebase emulators:start --only firestore` | Firestoreエミュレータを起動 |
| `firebase deploy --only firestore:indexes` | Firestoreインデックスをデプロイ |
| `firebase firestore:indexes` | 現在のインデックス一覧を取得 |

### Firestoreインデックス管理

リポジトリ層でFirestoreクエリを追加・変更した場合、複合インデックスが必要になることがあります。

```bash
# インデックスエラーが発生した場合
# 1. エラーメッセージのURLをブラウザで開いてインデックスを作成
# または
# 2. firestore.indexes.jsonを更新してデプロイ
firebase deploy --only firestore:indexes --project utikomi-dev
```

詳細は[.claude/rules/firestore.md](../.claude/rules/firestore.md)を参照してください。

### iOS

| コマンド | 説明 |
|:---|:---|
| `task ios:generate-mocks` | Mockoloでモックを生成 |
| `task ios:test` | iOSのテストを実行 |
| `task ios:test:coverage` | iOSカバレッジ計測付きテスト |
| `task ios:lint` | SwiftLintを実行 |
| `task ios:deadcode` | Swiftデッドコード検知（Periphery）を実行 |
| `task ios:format` | SwiftFormatを実行（コード整形） |
| `task ios:format-check` | SwiftFormatチェック（ローカル検証用） |
| `task ios:clean` | DerivedDataを削除 |
| `task ios:clean-all` | SPMキャッシュを含む完全クリア |
| `task ios:reset-packages` | Package.resolvedを削除して再解決 |

### デプロイ

デプロイはローカル実行のシェルスクリプトを使用します。

```bash
# 開発環境へデプロイ
task deploy:dev

# E2Eテスト込みでデプロイ
./tools/deploy/deploy-dev.sh --run-e2e
```

詳細は[RUNBOOK.md](./RUNBOOK.md#デプロイ)を参照してください。

## テスト手順

### バックエンドテスト

```bash
# 全テストを実行
task test

# Firestoreエミュレータを使用したテスト
firebase emulators:start --only firestore &
FIRESTORE_EMULATOR_HOST=localhost:8080 task test

# 詳細出力
cd backend && go test ./... -v

# カバレッジ
cd backend && go test ./... -cover
```

注意: Firestore Repositoryのテストは`FIRESTORE_EMULATOR_HOST`が設定されていない場合、スキップされます。

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
| Cloud Run | バックエンドAPI |
| Artifact Registry | Dockerイメージ |
| Firestore | データベース |
| Cloud Storage | 画像保存 |
| Firebase Auth | ユーザー認証 |
| Secret Manager | APIキー等のシークレット管理 |
| Workload Identity Federation | Terraform管理（既存構成） |
| GitHub Secrets/Variables | Terraform管理（既存構成） |

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
- [RUNBOOK.md](./RUNBOOK.md) - 運用手順書
- [infrastructure/README.md](../infrastructure/README.md) - Terraformインフラ管理
- [ios/README.md](../ios/README.md) - iOSアプリ開発ガイド
- [.claude/rules/](../.claude/rules/) - 詳細規約
