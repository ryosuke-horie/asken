# =============================================================================
# Firestore Module - Variables
# =============================================================================

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "location" {
  description = "Firestoreのロケーション"
  type        = string
  default     = "asia-northeast1"
}

variable "delete_protection" {
  description = "削除保護の有効化"
  type        = bool
  default     = false
}
