# Deploy Tools

## `deploy-dev.sh`

開発環境（Cloud Run）へバックエンドをデプロイするスクリプトです。

```bash
# デフォルト設定でデプロイ
./tools/deploy/deploy-dev.sh

# E2Eテストも実行
./tools/deploy/deploy-dev.sh --run-e2e
```

### 設定方法

以下の順で値を解決します。

1. CLIオプション
2. 環境変数
3. Terraform output（`infrastructure/environments/dev`）

使用する主な環境変数:

- `GCP_PROJECT_ID`
- `GCP_REGION`
- `ARTIFACT_REGISTRY_URL`
- `CLOUD_RUN_SERVICE_NAME`
- `E2E_FIREBASE_API_KEY`（`--run-e2e` 時）
- `SERVICE_ACCOUNT_EMAIL`（必要に応じて）

### 前提

- `gcloud` でログイン済み (`gcloud auth login`)
- Docker が起動している
- Artifact Registry / Cloud Run へ必要な権限がある
