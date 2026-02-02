# 運用手順書（RUNBOOK）

最終更新: 2026-02-02

GCPサーバレス環境の運用手順、監視、トラブルシューティングを説明します。

## アーキテクチャ

| サービス | GCPサービス | 説明 |
|:---|:---|:---|
| データベース | Firestore | NoSQLデータベース |
| ストレージ | Cloud Storage | 画像保存 |
| 認証 | Firebase Auth | ユーザー認証 |
| AI | Gemini API | 画像認識・栄養素分析 |

インフラはTerraformで管理: [infrastructure/README.md](../infrastructure/README.md)

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

## 関連ドキュメント

- [README.md](../README.md) - プロジェクト概要
- [CONTRIB.md](./CONTRIB.md) - 開発者ガイド
- [infrastructure/README.md](../infrastructure/README.md) - Terraformインフラ管理
