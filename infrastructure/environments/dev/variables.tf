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
# Firebase Auth設定（Google Sign-In）
# -----------------------------------------------------------------------------

variable "google_oauth_client_id" {
  description = "Google OAuth Client ID（手動作成が必要）"
  type        = string
  default     = ""
}

variable "google_oauth_client_secret" {
  description = "Google OAuth Client Secret（手動作成が必要）"
  type        = string
  default     = ""
  sensitive   = true
}

# -----------------------------------------------------------------------------
# Cloud Run設定
# -----------------------------------------------------------------------------

variable "cloud_run_initial_image" {
  description = "Cloud Run初期デプロイ用のDockerイメージ（GitHub Actionsで上書きされる）"
  type        = string
  default     = "gcr.io/cloudrun/hello"
}

variable "cloud_run_allowed_origins" {
  description = "CORSで許可するオリジン"
  type        = list(string)
  default     = ["http://localhost:3000", "http://localhost:3001"]
}

# -----------------------------------------------------------------------------
# Workload Identity Federation設定
# -----------------------------------------------------------------------------

variable "github_owner" {
  description = "GitHubリポジトリオーナー"
  type        = string
  default     = "ryosuke-horie"
}

variable "github_repo" {
  description = "GitHubリポジトリ名"
  type        = string
  default     = "utikomi"
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

# -----------------------------------------------------------------------------
# Gemini API設定
# -----------------------------------------------------------------------------

variable "gemini_api_key" {
  description = "Gemini API Key（Google AI Studio発行）"
  type        = string
  sensitive   = true
}

# -----------------------------------------------------------------------------
# E2Eテスト設定
# -----------------------------------------------------------------------------

variable "e2e_firebase_api_key" {
  description = "E2Eテスト用のFirebase Web API Key"
  type        = string
  sensitive   = true
}
