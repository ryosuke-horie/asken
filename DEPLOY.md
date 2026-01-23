# デプロイ手順

本番環境（exe.dev VM）へのデプロイ手順を説明します。

## 概要

- **本番サーバー**: exe.dev VM
- **デプロイ方法**: `deploy.sh` を本番サーバーで実行
- **CI/CD**: 使用しない（手動デプロイ）

## サービス構成

| サービス | 管理方法 | ポート |
|---------|----------|--------|
| PostgreSQL | Docker Compose + systemd | 5432 |
| Backend (Go) | systemd | 8080 |
| Frontend (Next.js) | systemd | 3000 |

## デプロイ手順

### 通常のデプロイ

1. **本番サーバーにSSH接続**

```bash
ssh exedev@exe.dev
```

2. **プロジェクトディレクトリに移動**

```bash
cd /home/exedev/uchikomi
```

3. **デプロイスクリプトを実行**

```bash
./deploy.sh
```

スクリプトが以下を自動で実行します:
- `git pull origin main` で最新コードを取得
- PostgreSQL の稼働確認（停止中なら起動）
- DBマイグレーションの実行（golang-migrate）
- バックエンドサービスの再起動
- フロントエンドサービスの再起動
- 各サービスの状態表示

### 手動デプロイ

deploy.sh を使わない場合の手順:

```bash
# 最新コードを取得
cd /home/exedev/uchikomi
git pull origin main

# DBマイグレーション
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        up

# サービス再起動
sudo systemctl restart uchikomi-backend
sudo systemctl restart uchikomi-frontend

# 状態確認
systemctl status uchikomi-backend
systemctl status uchikomi-frontend
```

## 初回セットアップ

新しい環境に初めてデプロイする場合の手順です。

### 1. 前提条件

- Ubuntu 22.04 以上
- Docker / Docker Compose
- Go 1.23 以上
- Node.js 18 以上
- Git
- golang-migrate

#### golang-migrate のインストール（本番サーバー）

```bash
MIGRATE_VERSION="v4.17.0"
curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
migrate -version
```

### 2. リポジトリのクローン

```bash
cd /home/exedev
git clone https://github.com/ryosuke-horie/uchikomi.git uchikomi
cd uchikomi
```

### 3. systemd サービスファイルの配置

```bash
# PostgreSQL用
sudo cp docker-postgres.service /etc/systemd/system/

# バックエンド用
sudo cp backend/uchikomi-backend.service /etc/systemd/system/

# フロントエンド用
sudo cp frontend/uchikomi-frontend.service /etc/systemd/system/

# デーモンをリロード
sudo systemctl daemon-reload
```

### 4. サービスの有効化

```bash
sudo systemctl enable docker-postgres
sudo systemctl enable uchikomi-backend
sudo systemctl enable uchikomi-frontend
```

### 5. サービスの起動

```bash
sudo systemctl start docker-postgres
sudo systemctl start uchikomi-backend
sudo systemctl start uchikomi-frontend
```

### 6. 既存DBへの golang-migrate 初回適用

既にマイグレーションが適用されているDBに対して、golang-migrateの管理に移行する場合:

```bash
# version 3まで適用済みとして登録（schema_migrationsテーブルが作成される）
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        force 3

# バージョン確認
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        version
# 出力: 3
```

**注意**: この手順は既存の本番環境で1回だけ実行します。新規環境では不要です。

## サービス管理コマンド

### 状態確認

```bash
systemctl status docker-postgres
systemctl status uchikomi-backend
systemctl status uchikomi-frontend
```

### ログ確認

```bash
# リアルタイムログ
journalctl -u uchikomi-backend -f
journalctl -u uchikomi-frontend -f

# 直近のログ（100行）
journalctl -u uchikomi-backend -n 100
journalctl -u uchikomi-frontend -n 100
```

### 全サービスの停止

```bash
sudo systemctl stop uchikomi-frontend
sudo systemctl stop uchikomi-backend
sudo systemctl stop docker-postgres
```

### 全サービスの起動

```bash
sudo systemctl start docker-postgres
sudo systemctl start uchikomi-backend
sudo systemctl start uchikomi-frontend
```

## トラブルシューティング

### サービスが起動しない

```bash
# 詳細なエラーログを確認
journalctl -u <サービス名> -n 50 --no-pager

# 例: バックエンドのエラーを確認
journalctl -u uchikomi-backend -n 50 --no-pager
```

### PostgreSQL に接続できない

```bash
# Docker コンテナの状態確認
docker ps

# PostgreSQL コンテナのログ確認
docker logs uchikomi-postgres
```

### フロントエンドのビルドが失敗する

```bash
# 手動でビルドを実行してエラーを確認
cd /home/exedev/uchikomi/frontend
npm install
npm run build
```

### API接続エラー（「サーバーに接続できません」）

フロントエンドからバックエンドAPIに接続できない場合、以下を確認してください：

1. **環境変数の確認**: `frontend/uchikomi-frontend.service`の`NEXT_PUBLIC_API_URL`が正しいドメインを指しているか確認

```bash
# 現在の設定を確認
grep NEXT_PUBLIC_API_URL /home/exedev/uchikomi/frontend/uchikomi-frontend.service

# 正しい設定例
# Environment=NEXT_PUBLIC_API_URL=https://asken.exe.xyz:8080
```

2. **CORS設定の確認**: バックエンドの`backend/cmd/server/main.go`でフロントエンドのオリジンが許可されているか確認

3. **サービス再起動**: 設定変更後はサービスの再起動が必要

```bash
sudo systemctl daemon-reload
sudo systemctl restart uchikomi-frontend
```

### ロールバック

特定のコミットに戻す場合:

```bash
cd /home/exedev/uchikomi

# 戻したいコミットを確認
git log --oneline -10

# 特定のコミットに戻す
git checkout <commit-hash>

# サービス再起動
sudo systemctl restart uchikomi-backend
sudo systemctl restart uchikomi-frontend
```

## URL

- **フロントエンド**: https://asken.exe.xyz:3000
- **バックエンドAPI**: https://asken.exe.xyz:8080
