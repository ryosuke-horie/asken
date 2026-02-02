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

  project_id         = var.gcp_project_id
  authorized_domains = ["${var.gcp_project_id}.firebaseapp.com", "localhost"]
}

# -----------------------------------------------------------------------------
# GitHub Secrets/Variables
# -----------------------------------------------------------------------------

module "github" {
  source = "../../modules/github"

  github_repository  = var.github_repository
  environment        = local.environment
  gcp_project_id     = var.gcp_project_id
  gcp_region         = var.gcp_region
  gcp_sa_key         = var.gcp_sa_key
  firestore_database = module.firestore.database_name
  storage_bucket     = module.storage.bucket_name
}
