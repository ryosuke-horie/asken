# バックエンドE2Eテストガイドライン

> E2Eテストはデプロイされた環境に対して実行される統合テストです。

## 基本方針

| 項目 | 方針 |
|:---|:---|
| 実行タイミング | Deploy ワークフローのみ（CIでは実行しない） |
| ビルドタグ | `-tags=e2e` が必要 |
| 対象環境 | Cloud Run (dev) |
| テスト対象 | デプロイされたAPIエンドポイント |

## 実行方法

### ローカル実行

```bash
# 環境変数を設定
export E2E_BASE_URL="https://uchikomi-api-dev-xxx.a.run.app"
export E2E_FIREBASE_API_KEY="your-api-key"
export E2E_TEST_UID="e2e-test-user"

# E2Eテストを実行
cd backend
go test -v -tags=e2e ./e2e/...
```

### CI/CDでの実行

Deploy ワークフロー (`.github/workflows/deploy.yml`) で自動実行されます：

```yaml
- name: Run E2E tests
  env:
    E2E_BASE_URL: ${{ steps.get-url.outputs.url }}
    E2E_FIREBASE_API_KEY: ${{ secrets.E2E_FIREBASE_API_KEY }}
    E2E_TEST_UID: "e2e-test-user"
  working-directory: backend
  run: |
    go test -v -tags=e2e ./e2e/...
```

## レート制限への対応

### レート制限設定

アプリケーションのレート制限ミドルウェア (`internal/middleware/rate_limit.go`) により、以下の制限があります：

| 設定項目 | 値 | 意味 |
|:---|:---|:---|
| `GeminiRateLimit` | 0.2 | 5秒に1回 |
| `GeminiBurstSize` | 2 | バーストで2リクエストまで許可 |

### Gemini APIを呼び出すエンドポイント

- `POST /api/analyze`
- `GET /api/history/{id}` (分析結果取得時)
- `PUT /api/history/{id}` (履歴更新時)

### E2Eテストでの対応

Gemini APIを呼び出すテストケースでは、テストの先頭で以下のヘルパー関数を呼び出してください：

```go
// waitForGeminiRateLimit はGemini APIのレート制限リセットを待つ
//
// レート制限設定: GeminiRateLimit=0.2 (5秒に1回), GeminiBurstSize=2
// バーストを使い切った後は5秒待つ必要がある
func waitForGeminiRateLimit() {
    time.Sleep(5 * time.Second)
}
```

#### 使用例

```go
func TestAnalyze_GetStatus_Success(t *testing.T) {
    // 前のテストからのレート制限リセットを待つ
    waitForGeminiRateLimit()

    client, ctx := authenticatedClient(t, 30*time.Second)
    // ... テストコード
}
```

### 注意点

- 複数のテストが連続してGemini APIを呼び出す場合、各テストの先頭で待機する
- 待機時間は5秒を推奨（レート制限のリセット期間）
- テストUIDは固定値 (`e2e-test-user`) を使用しているため、同じUIDでのリクエストが累積する

## 環境変数

| 変数名 | 必須 | 説明 |
|:---|:---:|:---|
| `E2E_BASE_URL` | Yes | デプロイされたAPIのベースURL |
| `E2E_FIREBASE_API_KEY` | Yes | Firebase APIキー |
| `E2E_TEST_UID` | Yes | テスト用ユーザーUID（認証済みとして扱われる） |

## テストの命名規則

日本語「〜すべき」表現を使用（他のテストと一貫性を保つため）：

```go
func TestAnalyze_TextInput_Success(t *testing.T) {
    // テスト実装
}

func TestAnalyze_GetStatus_Success(t *testing.T) {
    // テスト実装
}
```

## テスト構成

```
backend/
├── e2e/
│   ├── analyze_test.go      # 分析APIのE2Eテスト
│   └── client.go            # E2Eテスト用HTTPクライアント
└── cmd/server/
    └── main.go              # 本番コード
```

## CI での実行可否

| ワークフロー | E2Eテスト実行 |
|:---|:---:|
| CI (`ci.yml`) | No - ユニットテストのみ |
| Deploy (`deploy.yml`) | Yes - デプロイ後に実行 |

CIではE2Eテストを実行しないため、レート制限の問題は発生しません。

## 既知の問題

- 同時に複数のDeployが実行されると、同じ `e2e-test-user` UIDで競合が発生する可能性
- これは通常の運用では稀（mainブランチへのマージは直列に進むため）

## 関連ドキュメント

- `.claude/rules/testing-details.md` - 一般的なテスト詳細
- `.claude/rules/backend-golang.md` - バックエンド開発規約
- `.github/workflows/deploy.yml` - Deployワークフロー定義
