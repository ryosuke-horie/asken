#!/bin/bash
set -e

# ウチコミ デプロイスクリプト
# 本番サーバー (exe.dev VM) で実行してください

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== ウチコミ デプロイ開始 ==="
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

# migrateコマンドの確認
if ! command -v migrate &> /dev/null; then
    echo "エラー: migrateコマンドが見つかりません"
    echo "インストール方法: DEPLOY.md の「golang-migrate のインストール」を参照"
    exit 1
fi

MIGRATION_DIR="$SCRIPT_DIR/backend/database/migrations"
MIGRATE_URL="postgres://uchikomi:uchikomi@localhost:5432/uchikomi?sslmode=disable"

if [ -d "$MIGRATION_DIR" ]; then
    echo "  マイグレーションディレクトリ: $MIGRATION_DIR"
    if migrate -path "$MIGRATION_DIR" -database "$MIGRATE_URL" up; then
        echo "マイグレーション完了"
    else
        echo "マイグレーション失敗"
        exit 1
    fi
else
    echo "マイグレーションディレクトリが見つかりません: $MIGRATION_DIR"
    exit 1
fi
echo ""

# 4. バックエンドサービス再起動
echo "[4/5] バックエンドサービス再起動..."
sudo systemctl restart uchikomi-backend
echo ""

# 5. フロントエンドサービス再起動
echo "[5/5] フロントエンドサービス再起動..."
sudo systemctl restart uchikomi-frontend
echo ""

# 状態確認
echo "=== サービス状態 ==="
echo ""
echo "--- docker-postgres ---"
systemctl status docker-postgres --no-pager -l | head -10
echo ""
echo "--- uchikomi-backend ---"
systemctl status uchikomi-backend --no-pager -l | head -10
echo ""
echo "--- uchikomi-frontend ---"
systemctl status uchikomi-frontend --no-pager -l | head -10
echo ""

echo "=== デプロイ完了 ==="
echo "フロントエンド: https://uchikomi.exe.xyz:3000"
echo "バックエンドAPI: https://uchikomi.exe.xyz:8080"
