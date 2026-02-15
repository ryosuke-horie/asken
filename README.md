# ウチコミ - 格闘技向け体重管理アプリ

柔術/キックボクシングなど格闘技の体重コントロールを支援するAIエージェント付きアプリケーション。

AIエージェント（未実装）による減量サポートが差別化ポイント。

## 技術スタック

| レイヤー | 技術 |
|:---|:---|
| iOSアプリ | Swift / SwiftUI |
| バックエンド | Go 1.25 |
| AI | Gemini API（将来的にLangChain等でAIエージェント自作予定） |

## ディレクトリ構成

```
utikomi/
├── backend/                    # Goバックエンド
│   ├── cmd/server/            # HTTPサーバー
│   ├── internal/              # 内部パッケージ
│   │   ├── handler/           # HTTPハンドラー
│   │   ├── middleware/        # ミドルウェア（認証等）
│   │   ├── service/           # ビジネスロジック
│   │   ├── repository/        # データアクセス
│   │   ├── worker/            # バックグラウンドワーカー
│   │   ├── testutil/          # テストユーティリティ
│   │   └── util/              # ユーティリティ
│   └── pkg/                   # 共有パッケージ
│       ├── gemini/            # Gemini API クライアント
│       ├── database/          # データベース接続
│       └── storage/           # Cloud Storage クライアント
├── ios/                        # iOSアプリ
│   ├── Uchikomi/              # メインアプリ（SwiftUI）
│   │   └── Features/          # 機能モジュール
│   │       ├── Auth/          # 認証
│   │       ├── Meals/         # 食事記録
│   │       ├── Weight/        # 体重記録
│   │       ├── MyMenu/        # マイメニュー
│   │       └── Settings/      # 設定
│   ├── UchikomiCore/          # コアフレームワーク
│   └── UchikomiTests/         # ユニットテスト
├── infrastructure/             # Terraform（GCPインフラ管理）
└── docs/                       # ドキュメント
    ├── CODEMAPS/              # コードマップ
    ├── adr/                   # アーキテクチャ決定記録
    └── specs/                 # 仕様書
```

## セットアップ

### 前提条件

| ツール | バージョン | 用途 |
|:---|:---|:---|
| mise | 最新 | ツール管理（Go/Terraform/Lefthook） |
| Go | 1.25以上 | バックエンド開発 |
| Xcode | 16以上 | iOS開発 |
| Task | 3.x | タスクランナー |
| Lefthook | 2.x | Gitフック（lint/test/format） |

### バックエンドのセットアップ

```bash
mise trust
mise install
task setup
```

### iOSアプリのセットアップ

Xcodeで`ios/Uchikomi.xcodeproj`を開く。

## 起動方法

### ローカル開発

```bash
# バックエンドサーバーを起動
task run
```

サーバーが `http://localhost:8080` で起動します。

### 開発環境（Cloud Run）

| 環境 | エンドポイント |
|:---|:---|
| dev | https://uchikomi-api-dev-ah4e2vgm6q-an.a.run.app |

ヘルスチェック:

```bash
curl https://uchikomi-api-dev-ah4e2vgm6q-an.a.run.app/api/health
# {"status":"ok"}
```

### デプロイ

- 手動デプロイ（推奨）: `task deploy:dev`
- スクリプト直接実行: `./tools/deploy/deploy-dev.sh`

### E2E

- 開発環境E2E実行: `task e2e:dev`
- スクリプト直接実行: `./tools/e2e/run-backend-e2e-dev.sh`

詳細は [docs/RUNBOOK.md](./docs/RUNBOOK.md) を参照してください。

## 使用方法

### iOSアプリ

1. バックエンドサーバーを起動（`task run`）
2. XcodeでiOSアプリをビルド・実行
3. ログイン（開発環境ではモック認証が使用可能）
4. 食事記録画面でカメラまたはフォトライブラリから画像を選択
5. Gemini APIによる自動栄養素分析が実行される
6. 結果が表示され、食事として保存可能

## テスト実行

通常の開発では、Lefthookにより `pre-commit` / `pre-push` で lint・format・backend test が実行されます（iOSテストは一時停止中）。
初回のみ以下を実行してください:

```bash
task hooks:install
```

### バックエンドテスト

```bash
task test
```

### iOSテスト

```bash
task ios:test
```

### テストカバレッジ確認

```bash
cd backend && go test ./... -cover
```

## API仕様

### エンドポイント一覧

| メソッド | パス | 説明 |
|:---|:---|:---|
| GET | /api/health | ヘルスチェック |
| POST | /api/analyze | 食事画像分析 |
| GET | /api/analyze/{id} | 分析ステータス確認 |
| POST | /api/upload-image | 画像アップロード |
| GET | /api/images/{uuid} | 画像取得 |
| GET | /api/history | 分析履歴一覧 |
| GET | /api/history/{id} | 分析履歴詳細 |
| PUT | /api/history/{id} | 分析履歴更新 |
| DELETE | /api/history/{id} | 分析履歴削除 |
| GET | /api/meals/daily | 日別食事取得 |
| POST | /api/meals/skip | 食事スキップ |
| GET | /api/weight/records | 体重記録一覧 |
| POST | /api/weight/records | 体重記録作成 |
| GET | /api/weight/records/{id} | 体重記録詳細 |
| PUT | /api/weight/records/{id} | 体重記録更新 |
| DELETE | /api/weight/records/{id} | 体重記録削除 |
| GET | /api/weight/goal | 体重目標取得 |
| PUT | /api/weight/goal | 体重目標更新 |
| GET | /api/my-menu | マイメニュー一覧 |
| POST | /api/my-menu | マイメニュー作成 |
| GET | /api/my-menu/{id} | マイメニュー詳細 |
| PUT | /api/my-menu/{id} | マイメニュー更新 |
| DELETE | /api/my-menu/{id} | マイメニュー削除 |

### POST /api/analyze

食事画像を分析し、カロリーと栄養素を返却します。

リクエスト:

```
Content-Type: multipart/form-data

image: <画像ファイル（JPEG, PNG, HEIC、最大10MB）>
```

レスポンス:

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

- ファイルアップロード: 拡張子・MIMEタイプ・サイズチェック
- ディレクトリトラバーサル対策: `/tmp/uchikomi/uploads/` に保存制限
- ファイル名サニタイズ: UUIDを使用
- コマンドインジェクション対策: 画像パスの絶対パス変換
- CORS設定: 許可オリジンを制限（localhostは開発環境のみ）
- dev認証バイパス: ビルドタグ(`production`)で本番ビルドから除外
- 入力バリデーション: meal_dateのYYYY-MM-DD形式検証

## トラブルシューティング

詳細は [docs/RUNBOOK.md](./docs/RUNBOOK.md) を参照してください。

## 今後の拡張

- [x] 食事履歴保存
- [ ] 栄養バランス分析
- [x] 体重記録・推移グラフ
- [x] マイメニュー機能
- [x] 食事・体重リマインダー通知
- [ ] AIエージェントによる減量サポート

## 開発者

Ryosuke Horie

## 関連ドキュメント

- [CLAUDE.md](./CLAUDE.md): プロジェクトガイドライン
- [docs/CONTRIB.md](./docs/CONTRIB.md): 開発者ガイド
- [.claude/rules/](./.claude/rules/): 詳細規約
