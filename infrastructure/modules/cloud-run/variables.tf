# =============================================================================
# Cloud Run Variables
# =============================================================================

# -----------------------------------------------------------------------------
# 基本設定
# -----------------------------------------------------------------------------

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "location" {
  description = "Cloud Runのリージョン"
  type        = string
  default     = "asia-northeast1"
}

variable "service_name" {
  description = "Cloud Runサービス名"
  type        = string
  default     = "uchikomi-api"
}

# -----------------------------------------------------------------------------
# コンテナ設定
# -----------------------------------------------------------------------------

variable "container_image" {
  description = "コンテナイメージのURL"
  type        = string
}

variable "container_port" {
  description = "コンテナがリッスンするポート"
  type        = number
  default     = 8080
}

# -----------------------------------------------------------------------------
# リソース設定
# -----------------------------------------------------------------------------

variable "cpu_limit" {
  description = "CPU制限"
  type        = string
  default     = "1"
}

variable "memory_limit" {
  description = "メモリ制限"
  type        = string
  default     = "512Mi"
}

variable "min_instances" {
  description = "最小インスタンス数（0でスケールダウン）"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "最大インスタンス数"
  type        = number
  default     = 10
}

variable "max_concurrent_requests" {
  description = "インスタンスあたりの最大同時リクエスト数"
  type        = number
  default     = 80
}

variable "request_timeout_seconds" {
  description = "リクエストタイムアウト（秒）"
  type        = number
  default     = 60
}

# -----------------------------------------------------------------------------
# ヘルスチェック
# -----------------------------------------------------------------------------

variable "health_check_path" {
  description = "ヘルスチェックのパス"
  type        = string
  default     = "/api/health"
}

# -----------------------------------------------------------------------------
# 環境変数・シークレット
# -----------------------------------------------------------------------------

variable "env_vars" {
  description = "環境変数（key-value）"
  type        = map(string)
  default     = {}
}

variable "secrets" {
  description = "Secret Managerからの環境変数"
  type = map(object({
    secret_name = string
    version     = string
  }))
  default = {}
}

# -----------------------------------------------------------------------------
# ネットワーク
# -----------------------------------------------------------------------------

variable "allow_unauthenticated" {
  description = "未認証リクエストを許可（APIは独自認証を使用）"
  type        = bool
  default     = true
}

variable "vpc_connector" {
  description = "VPCコネクタ名（空で無効）"
  type        = string
  default     = ""
}

# -----------------------------------------------------------------------------
# ラベル
# -----------------------------------------------------------------------------

variable "labels" {
  description = "リソースラベル"
  type        = map(string)
  default     = {}
}
