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
    # E2E_RUN_GEMINI: "true"  # Gemini APIテストを有効化する場合
  working-directory: backend
  run: |
    go test -v -tags=e2e ./e2e/...
```

デプロイ時のGeminiテスト実行方針: デフォルトはスキップ。必要に応じて `E2E_RUN_GEMINI: "true"` を追加することで有効化できる。

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
// バーストを使い切った後は10秒待つ（余裕を持たせるため5秒→10秒に増量）
func waitForGeminiRateLimit() {
    time.Sleep(10 * time.Second)
}
```

#### 使用例

```go
func TestAnalyze_GetStatus_Success(t *testing.T) {
    skipIfGeminiDisabled(t)
    // 前のテストからのレート制限リセットを待つ
    waitForGeminiRateLimit()

    client, ctx := authenticatedClient(t, 30*time.Second)
    // ... テストコード
}
```

### 注意点

- 複数のテストが連続してGemini APIを呼び出す場合、各テストの先頭で待機する
- 待機時間は10秒を推奨（GeminiRateLimit=0.2 の逆数5秒に余裕を持たせた値）
- テストUIDは固定値 (`e2e-test-user`) を使用しているため、同じUIDでのリクエストが累積する

## Gemini APIテストの制御

Gemini APIを呼び出すテストはデフォルトでスキップされる。APIコストとレート制限を回避するためのオプトイン方式。

### スキップ対象テスト

`skipIfGeminiDisabled(t)` が先頭に追加されており、`E2E_RUN_GEMINI=true` でなければスキップされる:

- `TestAnalyze_TextInput_Success`（`analyze_test.go`）
- `TestAnalyze_GetStatus_Success`（`analyze_test.go`）
- `TestHistory_List_Success`（`history_test.go`）
- `TestHistory_Detail_Success`（`history_test.go`）
- `TestHistory_Update_Success`（`history_test.go`）
- `TestHistory_Update_InvalidRequest_EmptyFoods`（`history_test.go`）
- `TestHistory_Update_InvalidRequest_NegativeCalories`（`history_test.go`）
- `TestHistory_Update_InvalidRequest_EmptyName`（`history_test.go`）
- `TestHistory_Delete_Success`（`history_test.go`）
- `TestExerciseRecords_Create_WithGemini_Success`（`exercise_test.go`）

### Geminiテストを有効化する方法

```bash
# 環境変数で有効化
export E2E_RUN_GEMINI=true
go test -v -tags=e2e ./e2e/...

# または実行スクリプトでオプション指定
tools/e2e/run-backend-e2e-dev.sh --run-gemini
```

### Geminiテストを新規追加する場合

Gemini APIを呼び出すテストを追加する際は、テストの先頭に以下を追加すること:

```go
func TestSomething_WithGemini(t *testing.T) {
    skipIfGeminiDisabled(t)
    waitForGeminiRateLimit()
    // ... テストコード
}
```

## 環境変数

| 変数名 | 必須 | 説明 |
|:---|:---:|:---|
| `E2E_BASE_URL` | Yes | デプロイされたAPIのベースURL |
| `E2E_FIREBASE_API_KEY` | Yes | Firebase APIキー |
| `E2E_TEST_UID` | Yes | テスト用ユーザーUID（認証済みとして扱われる） |
| `E2E_RUN_GEMINI` | No | `true` に設定するとGemini APIテストを有効化（デフォルト: スキップ） |

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
│   ├── meals_test.go        # 食事APIのE2Eテスト
│   ├── history_test.go      # 分析履歴APIのE2Eテスト
│   ├── image_test.go        # 画像APIのE2Eテスト
│   ├── weight_test.go       # 体重管理APIのE2Eテスト
│   ├── exercise_test.go     # 運動記録APIのE2Eテスト
│   ├── health_test.go       # ヘルスチェックのE2Eテスト
│   ├── helpers.go           # E2Eテスト用ヘルパー関数
│   ├── auth.go              # Firebase認証ヘルパー
│   ├── cleanup.go           # テストデータクリーンアップ
│   └── e2e_test.go          # TestMain・共通初期化
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
