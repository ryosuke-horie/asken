# =============================================================================
# GitHub Module
# =============================================================================

# GitHub Repository Data
data "github_repository" "main" {
  full_name = var.github_repository
}

# =============================================================================
# Repository Variables（全環境共通）
# =============================================================================

resource "github_actions_variable" "gcp_project_id" {
  repository    = data.github_repository.main.name
  variable_name = "GCP_PROJECT_ID"
  value         = var.gcp_project_id
}

resource "github_actions_variable" "gcp_region" {
  repository    = data.github_repository.main.name
  variable_name = "GCP_REGION"
  value         = var.gcp_region
}

# =============================================================================
# Repository Secrets（全環境共通）
# =============================================================================

resource "github_actions_secret" "gcp_sa_key" {
  repository      = data.github_repository.main.name
  secret_name     = "GCP_SA_KEY"
  plaintext_value = var.gcp_sa_key
}

# =============================================================================
# Environment
# =============================================================================

resource "github_repository_environment" "env" {
  repository  = data.github_repository.main.name
  environment = var.environment
}

# =============================================================================
# Environment Variables
# =============================================================================

resource "github_actions_environment_variable" "firestore_database" {
  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "FIRESTORE_DATABASE"
  value         = var.firestore_database
}

resource "github_actions_environment_variable" "storage_bucket" {
  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "STORAGE_BUCKET"
  value         = var.storage_bucket
}
