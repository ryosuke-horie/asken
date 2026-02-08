# プラン: Cloud Run デプロイ後の API E2E テスト設計

## Linear Issue
- Issue: EDG-573
- URL: https://linear.app/ryosuke-horie/issue/EDG-573

## 概要

Cloud Run デプロイ後に HTTP で API の E2E テストを自動実行する仕組みを設計・実装する。
iOSでのE2Eテストはコストが高いため、APIレベルでの自動テストにより効率的にバグを検出する。

## 設計決定

### 1. 実装方法: Go による E2E テスト

理由:
- 既存のテストパターン（testify）を活用可能
- CIに追加ツールのインストール不要
- 型安全性があり保守しやすい

ディレクトリ構成:
```
backend/
└── e2e/
    ├── e2e_test.go         # テストスイート、セットアップ
    ├── health_test.go      # ヘルスチェック
    ├── analyze_test.go     # 分析API
    ├── helpers.go          # HTTPクライアントラッパー
    ├── auth.go             # 認証トークン生成
    └── testdata/
        └── sample.jpg      # テスト用画像
```

### 2. 認証方式: Firebase カスタムトークン生成

理由:
- 本番と同じ認証フローをテスト
- テスト用ユーザーを明確に分離
- 追加のバイパス経路を作らないためセキュリティリスクが低い

実装:
- Firebase Admin SDK でカスタムトークンを生成
- REST API で ID トークンに交換
- テスト用ユーザーID: `e2e-test-user`

### 3. テスト範囲（優先度順）

| 優先度 | エンドポイント | テスト内容 |
|--------|----------------|------------|
| P0 | GET /api/health | ステータスコード200、JSON形式 |
| P0 | POST /api/analyze | 認証、202 Accepted、ID返却 |
| P0 | GET /api/analyze/:id | ステータス取得 |
| P1 | GET /api/history | 一覧取得 |
| P1 | GET /api/meals/daily | 日別データ取得 |

初回実装はP0のみ。P1以降は段階的に追加。

### 4. GitHub Actions 統合

deploy.yml に e2e-test ジョブを追加:

```yaml
e2e-test:
  needs: deploy
  runs-on: ubuntu-latest
  environment: dev
  steps:
    - uses: actions/checkout@v6
    - name: Setup Go
      uses: actions/setup-go@v6
      with:
        go-version: '1.25'
    - name: Authenticate to Google Cloud
      uses: google-github-actions/auth@v3
      with:
        workload_identity_provider: ${{ vars.WORKLOAD_IDENTITY_PROVIDER }}
        service_account: ${{ vars.SERVICE_ACCOUNT_EMAIL }}
    - name: Get Cloud Run URL
      id: get-url
      run: |
        URL=$(gcloud run services describe "${{ vars.CLOUD_RUN_SERVICE_NAME }}" \
          --region "${{ vars.GCP_REGION }}" --format 'value(status.url)')
        echo "url=${URL}" >> "$GITHUB_OUTPUT"
    - name: Run E2E tests
      env:
        E2E_BASE_URL: ${{ steps.get-url.outputs.url }}
        E2E_TEST_UID: "e2e-test-user"
      working-directory: backend
      run: go test -v -tags=e2e ./e2e/...
```

### 5. 失敗時の対応

- GitHub Step Summary に結果を出力
- ロールバックは当面実装しない（個人プロジェクトのため手動対応で十分）
- 必要に応じてSlack通知を追加可能

### 6. テストデータのクリーンアップ

- テスト用ユーザーID（`e2e-test-user`）を固定
- テスト終了時にそのユーザーのデータを削除
- 冪等性を確保

## 実装計画

### Phase 1: 基盤構築
1. `backend/e2e/` ディレクトリ作成
2. テストヘルパー関数の実装（HTTPクライアント）
3. ヘルスチェックテスト実装（認証不要）

### Phase 2: 認証付きテスト
1. Firebase トークン生成機能の実装
2. POST /api/analyze テスト実装
3. GET /api/analyze/:id テスト実装

### Phase 3: CI 統合
1. deploy.yml に e2e-test ジョブ追加
2. テスト結果のサマリ出力

### Phase 4: 追加テスト（オプション）
1. 履歴 API テスト
2. 日別食事 API テスト

## 技術的な考慮事項

### 参照すべき既存コード
- `backend/internal/handler/health_handler_test.go` - テストパターン
- `backend/internal/middleware/auth.go` - 認証処理
- `backend/internal/service/firebase_auth_service.go` - Firebase認証
- `.github/workflows/deploy.yml` - CI統合先

### 依存関係
- Firebase Admin SDK（既存）
- testify（既存）
- Workload Identity Federation（既存）

## 検証方法

1. ローカルでE2Eテストを実行
   ```bash
   E2E_BASE_URL=http://localhost:8080 go test -v -tags=e2e ./e2e/...
   ```

2. GitHub Actions でデプロイ後に自動実行されることを確認

3. テスト失敗時にStep Summaryに結果が出力されることを確認
