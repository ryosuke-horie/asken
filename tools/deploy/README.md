# Deploy Tools

## `deploy-dev.sh`

開発環境（Cloud Run）へバックエンドをデプロイするスクリプトです。
E2E実行は含まず、デプロイ処理のみに責務を限定しています。

```bash
# デフォルト設定でデプロイ
./tools/deploy/deploy-dev.sh
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

### 前提

- `gcloud` でログイン済み (`gcloud auth login`)
- Docker が起動している
- Artifact Registry / Cloud Run へ必要な権限がある

E2E実行は `tools/e2e/run-backend-e2e-dev.sh` を使用してください。
