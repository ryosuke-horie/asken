# =============================================================================
# Artifact Registry - Dockerイメージリポジトリ
# =============================================================================

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

resource "google_artifact_registry_repository" "docker" {
  project       = var.project_id
  location      = var.location
  repository_id = var.repository_id
  description   = var.description
  format        = "DOCKER"

  # クリーンアップポリシー（開発環境用）
  dynamic "cleanup_policies" {
    for_each = var.cleanup_policy_keep_count > 0 ? [1] : []
    content {
      id     = "keep-recent"
      action = "KEEP"
      most_recent_versions {
        keep_count = var.cleanup_policy_keep_count
      }
    }
  }

  labels = var.labels
}

# GitHub ActionsからのPush権限（サービスアカウント）
resource "google_artifact_registry_repository_iam_member" "github_actions_writer" {
  count = var.github_actions_sa_email != "" ? 1 : 0

  project    = var.project_id
  location   = var.location
  repository = google_artifact_registry_repository.docker.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${var.github_actions_sa_email}"
}

# Note: Cloud Runはサービスエージェント（serverless-robot-prod）を使用して
# Artifact Registryからイメージを取得するため、Cloud Runサービスアカウントへの
# 明示的なreader権限は不要です。
