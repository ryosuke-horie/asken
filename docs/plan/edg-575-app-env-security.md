# プラン: APP_ENV=production の明示的設定

## Linear Issue
- Issue: EDG-575
- URL: https://linear.app/ryosuke-horie/issue/EDG-575

## 概要

`APP_ENV=development`での認証バイパスに関するセキュリティ確認を行い、本番環境の設定を明示的にすることでリスクを低減する。

## 現状の調査結果

| 項目 | 状況 | 評価 |
|:---|:---|:---|
| `IsDevMode()` | `APP_ENV=development`のみで`true`を返す | 安全 |
| Dockerfile | `APP_ENV`ハードコードなし | 安全 |
| Terraform Cloud Run | `APP_ENV`設定なし（未設定=本番認証） | 安全 |
| deploy.yml | `APP_ENV`設定なし | 安全 |

現状でもセキュリティリスクはないが、意図を明確にするため改善を行う。

## 実装計画

### 1. Terraformで`APP_ENV=production`を明示的に設定

ファイル: `infrastructure/environments/dev/main.tf`

```hcl
env_vars = {
  GCP_PROJECT_ID  = var.gcp_project_id
  ALLOWED_ORIGINS = join(",", var.cloud_run_allowed_origins)
  APP_ENV         = "production"  # 追加
}
```

### 2. 起動ログでAPP_ENVを出力（既存で十分）

`main.go:188-189`で既に警告ログを出力しているため、追加対応不要:
```go
if middleware.IsDevMode() {
    log.Println("WARNING: Running in development mode with mock authentication")
```

## 変更ファイル

- `infrastructure/environments/dev/main.tf` - 環境変数に`APP_ENV=production`を追加

## テスト計画

1. `terraform plan`で変更内容を確認
2. 変更が環境変数追加のみであることを確認

## 検証方法

1. Terraform planの実行:
   ```bash
   cd infrastructure/environments/dev
   terraform plan
   ```

2. 期待される変更:
   - Cloud Runサービスの環境変数に`APP_ENV=production`が追加される
