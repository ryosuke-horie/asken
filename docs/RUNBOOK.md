# 運用手順書（RUNBOOK）

最終更新: 2026-02-02

本番環境の運用手順、監視、トラブルシューティング、ロールバック手順を説明します。

## アーキテクチャ

### 現行環境（exe.dev VM）

| サービス | 管理方法 | ポート | 説明 |
|:---|:---|:---|:---|
| PostgreSQL | Docker Compose + systemd | 5432 | データベース |
| Backend (Go) | systemd | 8080 | APIサーバー |

### 新アーキテクチャ（GCPサーバレス）- 移行中

| サービス | GCPサービス | 説明 |
|:---|:---|:---|
| データベース | Firestore | NoSQLデータベース |
| ストレージ | Cloud Storage | 画像保存 |
| 認証 | Firebase Auth | ユーザー認証 |
| AI | Gemini API | 画像認識・栄養素分析 |

インフラはTerraformで管理: [infrastructure/README.md](../infrastructure/README.md)

### URL

| 環境 | バックエンドAPI |
|:---|:---|
| 本番（現行） | https://utikomi.exe.xyz:8080 |
| ローカル | http://localhost:8080 |

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

# 状態確認
systemctl status uchikomi-backend
```

## サービス管理コマンド

### 状態確認

```bash
# 個別サービス
systemctl status docker-postgres
systemctl status uchikomi-backend

# 全サービス一括
systemctl status docker-postgres uchikomi-backend
```

### ログ確認

```bash
# リアルタイムログ
journalctl -u uchikomi-backend -f

# 直近のログ（100行）
journalctl -u uchikomi-backend -n 100

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

# 全サービス停止
sudo systemctl stop uchikomi-backend
sudo systemctl stop docker-postgres

# 再起動
sudo systemctl restart uchikomi-backend
```

## 監視とアラート

### ヘルスチェック

```bash
# バックエンドAPI
curl -s http://localhost:8080/health || echo "Backend is down"

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

### 3. マイグレーションエラー

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

### 4. ディスク容量不足

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

# 4. 動作確認
curl -s http://localhost:8080/health
```

## 初回セットアップ

新しい環境に初めてデプロイする場合の手順は[DEPLOY.md](../DEPLOY.md#初回セットアップ)を参照してください。

## GCPインフラ管理（Terraform）

### インフラの変更

```bash
cd infrastructure/environments/dev

# 変更内容を確認
terraform plan

# 変更を適用
terraform apply
```

### リソースの状態確認

```bash
# Terraform状態を確認
terraform state list

# 特定リソースの詳細
terraform state show <resource_name>
```

### よくある問題

#### Firebase権限エラー

```
Error: Error creating Database: googleapi: Error 403
```

サービスアカウントに`roles/firebase.admin`が付与されているか確認:

```bash
gcloud projects get-iam-policy utikomi-dev \
  --flatten="bindings[].members" \
  --filter="bindings.members:terraform-admin@utikomi-dev.iam.gserviceaccount.com"
```

#### Terraform認証エラー

`GOOGLE_APPLICATION_CREDENTIALS`が正しく設定されているか確認:

```bash
echo $GOOGLE_APPLICATION_CREDENTIALS
# miseを使用している場合は自動設定される
```

詳細は[infrastructure/README.md](../infrastructure/README.md#トラブルシューティング)を参照。

## 関連ドキュメント

- [README.md](../README.md) - プロジェクト概要
- [DEPLOY.md](../DEPLOY.md) - デプロイ詳細手順
- [CONTRIB.md](./CONTRIB.md) - 開発者ガイド
- [infrastructure/README.md](../infrastructure/README.md) - Terraformインフラ管理
