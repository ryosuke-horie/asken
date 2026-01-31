# 運用手順書（RUNBOOK）

最終更新: 2026-01-31

本番環境（exe.dev VM）の運用手順、監視、トラブルシューティング、ロールバック手順を説明します。

## サービス構成

| サービス | 管理方法 | ポート | 説明 |
|:---|:---|:---|:---|
| PostgreSQL | Docker Compose + systemd | 5432 | データベース |
| Backend (Go) | systemd | 8080 | APIサーバー |
| Frontend (Next.js) | systemd | 3000 | Webアプリケーション |

### URL

| 環境 | フロントエンド | バックエンドAPI |
|:---|:---|:---|
| 本番 | https://utikomi.exe.xyz:3000 | https://utikomi.exe.xyz:8080 |
| ローカル | http://localhost:3000 | http://localhost:8080 |

## デプロイ手順

### 通常のデプロイ

```bash
# 1. 本番サーバーにSSH接続
ssh exedev@exe.dev

# 2. プロジェクトディレクトリに移動
cd /home/exedev/uchikomi

# 3. デプロイスクリプトを実行
./deploy.sh
```

`deploy.sh`が自動で実行する処理:
- `git pull origin main`で最新コードを取得
- PostgreSQLの稼働確認（停止中なら起動）
- DBマイグレーションの実行（golang-migrate）
- バックエンドサービスの再起動
- フロントエンドサービスの再起動
- 各サービスの状態表示

### 手動デプロイ

deploy.shを使わない場合:

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

## サービス管理コマンド

### 状態確認

```bash
# 個別サービス
systemctl status docker-postgres
systemctl status uchikomi-backend
systemctl status uchikomi-frontend

# 全サービス一括
systemctl status docker-postgres uchikomi-backend uchikomi-frontend
```

### ログ確認

```bash
# リアルタイムログ
journalctl -u uchikomi-backend -f
journalctl -u uchikomi-frontend -f

# 直近のログ（100行）
journalctl -u uchikomi-backend -n 100
journalctl -u uchikomi-frontend -n 100

# 特定時間以降のログ
journalctl -u uchikomi-backend --since "2026-01-30 10:00:00"

# エラーのみ
journalctl -u uchikomi-backend -p err
```

### サービス起動/停止

```bash
# 全サービス起動
sudo systemctl start docker-postgres
sudo systemctl start uchikomi-backend
sudo systemctl start uchikomi-frontend

# 全サービス停止
sudo systemctl stop uchikomi-frontend
sudo systemctl stop uchikomi-backend
sudo systemctl stop docker-postgres

# 再起動
sudo systemctl restart uchikomi-backend
sudo systemctl restart uchikomi-frontend
```

## 監視とアラート

### ヘルスチェック

```bash
# バックエンドAPI
curl -s http://localhost:8080/health || echo "Backend is down"

# フロントエンド
curl -s http://localhost:3000 || echo "Frontend is down"

# PostgreSQL
docker exec uchikomi-postgres pg_isready -U uchikomi || echo "Database is down"
```

### リソース監視

```bash
# メモリ使用量
free -h

# ディスク使用量
df -h

# プロセス確認
ps aux | grep -E "(go|node|postgres)"

# Dockerコンテナ状態
docker ps
docker stats --no-stream
```

### ログ監視（エラー検出）

```bash
# バックエンドエラー
journalctl -u uchikomi-backend -p err --since "1 hour ago"

# フロントエンドエラー
journalctl -u uchikomi-frontend -p err --since "1 hour ago"
```

## よくある問題と修正

### 1. サービスが起動しない

```bash
# 詳細なエラーログを確認
journalctl -u <サービス名> -n 50 --no-pager

# 例: バックエンドのエラーを確認
journalctl -u uchikomi-backend -n 50 --no-pager
```

### 2. PostgreSQLに接続できない

```bash
# Dockerコンテナの状態確認
docker ps

# PostgreSQLコンテナのログ確認
docker logs uchikomi-postgres

# 手動で起動
sudo systemctl restart docker-postgres
sleep 5
docker ps
```

### 3. フロントエンドのビルドが失敗する

```bash
# 手動でビルドを実行してエラーを確認
cd /home/exedev/uchikomi/frontend
npm install
npm run build
```

### 4. API接続エラー（「サーバーに接続できません」）

1. 環境変数の確認:
```bash
grep NEXT_PUBLIC_API_URL /home/exedev/uchikomi/frontend/uchikomi-frontend.service
```

2. CORS設定の確認（`backend/cmd/server/main.go`でオリジンが許可されているか）

3. サービス再起動:
```bash
sudo systemctl daemon-reload
sudo systemctl restart uchikomi-frontend
```

### 5. マイグレーションエラー

```bash
# 現在のバージョンを確認
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        version

# dirty状態の解消（問題のあるバージョンを強制設定）
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        force <version>
```

### 6. ディスク容量不足

```bash
# Dockerの不要イメージ削除
docker system prune -a

# ログのローテーション
sudo journalctl --vacuum-time=7d
```

## ロールバック手順

### コードのロールバック

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

### データベースのロールバック

```bash
# 1つ前のバージョンに戻す
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        down 1

# 特定のバージョンまで戻す
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        goto <version>
```

### 完全なロールバック（緊急時）

```bash
# 1. 直前のデプロイコミットに戻す
cd /home/exedev/uchikomi
git log --oneline -5
git checkout <previous-commit-hash>

# 2. DBマイグレーションを戻す（必要な場合）
migrate -path backend/database/migrations \
        -database "postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable" \
        down 1

# 3. サービス再起動
sudo systemctl restart uchikomi-backend
sudo systemctl restart uchikomi-frontend

# 4. 動作確認
curl -s http://localhost:8080/health
curl -s http://localhost:3000
```

## 初回セットアップ

新しい環境に初めてデプロイする場合の手順は[DEPLOY.md](../DEPLOY.md#初回セットアップ)を参照してください。

## 関連ドキュメント

- [README.md](../README.md) - プロジェクト概要
- [DEPLOY.md](../DEPLOY.md) - デプロイ詳細手順
- [CONTRIB.md](./CONTRIB.md) - 開発者ガイド
