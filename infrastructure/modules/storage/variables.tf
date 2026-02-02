# =============================================================================
# Cloud Storage Module - Variables
# =============================================================================

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "location" {
  description = "バケットのロケーション"
  type        = string
  default     = "ASIA-NORTHEAST1"
}

variable "enable_versioning" {
  description = "バージョニングの有効化"
  type        = bool
  default     = false
}

variable "lifecycle_delete_days" {
  description = "自動削除までの日数（0で無効）"
  type        = number
  default     = 0
}

variable "cors_origins" {
  description = "CORS許可オリジン"
  type        = list(string)
  default     = ["*"]
}

variable "force_destroy" {
  description = "削除時にバケット内のオブジェクトも削除"
  type        = bool
  default     = false
}

variable "service_account_email" {
  description = "ストレージアクセス用サービスアカウントのメールアドレス"
  type        = string
  default     = ""
}
