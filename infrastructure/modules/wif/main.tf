# =============================================================================
# Workload Identity Federation - GitHub Actions用キーレス認証
# =============================================================================

# -----------------------------------------------------------------------------
# Workload Identity Pool
# -----------------------------------------------------------------------------

resource "google_iam_workload_identity_pool" "github" {
  project                   = var.project_id
  workload_identity_pool_id = var.pool_id
  display_name              = "GitHub Actions"
  description               = "Workload Identity Pool for GitHub Actions"
  disabled                  = false
}

# -----------------------------------------------------------------------------
# OIDC Provider (GitHub)
# -----------------------------------------------------------------------------

resource "google_iam_workload_identity_pool_provider" "github" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.provider_id
  display_name                       = "GitHub"
  description                        = "OIDC Provider for GitHub Actions"
  disabled                           = false

  # GitHubのOIDCトークンからの属性マッピング
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.aud"        = "assertion.aud"
    "attribute.repository" = "assertion.repository"
  }

  # 特定リポジトリのみに制限（セキュリティ強化）
  attribute_condition = "assertion.repository == '${var.github_owner}/${var.github_repo}'"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

# -----------------------------------------------------------------------------
# サービスアカウントへのWorkload Identity User権限付与
# -----------------------------------------------------------------------------

resource "google_service_account_iam_member" "workload_identity_user" {
  service_account_id = var.service_account_id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.github_owner}/${var.github_repo}"
}
