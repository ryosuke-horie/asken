# =============================================================================
# Variables
# =============================================================================

# -----------------------------------------------------------------------------
# GCP共通設定
# -----------------------------------------------------------------------------

variable "gcp_project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "gcp_region" {
  description = "GCPリージョン"
  type        = string
  default     = "asia-northeast1"
}

# -----------------------------------------------------------------------------
# Firestore設定
# -----------------------------------------------------------------------------

variable "firestore_location" {
  description = "Firestoreのロケーション"
  type        = string
  default     = "asia-northeast1"
}

variable "firestore_delete_protection" {
  description = "Firestore削除保護の有効化"
  type        = bool
  default     = false
}

# -----------------------------------------------------------------------------
# Cloud Storage設定
# -----------------------------------------------------------------------------

variable "storage_location" {
  description = "Cloud Storageのロケーション"
  type        = string
  default     = "ASIA-NORTHEAST1"
}

variable "storage_versioning" {
  description = "Cloud Storageバージョニングの有効化"
  type        = bool
  default     = false
}

variable "storage_lifecycle_delete_days" {
  description = "Cloud Storageの自動削除日数（0で無効）"
  type        = number
  default     = 90
}

variable "storage_cors_origins" {
  description = "Cloud StorageのCORS許可オリジン"
  type        = list(string)
  default     = ["*"]
}

variable "storage_force_destroy" {
  description = "削除時にバケット内のオブジェクトも削除"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# GitHub設定
# -----------------------------------------------------------------------------

variable "github_repository" {
  description = "GitHubリポジトリ（owner/repo形式）"
  type        = string
}

variable "github_token" {
  description = "GitHub Personal Access Token"
  type        = string
  sensitive   = true
}

variable "gcp_sa_key" {
  description = "GCPサービスアカウントキー（JSON）"
  type        = string
  sensitive   = true
}
