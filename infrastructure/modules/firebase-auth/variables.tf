# =============================================================================
# Firebase Authentication Module - Variables
# =============================================================================

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "authorized_domains" {
  description = "認可されたドメイン"
  type        = list(string)
  default     = []
}

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
