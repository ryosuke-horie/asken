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
cd /home/exedev/asken
```

3. **デプロイスクリプトを実行**

```bash
./deploy.sh
```

スクリプトが以下を自動で実行します:
- `git pull origin main` で最新コードを取得
- PostgreSQL の稼働確認（停止中なら起動）
- バックエンドサービスの再起動
- フロントエンドサービスの再起動
- 各サービスの状態表示

### 手動デプロイ

deploy.sh を使わない場合の手順:

```bash
# 最新コードを取得
cd /home/exedev/asken
git pull origin main

# サービス再起動
sudo systemctl restart asken-backend
sudo systemctl restart asken-frontend

# 状態確認
systemctl status asken-backend
systemctl status asken-frontend
```

## 初回セットアップ

新しい環境に初めてデプロイする場合の手順です。

### 1. 前提条件

- Ubuntu 22.04 以上
- Docker / Docker Compose
- Go 1.23 以上
- Node.js 18 以上
- Git

### 2. リポジトリのクローン

```bash
cd /home/exedev
git clone https://github.com/ryosuke-horie/asken-sub.git asken
cd asken
```

### 3. systemd サービスファイルの配置

```bash
# PostgreSQL用
sudo cp docker-postgres.service /etc/systemd/system/

# バックエンド用
sudo cp backend/asken-backend.service /etc/systemd/system/

# フロントエンド用
sudo cp frontend/asken-frontend.service /etc/systemd/system/

# デーモンをリロード
sudo systemctl daemon-reload
```

### 4. サービスの有効化

```bash
sudo systemctl enable docker-postgres
sudo systemctl enable asken-backend
sudo systemctl enable asken-frontend
```

### 5. サービスの起動

```bash
sudo systemctl start docker-postgres
sudo systemctl start asken-backend
sudo systemctl start asken-frontend
```

## サービス管理コマンド

### 状態確認

```bash
systemctl status docker-postgres
systemctl status asken-backend
systemctl status asken-frontend
```

### ログ確認

```bash
# リアルタイムログ
journalctl -u asken-backend -f
journalctl -u asken-frontend -f

# 直近のログ（100行）
journalctl -u asken-backend -n 100
journalctl -u asken-frontend -n 100
```

### 全サービスの停止

```bash
sudo systemctl stop asken-frontend
sudo systemctl stop asken-backend
sudo systemctl stop docker-postgres
```

### 全サービスの起動

```bash
sudo systemctl start docker-postgres
sudo systemctl start asken-backend
sudo systemctl start asken-frontend
```

## トラブルシューティング

### サービスが起動しない

```bash
# 詳細なエラーログを確認
journalctl -u <サービス名> -n 50 --no-pager

# 例: バックエンドのエラーを確認
journalctl -u asken-backend -n 50 --no-pager
```

### PostgreSQL に接続できない

```bash
# Docker コンテナの状態確認
docker ps

# PostgreSQL コンテナのログ確認
docker logs asken-postgres
```

### フロントエンドのビルドが失敗する

```bash
# 手動でビルドを実行してエラーを確認
cd /home/exedev/asken/frontend
npm install
npm run build
```

### ロールバック

特定のコミットに戻す場合:

```bash
cd /home/exedev/asken

# 戻したいコミットを確認
git log --oneline -10

# 特定のコミットに戻す
git checkout <commit-hash>

# サービス再起動
sudo systemctl restart asken-backend
sudo systemctl restart asken-frontend
```

## URL

- **フロントエンド**: https://asken.exe.xyz:3000
- **バックエンドAPI**: https://asken.exe.xyz:8080
