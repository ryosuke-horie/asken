# EDG-323 統合テストガイド

非同期分析アーキテクチャの統合テスト手順を説明します。

## 前提条件

- Docker（PostgreSQL起動用）
- Go 1.23+
- Node.js 18+
- Gemini CLI設定済み

## テスト手順

### 1. PostgreSQL起動

```bash
# プロジェクトルートで実行
docker-compose up -d postgres

# 接続確認
docker-compose ps
```

### 2. データベースマイグレーション

```bash
# PostgreSQLコンテナに接続
docker exec -it asken-postgres psql -U asken -d asken

# マイグレーション実行（SQLファイルの内容を手動実行）
\i /docker-entrypoint-initdb.d/001_create_analysis_tables.sql

# テーブル確認
\dt

# 期待される出力:
#  Schema |       Name        | Type  | Owner
# --------+-------------------+-------+-------
#  public | analysis_requests | table | asken
#  public | analysis_results  | table | asken

# 終了
\q
```

**または、コンテナ再作成で自動マイグレーション:**

```bash
docker-compose down
docker-compose up -d postgres
```

### 3. バックエンド起動

```bash
# Terminal 1: バックエンド
cd backend
export DATABASE_URL="postgres://asken:asken@localhost:5432/asken?sslmode=disable"
go run cmd/server/main.go
```

**期待されるログ:**
```
Database connection established
Server starting on :8080
Analysis worker started with interval: 5s
No pending requests found
```

### 4. フロントエンド起動

```bash
# Terminal 2: フロントエンド
cd frontend
npm run dev
```

### 5. エンドツーエンドテスト

#### シナリオ1: 正常系（アップロード → ポーリング → 完了）

1. http://localhost:3001 にアクセス
2. テスト画像を選択（JPEG/PNG）
3. 「アップロードして分析」をクリック
4. **確認項目:**
   - ✅ ステータスメッセージが表示される
     - 「アップロード中...」
     - 「分析リクエストを受け付けました...」
     - 「分析処理中です...」
     - 「分析が完了しました」
   - ✅ 約2分後に結果が表示される
   - ✅ 栄養素情報が正しく表示される

5. **バックエンドログ確認:**
```
Analysis worker started with interval: 5s
File saved permanently to: uploads/xxx.jpg
Analysis request created with ID: xxx
Found 1 pending requests
Processing request: xxx
Status updated to processing
Analysis completed for request: xxx
Result saved successfully
Image file removed: uploads/xxx.jpg
```

6. **データベース確認:**
```bash
docker exec -it asken-postgres psql -U asken -d asken -c "SELECT id, status FROM analysis_requests ORDER BY created_at DESC LIMIT 1;"
# 期待: status = 'completed'

docker exec -it asken-postgres psql -U asken -d asken -c "SELECT total_calories FROM analysis_results ORDER BY created_at DESC LIMIT 1;"
# 期待: カロリー値が表示される
```

#### シナリオ2: リロード復旧

1. 画像をアップロード
2. 「分析処理中です...」が表示されている間に**ブラウザをリロード（F5）**
3. **確認項目:**
   - ✅ リロード後も「分析処理中です...」が再表示される
   - ✅ ポーリングが自動再開される
   - ✅ 分析完了後に結果が表示される

4. **localStorage確認（開発者ツール）:**
```javascript
// Console タブで実行
localStorage.getItem('asken_analysis_id')
// 期待: analysis_id が保存されている（処理中）
// 完了後: null（自動削除される）
```

#### シナリオ3: エラーハンドリング

**3-1. 無効なファイル形式**
1. `.txt`ファイルをアップロード
2. **確認項目:**
   - ✅ 「サポートされていないファイル形式です」エラーが表示される

**3-2. バックエンド停止**
1. バックエンドを停止（Ctrl+C）
2. 画像をアップロード
3. **確認項目:**
   - ✅ ネットワークエラーが表示される

**3-3. データベース停止**
1. `docker-compose stop postgres`
2. バックエンド再起動を試みる
3. **確認項目:**
   - ✅ `Failed to connect to database` エラーが表示される
   - ✅ サーバーが起動しない

### 6. パフォーマンステスト

#### ポーリング間隔確認

1. ブラウザの開発者ツール → Networkタブを開く
2. 画像をアップロード
3. **確認項目:**
   - ✅ GET `/api/analyze/:id` が約2秒間隔で呼ばれている
   - ✅ 完了後にポーリングが停止する

#### 複数リクエスト

1. 複数の画像を連続でアップロード（3枚程度）
2. **確認項目:**
   - ✅ それぞれのanalysis_idが発行される
   - ✅ ワーカーが順次処理する（ログ確認）
   - ✅ 古いリクエストから処理される（created_at昇順）

**データベース確認:**
```bash
docker exec -it asken-postgres psql -U asken -d asken -c "SELECT id, status, created_at FROM analysis_requests ORDER BY created_at DESC LIMIT 5;"
```

### 7. クリーンアップ

```bash
# バックエンド停止（Ctrl+C）
# フロントエンド停止（Ctrl+C）

# PostgreSQL停止
docker-compose down

# データボリューム削除（オプション）
docker-compose down -v

# アップロードファイル削除
rm -rf backend/uploads/*
```

## トラブルシューティング

### PostgreSQL接続エラー

```bash
# エラー: Failed to connect to database
# 解決策:
docker-compose ps  # PostgreSQLが起動しているか確認
docker-compose logs postgres  # ログ確認
```

### マイグレーション未実行

```bash
# エラー: relation "analysis_requests" does not exist
# 解決策:
docker-compose down
docker-compose up -d postgres
# マイグレーションが自動実行される
```

### ワーカー起動失敗

```bash
# エラー: Worker not starting
# 解決策: ログを確認
# 期待されるログ: "Analysis worker started with interval: 5s"
```

### フロントエンドCORSエラー

```bash
# エラー: Access to fetch... has been blocked by CORS policy
# 解決策: バックエンドが http://localhost:8080 で起動していることを確認
# CORS設定が http://localhost:3001 を許可していることを確認
```

## 成功基準

- ✅ 全ユニットテスト成功（`go test ./... -v`）
- ✅ シナリオ1: 正常系が動作
- ✅ シナリオ2: リロード復旧が動作
- ✅ シナリオ3: エラーハンドリングが適切
- ✅ データベースにデータが正しく保存される
- ✅ localStorageの保存・削除が正しく動作
- ✅ ポーリング間隔が2秒
- ✅ グレースフルシャットダウンが動作（Ctrl+C）

## 参考: データベーススキーマ確認

```sql
-- analysis_requests テーブル構造
\d analysis_requests

-- analysis_results テーブル構造
\d analysis_results

-- インデックス確認
\di

-- データ件数確認
SELECT COUNT(*) FROM analysis_requests;
SELECT COUNT(*) FROM analysis_results;

-- ステータス別集計
SELECT status, COUNT(*) FROM analysis_requests GROUP BY status;
```
