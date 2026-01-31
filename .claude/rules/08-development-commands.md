# 開発環境とコマンド操作

このプロジェクトでは`Taskfile`で開発コマンドを統一している。
コマンド実行やサーバー起動を求められた場合、以下を参照すること。

## 開発環境の立ち上げ

```bash
# 1. DBを起動
task db-up

# 2. テストデータを投入（必要に応じて）
task db-seed

# 3. バックエンドサーバーを起動
task run
```

## コード変更後の確認

```bash
# リントとテストを実行
task lint
task test
```

## DBのリセット

```bash
# ボリューム削除して再起動
task db-clean

# テストデータを再投入
task db-seed
```

## 開発終了時

```bash
task db-down
```

## コマンド一覧

| コマンド | 説明 |
|:---|:---|
| `task --list` | 利用可能なコマンド一覧を表示 |
| `task help` | 利用可能なコマンド一覧を表示 |
| `task setup` | Go依存関係をダウンロード |
| `task clean` | ビルド成果物を削除 |
| `task test` | Goテストを実行 |
| `task lint` | Goリントを実行 |
| `task build` | Goバイナリをビルド |
| `task run` | バックエンドサーバーを起動 |
| `task db-up` | PostgreSQLを起動 |
| `task db-down` | PostgreSQLを停止 |
| `task db-seed` | テストデータを投入 |
| `task db-clean` | DBをリセット |
| `task deploy` | 本番デプロイ（本番サーバーで実行） |
| `task ios:generate-mocks` | Mockoloでモックを生成 |
| `task ios:test` | iOSのテストを実行 |

## 注意

- **db-seed**は`backend/.env`の環境変数を読み込む
- **deploy**は本番サーバーでのみ実行すること
