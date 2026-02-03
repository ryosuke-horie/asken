# =============================================================================
# GitHub Module - Variables
# =============================================================================

variable "github_repository" {
  description = "GitHubリポジトリ（owner/repo形式）"
  type        = string
}

variable "environment" {
  description = "GitHub Environment名"
  type        = string
}

variable "gcp_project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "gcp_region" {
  description = "GCPリージョン"
  type        = string
}

variable "gcp_sa_key" {
  description = "GCPサービスアカウントキー（JSON）- WIF使用時は不要"
  type        = string
  sensitive   = true
  default     = ""
}

variable "firestore_database" {
  description = "Firestoreデータベース名"
  type        = string
}

variable "storage_bucket" {
  description = "Cloud Storageバケット名"
  type        = string
}

variable "artifact_registry_url" {
  description = "Artifact RegistryのベースURL"
  type        = string
  default     = ""
}

variable "cloud_run_service_name" {
  description = "Cloud Runサービス名"
  type        = string
  default     = ""
}

variable "workload_identity_provider" {
  description = "Workload Identity Provider フルパス"
  type        = string
  default     = ""
}

variable "service_account_email" {
  description = "GitHub ActionsがimpersonateするサービスアカウントのEmail"
  type        = string
  default     = ""
}
