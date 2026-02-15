# E2E Tools

## `run-backend-e2e-dev.sh`

開発環境向けのバックエンドE2Eテスト実行スクリプトです。
デプロイ処理は含まず、E2E実行のみを行います。

```bash
# デフォルト設定で実行（Terraform output / 環境変数から自動解決）
./tools/e2e/run-backend-e2e-dev.sh

# 明示的にURLを指定して実行
./tools/e2e/run-backend-e2e-dev.sh --base-url https://example.a.run.app
```

### 設定解決順

- `E2E_BASE_URL`:
  1. CLIオプション `--base-url`
  2. 環境変数 `E2E_BASE_URL`
  3. Terraform output `cloud_run_url`
  4. `gcloud run services describe`（`GCP_PROJECT_ID` + `CLOUD_RUN_SERVICE_NAME` 使用）

- その他:
  - `GCP_PROJECT_ID`, `GCP_REGION`, `CLOUD_RUN_SERVICE_NAME`
  - `E2E_FIREBASE_API_KEY`, `SERVICE_ACCOUNT_EMAIL`, `E2E_TEST_UID`

### 前提

- `go` がインストール済み
- URL自動解決に `gcloud` を使う場合はログイン済み (`gcloud auth login`)
- 認証付きE2Eを通す場合は `E2E_FIREBASE_API_KEY` を設定
