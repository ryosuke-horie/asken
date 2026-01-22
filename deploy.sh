#!/bin/bash
set -e

# Asken デプロイスクリプト
# 本番サーバー (exe.dev VM) で実行してください

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Asken デプロイ開始 ==="
echo "ディレクトリ: $SCRIPT_DIR"
echo ""

# 1. 最新コードを取得
echo "[1/4] git pull で最新コードを取得..."
git pull origin main
echo ""

# 2. PostgreSQLサービス確認
echo "[2/4] PostgreSQL サービス確認..."
if systemctl is-active --quiet docker-postgres; then
    echo "PostgreSQL: 稼働中"
else
    echo "PostgreSQL: 停止中 -> 起動します"
    sudo systemctl start docker-postgres
fi
echo ""

# 3. バックエンドサービス再起動
echo "[3/4] バックエンドサービス再起動..."
sudo systemctl restart asken-backend
echo ""

# 4. フロントエンドサービス再起動
echo "[4/4] フロントエンドサービス再起動..."
sudo systemctl restart asken-frontend
echo ""

# 状態確認
echo "=== サービス状態 ==="
echo ""
echo "--- docker-postgres ---"
systemctl status docker-postgres --no-pager -l | head -10
echo ""
echo "--- asken-backend ---"
systemctl status asken-backend --no-pager -l | head -10
echo ""
echo "--- asken-frontend ---"
systemctl status asken-frontend --no-pager -l | head -10
echo ""

echo "=== デプロイ完了 ==="
echo "フロントエンド: https://asken.exe.xyz:3000"
echo "バックエンドAPI: https://asken.exe.xyz:8080"
