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
│   │   ├── service/           # ビジネスロジック
│   │   └── repository/        # データアクセス
│   └── pkg/gemini/            # Gemini API クライアント
├── ios/                        # iOSアプリ
│   ├── Uchikomi/              # メインアプリ（SwiftUI）
│   ├── UchikomiTests/         # ユニットテスト
│   └── UchikomiUITests/       # UIテスト
└── docs/                       # ドキュメント
    ├── CODEMAPS/              # コードマップ
    └── adr/                   # アーキテクチャ決定記録
```

## セットアップ

### 前提条件

| ツール | バージョン | 用途 |
|:---|:---|:---|
| Go | 1.25以上 | バックエンド開発 |
| Xcode | 16以上 | iOS開発 |
| Task | 3.x | タスクランナー |

### バックエンドのセットアップ

```bash
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

- 自動デプロイ: `backend/**` への変更がmainブランチにマージされると自動デプロイ
- 手動デプロイ: GitHub Actions の「Deploy」ワークフローから手動実行可能

詳細は [docs/RUNBOOK.md](./docs/RUNBOOK.md) を参照してください。

## 使用方法

1. トップページにアクセス
2. 「ファイルを選択」ボタンをクリックして食事の画像を選択
3. プレビューが表示されたら「アップロードして分析」ボタンをクリック
4. 約2分待機（食材分類 + 栄養素計算）
5. 結果が表示されます

## テスト実行

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
- CORS設定: 許可オリジンを制限

## トラブルシューティング

詳細は [docs/RUNBOOK.md](./docs/RUNBOOK.md) を参照してください。

## 今後の拡張

- [ ] 食事履歴保存
- [ ] 栄養バランス分析
- [ ] 体重推移グラフ

## 開発者

Ryosuke Horie

## 関連ドキュメント

- [CLAUDE.md](./CLAUDE.md): プロジェクトガイドライン
- [docs/CONTRIB.md](./docs/CONTRIB.md): 開発者ガイド
- [.claude/rules/](./.claude/rules/): 詳細規約
