# =============================================================================
# dev環境メインエントリポイント
# =============================================================================

locals {
  environment = "dev"
}

# -----------------------------------------------------------------------------
# Firestore
# -----------------------------------------------------------------------------

module "firestore" {
  source = "../../modules/firestore"

  project_id        = var.gcp_project_id
  location          = var.firestore_location
  delete_protection = var.firestore_delete_protection
}

# -----------------------------------------------------------------------------
# Cloud Storage
# -----------------------------------------------------------------------------

module "storage" {
  source = "../../modules/storage"

  project_id            = var.gcp_project_id
  location              = var.storage_location
  enable_versioning     = var.storage_versioning
  lifecycle_delete_days = var.storage_lifecycle_delete_days
  cors_origins          = var.storage_cors_origins
  force_destroy         = var.storage_force_destroy
}

# -----------------------------------------------------------------------------
# Firebase Authentication
# -----------------------------------------------------------------------------

module "firebase_auth" {
  source = "../../modules/firebase-auth"

  project_id                 = var.gcp_project_id
  authorized_domains         = ["${var.gcp_project_id}.firebaseapp.com", "localhost"]
  google_oauth_client_id     = var.google_oauth_client_id
  google_oauth_client_secret = var.google_oauth_client_secret
}

# -----------------------------------------------------------------------------
# Artifact Registry
# -----------------------------------------------------------------------------

module "artifact_registry" {
  source = "../../modules/artifact-registry"

  project_id                = var.gcp_project_id
  location                  = var.gcp_region
  repository_id             = "uchikomi"
  description               = "ウチコミ バックエンドDockerイメージ（${local.environment}）"
  cleanup_policy_keep_count = 10
  cloud_run_sa_email        = module.cloud_run.service_account_email

  labels = {
    environment = local.environment
    managed_by  = "terraform"
  }
}

# -----------------------------------------------------------------------------
# Cloud Run
# -----------------------------------------------------------------------------

module "cloud_run" {
  source = "../../modules/cloud-run"

  project_id   = var.gcp_project_id
  location     = var.gcp_region
  service_name = "uchikomi-api-${local.environment}"

  # 初期デプロイ時のダミーイメージ（GitHub Actionsで上書きされる）
  container_image = var.cloud_run_initial_image

  # リソース設定（dev環境は最小構成）
  cpu_limit     = "1"
  memory_limit  = "512Mi"
  min_instances = 0
  max_instances = 2

  # 環境変数
  env_vars = {
    GCP_PROJECT_ID = var.gcp_project_id
    ALLOWED_ORIGINS = join(",", var.cloud_run_allowed_origins)
  }

  # 未認証リクエストを許可（APIは独自のFirebase Auth検証を使用）
  allow_unauthenticated = true

  labels = {
    environment = local.environment
    managed_by  = "terraform"
  }
}

# -----------------------------------------------------------------------------
# GitHub Secrets/Variables
# -----------------------------------------------------------------------------

module "github" {
  source = "../../modules/github"

  github_repository      = var.github_repository
  environment            = local.environment
  gcp_project_id         = var.gcp_project_id
  gcp_region             = var.gcp_region
  gcp_sa_key             = var.gcp_sa_key
  firestore_database     = module.firestore.database_name
  storage_bucket         = module.storage.bucket_name
  artifact_registry_url  = module.artifact_registry.repository_url
  cloud_run_service_name = module.cloud_run.service_name
}
