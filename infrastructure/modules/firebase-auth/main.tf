# =============================================================================
# Firebase Authentication Module
# =============================================================================

# Firebaseプロジェクト
resource "google_firebase_project" "main" {
  provider = google-beta
  project  = var.project_id
}

# Identity Platform設定（Firebase Auth）
resource "google_identity_platform_config" "auth" {
  provider = google-beta
  project  = var.project_id

  sign_in {
    allow_duplicate_emails = false

    email {
      enabled           = true
      password_required = true
    }
  }

  # 認可されたドメイン
  authorized_domains = var.authorized_domains

  depends_on = [google_firebase_project.main]
}
