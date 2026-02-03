# =============================================================================
# Outputs
# =============================================================================

# -----------------------------------------------------------------------------
# GCPプロジェクト情報
# -----------------------------------------------------------------------------

output "project_id" {
  description = "GCPプロジェクトID"
  value       = var.gcp_project_id
}

# -----------------------------------------------------------------------------
# Firestore情報
# -----------------------------------------------------------------------------

output "firestore_database" {
  description = "Firestoreデータベース名"
  value       = module.firestore.database_name
}

# -----------------------------------------------------------------------------
# Cloud Storage情報
# -----------------------------------------------------------------------------

output "storage_bucket" {
  description = "画像保存用バケット名"
  value       = module.storage.bucket_name
}

output "storage_bucket_url" {
  description = "画像保存用バケットURL"
  value       = module.storage.bucket_url
}

# -----------------------------------------------------------------------------
# Firebase Auth情報
# -----------------------------------------------------------------------------

output "firebase_auth_domain" {
  description = "Firebase認証ドメイン"
  value       = module.firebase_auth.auth_domain
}

# -----------------------------------------------------------------------------
# Artifact Registry情報
# -----------------------------------------------------------------------------

output "artifact_registry_url" {
  description = "Artifact RegistryのベースURL"
  value       = module.artifact_registry.repository_url
}

# -----------------------------------------------------------------------------
# Cloud Run情報
# -----------------------------------------------------------------------------

output "cloud_run_url" {
  description = "Cloud RunサービスのURL"
  value       = module.cloud_run.service_url
}

output "cloud_run_service_name" {
  description = "Cloud Runサービス名"
  value       = module.cloud_run.service_name
}

# -----------------------------------------------------------------------------
# GitHub Environment情報
# -----------------------------------------------------------------------------

output "github_environment" {
  description = "GitHub Environment名"
  value       = module.github.environment_name
}
