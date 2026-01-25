#!/bin/bash
# マイグレーションのupファイルのみを実行するスクリプト

set -e

MIGRATIONS_DIR="/migrations"

echo "Running migrations..."

# upファイルのみを番号順に実行
for file in $(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | sort); do
    echo "Executing: $(basename $file)"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
done

echo "All migrations completed."
