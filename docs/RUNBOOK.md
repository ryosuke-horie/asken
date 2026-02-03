# 運用手順書（RUNBOOK）

最終更新: 2026-02-03

GCPサーバレス環境の運用手順、監視、トラブルシューティングを説明します。

## アーキテクチャ

| サービス | GCPサービス | 説明 |
|:---|:---|:---|
| バックエンドAPI | Cloud Run | サーバーレスコンテナ |
| コンテナレジストリ | Artifact Registry | Dockerイメージ保存 |
| データベース | Firestore | NoSQLデータベース |
| ストレージ | Cloud Storage | 画像保存 |
| 認証 | Firebase Auth | ユーザー認証 |
| AI | Gemini API | 画像認識・栄養素分析 |
| CI/CD認証 | Workload Identity Federation | キーレス認証 |

インフラはTerraformで管理: [infrastructure/README.md](../infrastructure/README.md)

## デプロイ

### 自動デプロイ（推奨）

mainブランチにpushすると、GitHub Actionsが自動的にCloud Runにデプロイします。

```
Push to main → Build Docker → Push to Artifact Registry → Deploy to Cloud Run
```

ワークフロー: `.github/workflows/deploy.yml`

### デプロイの確認

```bash
# Cloud Runサービスの状態を確認
gcloud run services describe utikomi-api --region asia-northeast1

# デプロイ済みリビジョンを確認
gcloud run revisions list --service utikomi-api --region asia-northeast1

# ログを確認
gcloud run services logs read utikomi-api --region asia-northeast1 --limit 50
```

### ロールバック

問題が発生した場合、以前のリビジョンにロールバックできます:

```bash
# 利用可能なリビジョンを確認
gcloud run revisions list --service utikomi-api --region asia-northeast1

# 特定のリビジョンにトラフィックを切り替え
gcloud run services update-traffic utikomi-api \
  --region asia-northeast1 \
  --to-revisions=utikomi-api-00001-abc=100
```

## 監視

### Cloud Runメトリクス

GCPコンソール > Cloud Run > utikomi-api > メトリクス で以下を確認:

- リクエスト数
- レイテンシ
- エラー率
- インスタンス数

### ログ確認

```bash
# リアルタイムログ
gcloud run services logs tail utikomi-api --region asia-northeast1

# エラーログのみ
gcloud logging read "resource.type=cloud_run_revision AND severity>=ERROR" \
  --project utikomi-dev --limit 20
```

### ヘルスチェック

Cloud Runは自動的にヘルスチェックを実行します:
- エンドポイント: `/api/health`
- 起動プローブ: 10秒後、10秒間隔、3回失敗で再起動
- 存活プローブ: 30秒間隔、3回失敗で再起動

## GCPインフラ管理（Terraform）

### インフラの変更

```bash
cd infrastructure/environments/dev

# 変更内容を確認
terraform plan

# 変更を適用
terraform apply
```

### リソースの状態確認

```bash
# Terraform状態を確認
terraform state list

# 特定リソースの詳細
terraform state show <resource_name>
```

## トラブルシューティング

### Firebase権限エラー

```
Error: Error creating Database: googleapi: Error 403
```

サービスアカウントに`roles/firebase.admin`が付与されているか確認:

```bash
gcloud projects get-iam-policy utikomi-dev \
  --flatten="bindings[].members" \
  --filter="bindings.members:terraform-admin@utikomi-dev.iam.gserviceaccount.com"
```

### Terraform認証エラー

`GOOGLE_APPLICATION_CREDENTIALS`が正しく設定されているか確認:

```bash
echo $GOOGLE_APPLICATION_CREDENTIALS
# miseを使用している場合は自動設定される
```

詳細は[infrastructure/README.md](../infrastructure/README.md#トラブルシューティング)を参照。

### Cloud Runデプロイエラー

#### イメージプッシュ失敗

```
denied: Permission denied for resource
```

Artifact Registryへの認証が切れている可能性があります:

```bash
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
```

#### GitHub Actions WIF認証エラー

```
Error: google-github-actions/auth failed with: unable to generate credentials
```

WIF設定を確認:

```bash
# Workload Identity Poolの確認
gcloud iam workload-identity-pools list --location=global

# Providerの確認
gcloud iam workload-identity-pools providers list \
  --workload-identity-pool=github-pool --location=global
```

#### Cloud Run起動失敗

```
Container failed to start
```

1. ログを確認: `gcloud run services logs read utikomi-api --region asia-northeast1`
2. ヘルスチェックパス(`/api/health`)が正しく応答するか確認
3. 環境変数が正しく設定されているか確認

## 関連ドキュメント

- [README.md](../README.md) - プロジェクト概要
- [CONTRIB.md](./CONTRIB.md) - 開発者ガイド
- [infrastructure/README.md](../infrastructure/README.md) - Terraformインフラ管理
