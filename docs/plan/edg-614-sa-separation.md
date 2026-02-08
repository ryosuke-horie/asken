# プラン: Cloud RunのSA権限分離（ランタイム/CI/CD）

## Linear Issue
- Issue: EDG-614
- URL: https://linear.app/ryosuke-horie/issue/EDG-614

## 概要

セキュリティ監査（C-CICD-1）により指摘されたCloud RunのSA権限分離を実施。
ランタイムSAとデプロイ（CI/CD）SAを分離し、最小権限の原則に基づいた構成に変更する。

## Terraform State移行手順

コード変更後、`terraform apply`の前に以下のstate移行を実行すること。

```bash
cd infrastructure/environments/dev

# 1. state バックアップ
terraform state pull > state-backup-$(date +%Y%m%d-%H%M%S).json

# 2. ランタイムSAおよび権限のリソース名移動（GCP側は変更なし）
terraform state mv \
  'module.cloud_run.google_service_account.cloud_run' \
  'module.cloud_run.google_service_account.runtime'

terraform state mv \
  'module.cloud_run.google_project_iam_member.firestore_user' \
  'module.cloud_run.google_project_iam_member.runtime_firestore_user'

terraform state mv \
  'module.cloud_run.google_project_iam_member.storage_object_user' \
  'module.cloud_run.google_project_iam_member.runtime_storage_object_user'

terraform state mv \
  'module.cloud_run.google_project_iam_member.firebase_auth' \
  'module.cloud_run.google_project_iam_member.runtime_firebase_auth'

terraform state mv \
  'module.cloud_run.google_project_iam_member.secret_accessor[0]' \
  'module.cloud_run.google_project_iam_member.runtime_secret_accessor[0]'

terraform state mv \
  'module.cloud_run.google_service_account_iam_member.self_token_creator' \
  'module.cloud_run.google_service_account_iam_member.runtime_self_token_creator'

# 3. plan で差分確認
#
# 以下のリソースは state mv 不要（member/targetが変更されるためdestroy+createで対応）:
#   - artifact_registry_writer: ランタイムSA→デプロイSAへのmember変更
#   - run_developer: ランタイムSA→デプロイSAへのmember変更
#   - service_account_user: 自身→デプロイSAからの付与に変更
#   - WIF token_creator: WIFモジュールから削除、cloud-runモジュールのdeploy_token_creator_on_runtimeで管理
#
# 期待される変更:
#   + create: deploy SA（google_service_account.deploy）
#   + create: deploy用IAM bindings 4リソース（artifact_registry_writer, run_developer, acts_as_runtime, token_creator_on_runtime）
#   + create: DEPLOY_SERVICE_ACCOUNT_EMAIL GitHub Actions環境変数
#   - destroy: 旧CI/CD用IAM bindings（artifact_registry_writer, run_developer, service_account_user）
#   - destroy: WIF token_creator（cloud-runモジュールのdeploy_token_creator_on_runtimeに移管済み）
#   ~ update: WIF workload_identity_user（service_account_idがデプロイSAに変更）
#   ~ no change: ランタイムSA、ランタイム権限、Cloud Runサービス（state mvで対応済み）
terraform plan

# 4. 問題なければ apply
terraform apply
```

## ロールバック手順

```bash
# 1. バックアップからstate復元
terraform state push state-backup-YYYYMMDD-HHMMSS.json

# 2. コード変更をrevert
git checkout -- infrastructure/ .github/workflows/deploy.yml

# 3. 復旧確認
terraform plan
terraform apply
```

## 変更ファイル一覧

| ファイル | 変更内容 |
|:---|:---|
| `infrastructure/modules/cloud-run/main.tf` | SA分離、権限再配置 |
| `infrastructure/modules/cloud-run/outputs.tf` | deploy SA outputs追加 |
| `infrastructure/modules/wif/main.tf` | token_creator削除 |
| `infrastructure/modules/wif/variables.tf` | service_account_email削除 |
| `infrastructure/modules/github/main.tf` | DEPLOY_SERVICE_ACCOUNT_EMAIL追加 |
| `infrastructure/modules/github/variables.tf` | deploy_service_account_email追加 |
| `infrastructure/environments/dev/main.tf` | モジュール接続更新 |
| `.github/workflows/deploy.yml` | WIF認証SAをデプロイSAに変更 |
