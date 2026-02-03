# =============================================================================
# Artifact Registry Variables
# =============================================================================

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "location" {
  description = "Artifact Registryのリージョン"
  type        = string
  default     = "asia-northeast1"
}

variable "repository_id" {
  description = "リポジトリID"
  type        = string
  default     = "uchikomi"
}

variable "description" {
  description = "リポジトリの説明"
  type        = string
  default     = "ウチコミ バックエンドDockerイメージ"
}

variable "cleanup_policy_keep_count" {
  description = "保持するイメージの最大数（0で無効）"
  type        = number
  default     = 10
}

variable "github_actions_sa_email" {
  description = "GitHub ActionsサービスアカウントのEmail（空で無効）"
  type        = string
  default     = ""
}

variable "labels" {
  description = "リソースラベル"
  type        = map(string)
  default     = {}
}
