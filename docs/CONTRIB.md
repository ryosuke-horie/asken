# 開発者ガイド（CONTRIB）

最終更新: 2026-01-31

このドキュメントは、ウチコミプロジェクトの開発ワークフロー、利用可能なコマンド、環境セットアップ、テスト手順を説明します。

## 前提条件

| ツール | バージョン | 用途 |
|:---|:---|:---|
| Go | 1.23以上 | バックエンド開発 |
| Node.js | 18以上 | フロントエンド開発 |
| Docker / Docker Compose | 最新 | PostgreSQL起動 |
| Task | 3.x | タスクランナー |
| golangci-lint | 最新 | Goリント |

## 環境セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/ryosuke-horie/uchikomi.git
cd uchikomi
```

### 2. 依存関係のインストール

```bash
# Go依存関係
task setup

# フロントエンド依存関係
task frontend:install
```

### 3. 環境変数の設定

#### バックエンド

`backend/.env`を作成（`backend/.env.example`を参考）:

| 環境変数 | 説明 | 例 |
|:---|:---|:---|
| DATABASE_URL | PostgreSQL接続文字列 | `postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable` |

#### フロントエンド

`frontend/.env.local`を作成:

| 環境変数 | 説明 | 例 |
|:---|:---|:---|
| NEXT_PUBLIC_API_URL | バックエンドAPIのURL | `http://localhost:8080`（ローカル）<br>`https://utikomi.exe.xyz:8080`（本番） |

## 開発ワークフロー

### 日常の開発サイクル

```bash
# 1. DBを起動
task db-up

# 2. テストデータを投入（必要に応じて）
task db-seed

# 3. バックエンドサーバーを起動（ターミナル1）
task run

# 4. フロントエンドサーバーを起動（ターミナル2）
task frontend:dev

# 5. ブラウザでアクセス
open http://localhost:3000
```

### コード変更後の確認

```bash
# バックエンド
task lint
task test

# フロントエンド
task frontend:ci
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

### フロントエンド

| コマンド | 説明 |
|:---|:---|
| `task frontend:install` | npm依存関係をインストール |
| `task frontend:lint` | ESLintを実行 |
| `task frontend:knip` | 未使用コードをチェック |
| `task frontend:depcheck` | 未使用依存をチェック |
| `task frontend:test` | Vitestを実行 |
| `task frontend:build` | 本番ビルド |
| `task frontend:dev` | 開発サーバーを起動 |
| `task frontend:e2e` | Playwrightテストを実行 |
| `task frontend:e2e:ui` | PlaywrightをUIモードで実行 |
| `task frontend:ci` | CI用の全チェックを実行 |

### データベース

| コマンド | 説明 |
|:---|:---|
| `task db-up` | PostgreSQLを起動 |
| `task db-down` | PostgreSQLを停止 |
| `task db-seed` | テストデータを投入 |
| `task db-clean` | DBをリセット（ボリューム削除+再起動） |

### デプロイ

| コマンド | 説明 |
|:---|:---|
| `task deploy` | 本番デプロイ（本番サーバーでのみ実行） |

## フロントエンドnpmスクリプト

`frontend/package.json`で定義:

| スクリプト | 説明 |
|:---|:---|
| `npm run dev` | 開発サーバーを起動 |
| `npm run build` | 本番ビルド |
| `npm start` | 本番サーバーを起動 |
| `npm run lint` | ESLintを実行 |
| `npm run lint:fix` | ESLintで自動修正 |
| `npm run format` | Prettierでフォーマット |
| `npm run format:check` | フォーマットをチェック |
| `npm run typecheck` | TypeScriptの型チェック |
| `npm test` | Vitestを実行 |
| `npm run test:watch` | Vitestをウォッチモードで実行 |
| `npm run test:coverage` | カバレッジレポートを生成 |
| `npm run e2e` | Playwrightテストを実行 |
| `npm run e2e:ui` | PlaywrightをUIモードで実行 |
| `npm run knip` | 未使用コードをチェック |
| `npm run depcheck` | 未使用依存をチェック |

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

### フロントエンドテスト

```bash
# ユニットテスト
task frontend:test

# ウォッチモード（開発中に便利）
cd frontend && npm run test:watch

# カバレッジレポート
cd frontend && npm run test:coverage
```

### E2Eテスト

```bash
# ヘッドレスモードで実行
task frontend:e2e

# UIモードで実行（デバッグ用）
task frontend:e2e:ui
```

E2Eテストの前提:
- バックエンドサーバーが起動している（`task run`）
- DBにテストデータが投入されている（`task db-seed`）

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
refactor: NutritionDisplayコンポーネントを分割
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
- [.claude/rules/](../.claude/rules/) - 詳細規約
