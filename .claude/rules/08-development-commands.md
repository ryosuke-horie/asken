# 開発環境とコマンド操作

このプロジェクトでは`Makefile`で開発コマンドを統一している。
コマンド実行やサーバー起動を求められた場合、以下を参照すること。

## 開発環境の立ち上げ

```bash
# 1. DBを起動
make db-up

# 2. テストデータを投入（必要に応じて）
make db-seed

# 3. バックエンドサーバーを起動
make run

# 4. フロントエンドサーバーを起動（別ターミナル）
cd frontend && npm run dev
```

## コード変更後の確認

```bash
# リントとテストを実行
make lint
make test
```

## DBのリセット

```bash
# ボリューム削除して再起動
make db-clean

# テストデータを再投入
make db-seed
```

## 開発終了時

```bash
make db-down
```

## コマンド一覧

| コマンド | 説明 |
|:---|:---|
| `make help` | 利用可能なコマンド一覧を表示 |
| `make setup` | Go依存関係をダウンロード |
| `make clean` | ビルド成果物を削除 |
| `make test` | Goテストを実行 |
| `make lint` | Goリントを実行 |
| `make build` | Goバイナリをビルド |
| `make run` | バックエンドサーバーを起動 |
| `make db-up` | PostgreSQLを起動 |
| `make db-down` | PostgreSQLを停止 |
| `make db-seed` | テストデータを投入 |
| `make db-clean` | DBをリセット |
| `make deploy` | 本番デプロイ（本番サーバーで実行） |

## 注意

- **Frontend**は`Makefile`に含まれていない。`cd frontend && npm run xxx`で直接実行
- **db-seed**は`backend/.env`の環境変数を読み込む
- **deploy**は本番サーバーでのみ実行すること
