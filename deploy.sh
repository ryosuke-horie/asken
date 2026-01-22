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
echo "[1/5] git pull で最新コードを取得..."
git pull origin main
echo ""

# 2. PostgreSQLサービス確認
echo "[2/5] PostgreSQL サービス確認..."
if systemctl is-active --quiet docker-postgres; then
    echo "PostgreSQL: 稼働中"
else
    echo "PostgreSQL: 停止中 -> 起動します"
    sudo systemctl start docker-postgres
    echo "PostgreSQL の起動を待機中..."
    sleep 5
fi
echo ""

# 3. DBマイグレーション実行
echo "[3/5] DBマイグレーション実行..."
MIGRATION_DIR="$SCRIPT_DIR/backend/database/migrations"
if [ -d "$MIGRATION_DIR" ]; then
    for sql_file in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$sql_file" ]; then
            echo "  適用中: $(basename "$sql_file")"
            docker exec -i asken-postgres psql -U asken -d asken < "$sql_file" 2>&1 | grep -v "already exists" || true
        fi
    done
    echo "マイグレーション完了"
else
    echo "マイグレーションディレクトリが見つかりません: $MIGRATION_DIR"
fi
echo ""

# 4. バックエンドサービス再起動
echo "[4/5] バックエンドサービス再起動..."
sudo systemctl restart asken-backend
echo ""

# 5. フロントエンドサービス再起動
echo "[5/5] フロントエンドサービス再起動..."
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
