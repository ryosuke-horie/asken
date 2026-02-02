# Infrastructure

Terraformによるインフラ管理

## 前提条件

- [mise](https://mise.jdx.dev/) - ツールバージョン管理（Terraform自動インストール）
- [Google Cloud CLI](https://cloud.google.com/sdk/docs/install)
- [GitHub CLI](https://cli.github.com/) (オプション)

## 環境変数

このプロジェクトでは mise を使用してツールと環境変数を管理しています。
プロジェクトルートの `.mise.toml` に以下が定義されています：

- `GCP_PROJECT_ID` - GCPプロジェクトID
- `GCP_REGION` - GCPリージョン
- `GCLOUD_CONFIG_NAME` - gcloud構成名
- `TF_VAR_*` - Terraform変数

初回のみ信頼設定が必要です：

```bash
cd /path/to/utikomi
mise trust
mise install
```

## 初回セットアップ

### 1. GCPプロジェクト作成

GCPコンソールで新しいプロジェクトを作成:

- プロジェクトID: `utikomi-dev`
- 組織: 任意

### 2. 請求先アカウント紐付け

GCPコンソール > 請求 > プロジェクトの請求先アカウントを管理

### 3. サービスアカウント作成

CLIで作成（推奨）:

```bash
# サービスアカウント作成
gcloud iam service-accounts create terraform-admin \
  --display-name="Terraform Admin"

# 必要なロールを付与
gcloud projects add-iam-policy-binding utikomi-dev \
  --member="serviceAccount:terraform-admin@utikomi-dev.iam.gserviceaccount.com" \
  --role="roles/editor"

gcloud projects add-iam-policy-binding utikomi-dev \
  --member="serviceAccount:terraform-admin@utikomi-dev.iam.gserviceaccount.com" \
  --role="roles/firebase.admin"

# キーをJSON形式で取得
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=terraform-admin@utikomi-dev.iam.gserviceaccount.com
```

必要なロール:
- `roles/editor` - 基本的なリソース管理
- `roles/firebase.admin` - Firebase/Firestore管理

### 4. tfstate用バケット作成

GCPコンソール > Cloud Storage:

1. 「バケットを作成」をクリック
2. 名前: `utikomi-dev-tfstate`
3. リージョン: `asia-northeast1`
4. バージョニング: 有効

### 5. GitHub Personal Access Token作成

GitHub Settings > Developer settings > Personal access tokens > Tokens (classic):

1. 「Generate new token (classic)」をクリック
2. 名前: `terraform-utikomi`
3. スコープ:
   - `repo` (フルアクセス)
   - `admin:repo_hook` (webhook管理)
4. トークンをコピーして安全に保存

### 6. gcloud構成セットアップ

プロジェクト専用のgcloud構成を作成します（他のGCPプロジェクトとの混同を防ぐため）：

```bash
# 構成を作成
gcloud config configurations create utikomi-dev

# プロジェクトを設定
gcloud config set project utikomi-dev

# リージョン・ゾーンを設定
gcloud config set compute/region asia-northeast1
gcloud config set compute/zone asia-northeast1-a

# 認証
gcloud auth login
gcloud auth application-default login
```

構成の切り替え：

```bash
gcloud config configurations activate utikomi-dev
```

> **Note**: このプロジェクトでは mise により `CLOUDSDK_ACTIVE_CONFIG_NAME=utikomi-dev` が自動設定されます。
> プロジェクトディレクトリに入ると自動的に `utikomi-dev` 構成が使われます。

### 7. API有効化

```bash
cd infrastructure
./scripts/enable-apis.sh utikomi-dev
```

### 8. Terraform変数設定

```bash
cd environments/dev
cp terraform.tfvars.example terraform.tfvars
```

`terraform.tfvars`を編集:

```hcl
gcp_project_id    = "utikomi-dev"
github_repository = "ryosuke-horie/utikomi"
github_token      = "ghp_xxxx..."  # 手順5で取得したトークン
gcp_sa_key        = <<EOF
{
  "type": "service_account",
  ...
}
EOF
```

### 9. Terraformの実行

```bash
# 認証設定（サービスアカウントキーを使用する場合）
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa-key.json

# 初期化
terraform init

# プラン確認
terraform plan

# 適用
terraform apply
```

## ディレクトリ構造

```
infrastructure/
├── README.md                      # このファイル
├── scripts/
│   └── enable-apis.sh            # API有効化スクリプト
├── environments/
│   ├── dev/                      # dev環境
│   │   ├── main.tf
│   │   ├── providers.tf
│   │   ├── backend.tf
│   │   ├── variables.tf
│   │   ├── terraform.tfvars.example
│   │   └── outputs.tf
│   └── prod/                     # prod環境（将来用）
└── modules/
    ├── firestore/                # Firestoreデータベース
    ├── storage/                  # Cloud Storage
    ├── firebase-auth/            # Firebase Authentication
    └── github/                   # GitHub secrets/variables
```

## 環境ごとの設定

| 項目 | dev | prod |
|:---|:---|:---|
| プロジェクトID | utikomi-dev | utikomi-prod |
| Firestore削除保護 | 無効 | 有効 |
| Storageバージョニング | 無効 | 有効 |
| Storage自動削除 | 90日 | 無効 |
| CORS | 全許可 | 特定ドメイン |

## トラブルシューティング

### API有効化エラー

```
Error: Error enabling service: googleapi: Error 403: ...
```

請求先アカウントが紐付けられているか確認してください。

### 権限エラー

```
Error: Error creating Database: googleapi: Error 403: ...
```

サービスアカウントに必要なロールが付与されているか確認してください。

### tfstateバケットエラー

```
Error: Failed to get existing workspaces: querying Cloud Storage failed: ...
```

tfstateバケットが存在するか、アクセス権があるか確認してください。
