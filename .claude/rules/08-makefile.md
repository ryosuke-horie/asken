# Makefileコマンド集

## 概要

プロジェクトルートの`Makefile`でbackend/Docker関連の操作を統一的に実行できる。
frontendは対象外（`cd frontend && npm run xxx`で直接実行）。

## コマンド一覧

### 基本

| コマンド | 説明 | 使用タイミング |
|:---|:---|:---|
| `make help` | ターゲット一覧を表示 | 利用可能なコマンドを確認したいとき |
| `make setup` | 依存関係をダウンロード（go mod download） | 初回セットアップ時、go.mod変更後 |
| `make clean` | ビルド成果物を削除（backend/bin/） | ビルドをクリーンにやり直したいとき |

### Backend開発

| コマンド | 説明 | 使用タイミング |
|:---|:---|:---|
| `make test` | Goテストを実行（go test ./...） | コード変更後、PR作成前 |
| `make lint` | Goリントを実行（go vet ./...） | コード変更後、PR作成前 |
| `make build` | Goバイナリをビルド（backend/bin/server） | 本番用バイナリを作成するとき |
| `make run` | バックエンドサーバーを起動 | 開発時のローカル実行 |

### Docker/DB操作

| コマンド | 説明 | 使用タイミング |
|:---|:---|:---|
| `make db-up` | PostgreSQLコンテナを起動 | 開発開始時、DB停止後 |
| `make db-down` | PostgreSQLコンテナを停止 | 開発終了時、リソース解放 |
| `make db-seed` | テストデータを投入 | DBリセット後、テストデータが必要なとき |
| `make db-clean` | DBをリセット（ボリューム削除+再起動） | DBを完全にクリーンにしたいとき |

### デプロイ

| コマンド | 説明 | 使用タイミング |
|:---|:---|:---|
| `make deploy` | deploy.shを実行 | 本番環境へのデプロイ時（本番サーバーで実行） |

## 典型的なワークフロー

### 開発環境の立ち上げ

```bash
make db-up      # PostgreSQL起動
make db-seed    # テストデータ投入（必要に応じて）
make run        # バックエンドサーバー起動
# 別ターミナルで
cd frontend && npm run dev  # フロントエンド起動
```

### コード変更後の確認

```bash
make lint       # リント実行
make test       # テスト実行
```

### DBの完全リセット

```bash
make db-clean   # ボリューム削除+再起動
make db-seed    # テストデータ再投入
```

### 開発終了時

```bash
make db-down    # PostgreSQL停止
```

## 注意事項

- `make db-seed`はbackend/.envの環境変数を読み込んで実行する
- `make deploy`は本番サーバーで実行すること（ローカルでは使用しない）
- frontendのコマンドはMakefileに含まれていない（npm scriptを直接使用）
