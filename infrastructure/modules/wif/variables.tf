# =============================================================================
# Workload Identity Federation Variables
# =============================================================================

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "pool_id" {
  description = "Workload Identity Pool ID"
  type        = string
  default     = "github-pool"
}

variable "provider_id" {
  description = "Workload Identity Provider ID"
  type        = string
  default     = "github-provider"
}

variable "github_owner" {
  description = "GitHubリポジトリオーナー（ユーザー名または組織名）"
  type        = string
}

variable "github_repo" {
  description = "GitHubリポジトリ名"
  type        = string
}

variable "service_account_id" {
  description = "WIFを関連付けるサービスアカウントのフルパス（projects/PROJECT/serviceAccounts/SA@PROJECT.iam.gserviceaccount.com）"
  type        = string
}

variable "service_account_email" {
  description = "サービスアカウントのメールアドレス（Token Creator権限付与に使用）"
  type        = string
}
