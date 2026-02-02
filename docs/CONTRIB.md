# 開発者ガイド（CONTRIB）

最終更新: 2026-01-31

このドキュメントは、ウチコミプロジェクトの開発ワークフロー、利用可能なコマンド、環境セットアップ、テスト手順を説明します。

## 前提条件

| ツール | バージョン | 用途 |
|:---|:---|:---|
| mise | 最新 | ツールバージョン管理 |
| Go | 1.25以上 | バックエンド開発 |
| Docker / Docker Compose | 最新 | PostgreSQL起動 |
| Task | 3.x | タスクランナー |
| golangci-lint | 最新 | Goリント |
| Xcode | 16以上 | iOS開発 |
| Mockolo | 最新 | iOSモック生成 |

## 環境セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/ryosuke-horie/uchikomi.git
cd uchikomi
```

### 2. miseのインストールとツールセットアップ

```bash
# miseをインストール（未インストールの場合）
curl https://mise.run | sh

# プロジェクトのツールをインストール
mise trust
mise install
```

### 3. 依存関係のインストール

```bash
# Go依存関係
task setup
```

### 3. 環境変数の設定

#### バックエンド

`backend/.env`を作成（`backend/.env.example`を参考）:

| 環境変数 | 説明 | 例 |
|:---|:---|:---|
| DATABASE_URL | PostgreSQL接続文字列 | `postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable` |

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

## 関連ドキュメント

- [README.md](../README.md) - プロジェクト概要
- [DEPLOY.md](../DEPLOY.md) - デプロイ手順
- [RUNBOOK.md](./RUNBOOK.md) - 運用手順書
- [ios/README.md](../ios/README.md) - iOSアプリ開発ガイド
- [.claude/rules/](../.claude/rules/) - 詳細規約
