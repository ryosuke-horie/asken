# asken - カロリー計算アプリ

画像から食事内容を判定し、カロリーと栄養素を計算するMVPアプリケーション。

## 概要

**asken**は、Gemini API（Gemini 3）を活用して、食事の画像から自動的に食材を認識し、カロリーと栄養素を計算するWebアプリケーションです。

### 主要機能

- 📸 **画像認識**: 食事の画像をアップロードして食材を自動判定
- 🔍 **2ステップアプローチ**:
  - Step 1: 食材分類（食材名と推定量を抽出）
  - Step 2: 栄養素計算（カロリー、タンパク質、脂質、炭水化物を算出）
- 📊 **栄養素表示**: テーブル形式で見やすく表示

## 技術スタック

### バックエンド

- **言語**: Go 1.23
- **AI**: Gemini CLI（Gemini 3 API）
- **テスト**: testify
- **依存関係**:
  - github.com/google/uuid
  - github.com/stretchr/testify

### フロントエンド

- **フレームワーク**: Next.js 14
- **言語**: TypeScript
- **UIライブラリ**: React 18

## ディレクトリ構成

```
asken/
├── backend/                    # Goバックエンド
│   ├── cmd/server/            # HTTPサーバー
│   ├── internal/              # 内部パッケージ
│   │   ├── handler/           # HTTPハンドラー
│   │   └── service/           # ビジネスロジック
│   └── pkg/gemini/            # Gemini CLI クライアント
├── frontend/                  # Next.jsフロントエンド
│   ├── app/                   # App Router
│   ├── components/            # Reactコンポーネント
│   └── types/                 # TypeScript型定義
└── experiments/               # 実験コード
    └── gemini-cli/            # Gemini CLI検証実験
```

## セットアップ

### 前提条件

- Go 1.23以上
- Node.js 18以上
- Gemini CLI（インストール手順は[公式ドキュメント](https://ai.google.dev/gemini-api/docs)参照）

### バックエンドのセットアップ

```bash
cd backend
go mod download
```

### フロントエンドのセットアップ

```bash
cd frontend
npm install
```

## 起動方法

### 本番環境（systemd + Docker 利用）

今回、以下の構成で本番運用できるようにしました：

- PostgreSQL: Docker Compose + systemd
- バックエンド API: Go サーバー（systemd サービス）
- フロントエンド: Next.js 本番ビルド + `next start`（systemd サービス）

#### 1. PostgreSQL コンテナの常時起動

Docker Compose 定義は `docker-compose.yml` にあり、Postgres サービスは `postgres` です。

systemd ユニット `/etc/systemd/system/docker-postgres.service` を作成し、起動・自動起動設定を行いました。

```ini
[Unit]
Description=Asken Postgres via Docker Compose
After=network-online.target docker.service
Requires=docker.service

[Service]
Type=oneshot
WorkingDirectory=/home/exedev/asken
ExecStart=/usr/bin/docker compose up -d postgres
ExecStop=/usr/bin/docker compose stop postgres
RemainAfterExit=yes
User=exedev

[Install]
WantedBy=multi-user.target
```

有効化と起動:

```bash
sudo systemctl enable docker-postgres
sudo systemctl start docker-postgres

# 状態確認
systemctl status docker-postgres
```

#### 2. バックエンド API サービス

Go サーバーを systemd 管理にしました。ユニットは `/etc/systemd/system/asken-backend.service` です。

```ini
[Unit]
Description=Asken Backend API Service
After=network-online.target docker.service
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/home/exedev/asken/backend
Environment=DATABASE_URL=postgres://asken:asken@localhost:5432/asken?sslmode=disable
ExecStart=/usr/local/go/bin/go run ./cmd/server
Restart=on-failure
RestartSec=5
User=exedev

[Install]
WantedBy=multi-user.target
```

有効化と起動:

```bash
sudo systemctl enable asken-backend
sudo systemctl start asken-backend

# 状態確認
systemctl status asken-backend
```

API は `http://localhost:8080` / `https://asken.exe.xyz:8080` で待ち受けます。

#### 3. フロントエンド（本番モード）

Next.js アプリを `npm run build` + `npm start` で本番実行するように systemd 化しました。

`/etc/systemd/system/asken-frontend.service`:

```ini
[Unit]
Description=Asken Frontend Next.js Service (Production)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/home/exedev/asken/frontend
ExecStartPre=/usr/bin/npm install
ExecStartPre=/usr/bin/npm run build
ExecStart=/usr/bin/npm start
Restart=on-failure
RestartSec=5
User=exedev
Environment=PORT=3000

[Install]
WantedBy=multi-user.target
```

有効化と起動:

```bash
sudo systemctl enable asken-frontend
sudo systemctl restart asken-frontend

# 状態確認
systemctl status asken-frontend
```

フロントエンドは `http://localhost:3000` / `https://asken.exe.xyz:3000` からアクセス可能です。

#### 4. 全体の起動・停止

```bash
# 起動
sudo systemctl start docker-postgres
sudo systemctl start asken-backend
sudo systemctl start asken-frontend

# 停止
sudo systemctl stop asken-frontend
sudo systemctl stop asken-backend
sudo systemctl stop docker-postgres

# 再起動
sudo systemctl restart docker-postgres
sudo systemctl restart asken-backend
sudo systemctl restart asken-frontend
```

## CI/CD（GitHub Actions self-hosted runner）

exe.dev VM 上で GitHub Actions の self-hosted runner を常駐させることで、この環境でビルド・テスト・デプロイを自動化できます。公式ガイド: [Running a self-hosted GitHub Actions Runner](https://exe.dev/docs/use-case-gh-action-runner.md)

### 推奨フロー
1. **専用ユーザーを作成（任意）**: `asken-runner` などを作成し runner を実行することで権限を分離。
2. **GitHub で runner を登録**: リポジトリ (または組織) の `Settings > Actions > Runners > New self-hosted runner` で Linux x64 を選択し、表示されるコマンドを VM 上で実行。
3. **インストール**: 例) `/home/exedev/actions-runner` に展開し、`./config.sh --url <repo-url> --token <token> --labels self-hosted,asken` を実行。
4. **systemd サービス化**: `./svc.sh install` でも良いが、exe.dev の標準に合わせて以下のようなユニットを `/etc/systemd/system/asken-runner.service` として配置し、`sudo systemctl enable --now asken-runner`。

```ini
[Unit]
Description=Asken GitHub Actions Runner
After=network-online.target

[Service]
Type=simple
User=asken-runner
WorkingDirectory=/home/asken-runner/actions-runner
ExecStart=/home/asken-runner/actions-runner/run.sh
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### ワークフロー例
`.github/workflows/ci.yml`
```yaml
name: asken CI
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
jobs:
  build:
    runs-on: [self-hosted, asken]
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Backend tests
        run: |
          cd backend
          go test ./...
      - name: Frontend lint
        run: |
          cd frontend
          npm ci
          npm run lint
```
`runs-on` ラベルは runner 登録時に付与したものに合わせてください。テストやデプロイ処理を追加しても良いです。

### GitHub 上での準備
- `Actions` を有効化
- 必要な Secrets (例: `GEMINI_API_KEY`) を `Settings > Secrets and variables > Actions` に登録
- ブランチ保護ルールで `Require status checks` を設定する場合は、このワークフロー名を指定

### exe.dev 側の運用
- `journalctl -u asken-runner -f` で runner ログを監視
- runner バイナリ更新時は `./svc.sh stop && ./svc.sh uninstall` の後に再インストール
- token 失効時は GitHub 側で runner を削除し、再登録

### ローカル開発モード

従来通り、手動で起動して開発することもできます。

#### 1. バックエンド起動

```bash
cd backend
export DATABASE_URL="postgres://asken:asken@localhost:5432/asken?sslmode=disable"
go run cmd/server/main.go
```

サーバーが `http://localhost:8080` で起動します。

#### 2. フロントエンド起動

```bash
cd frontend
npm run dev
```

フロントエンドが `http://localhost:3000` で起動します。

#### 3. ブラウザでアクセス

`http://localhost:3000` にアクセスして、画像をアップロードします。

## 使用方法

1. トップページにアクセス
2. 「ファイルを選択」ボタンをクリックして食事の画像を選択
3. プレビューが表示されたら「アップロードして分析」ボタンをクリック
4. 約2分待機（食材分類 + 栄養素計算）
5. 結果が表示されます

## テスト実行

### バックエンドテスト

```bash
cd backend
go test ./... -v
```

### テストカバレッジ確認

```bash
go test ./... -cover
```

## API仕様

### POST /api/analyze

食事画像を分析し、カロリーと栄養素を返却します。

**リクエスト**:

```
Content-Type: multipart/form-data

image: <画像ファイル（JPEG, PNG, HEIC、最大10MB）>
```

**レスポンス**:

```json
{
  "foods": [
    {
      "name": "刺身盛り合わせ",
      "estimated_amount": "8切れ",
      "calories_kcal": 360,
      "protein_g": 30.0,
      "fat_g": 24.6,
      "carbohydrates_g": 0.4,
      "source": "gemini"
    }
  ],
  "total_calories": 1289,
  "total_protein": 96.4,
  "total_fat": 80.4,
  "total_carbohydrates": 40.3
}
```

## セキュリティ

実装済みのセキュリティ対策：

- ✅ ファイルアップロード: 拡張子・MIMEタイプ・サイズチェック
- ✅ ディレクトリトラバーサル対策: `/tmp/asken/uploads/` に保存制限
- ✅ ファイル名サニタイズ: UUIDを使用
- ✅ コマンドインジェクション対策: 画像パスの絶対パス変換
- ✅ CORS設定: localhost:3000のみ許可

## 制限事項（MVP）

以下の機能は意図的に未実装です：

- ❌ 認証・認可（個人利用のみ）
- ❌ データベース（PostgreSQL）
- ❌ キャッシュ
- ❌ マルチユーザー対応

## トラブルシューティング

### Gemini CLIがタイムアウトする

- タイムアウト時間を延長: `backend/cmd/server/main.go` の `NewClassifier(120 * time.Second)` を調整
- Gemini CLIのワークスペース設定を確認

### CORS エラーが発生する

- バックエンドの `enableCORS` 関数でフロントエンドのURLが許可されているか確認

## 今後の拡張

- [ ] PostgreSQL食品マスタ連携
- [ ] データベース検索 → 見つからない場合のみGemini推定
- [ ] ユーザー認証
- [ ] 食事履歴保存
- [ ] 栄養バランス分析

## 開発者

Ryosuke Horie

## 関連ドキュメント

- [CLAUDE.md](./CLAUDE.md): プロジェクトガイドライン
- [IMPLEMENTATION.md](./IMPLEMENTATION.md): 実装サマリー
- [.claude/rules/](./claude/rules/): 詳細規約
