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
