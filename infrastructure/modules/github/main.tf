# =============================================================================
# GitHub Module
# =============================================================================

terraform {
  required_version = "~> 1.10"

  required_providers {
    github = {
      source  = "integrations/github"
      version = "6.10.2"
    }
  }
}

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

# GCP_SA_KEY は WIF 使用時は不要（後方互換のため条件付きで残す）
resource "github_actions_secret" "gcp_sa_key" {
  count = var.gcp_sa_key != "" ? 1 : 0

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

  deployment_branch_policy {
    protected_branches     = true
    custom_branch_policies = false
  }
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

resource "github_actions_environment_variable" "artifact_registry_url" {
  count = var.artifact_registry_url != "" ? 1 : 0

  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "ARTIFACT_REGISTRY_URL"
  value         = var.artifact_registry_url
}

resource "github_actions_environment_variable" "cloud_run_service_name" {
  count = var.cloud_run_service_name != "" ? 1 : 0

  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "CLOUD_RUN_SERVICE_NAME"
  value         = var.cloud_run_service_name
}

resource "github_actions_environment_variable" "workload_identity_provider" {
  count = var.workload_identity_provider != "" ? 1 : 0

  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "WORKLOAD_IDENTITY_PROVIDER"
  value         = var.workload_identity_provider
}

resource "github_actions_environment_variable" "service_account_email" {
  count = var.service_account_email != "" ? 1 : 0

  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "SERVICE_ACCOUNT_EMAIL"
  value         = var.service_account_email
}

resource "github_actions_environment_variable" "deploy_service_account_email" {
  repository    = data.github_repository.main.name
  environment   = github_repository_environment.env.environment
  variable_name = "DEPLOY_SERVICE_ACCOUNT_EMAIL"
  value         = var.deploy_service_account_email
}

# =============================================================================
# Environment Secrets
# =============================================================================

resource "github_actions_environment_secret" "e2e_firebase_api_key" {
  count = var.e2e_firebase_api_key != "" ? 1 : 0

  repository      = data.github_repository.main.name
  environment     = github_repository_environment.env.environment
  secret_name     = "E2E_FIREBASE_API_KEY"
  plaintext_value = var.e2e_firebase_api_key
}
