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

### 1. バックエンド起動

```bash
cd backend
go run cmd/server/main.go
```

サーバーが `http://localhost:8080` で起動します。

### 2. フロントエンド起動

```bash
cd frontend
npm run dev
```

フロントエンドが `http://localhost:3000` で起動します。

### 3. ブラウザでアクセス

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
