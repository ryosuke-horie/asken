# =============================================================================
# Firebase Authentication Module
# =============================================================================

terraform {
  required_version = "~> 1.10"

  required_providers {
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "5.45.2"
    }
  }
}

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

# =============================================================================
# Google Sign-In プロバイダ設定
# =============================================================================
# NOTE: 個人プロジェクトではIAP Brandを作成できないため、
# OAuth Client IDは手動で作成する必要があります。
# 手順:
# 1. https://console.cloud.google.com/apis/credentials?project=<PROJECT_ID>
# 2. OAuth同意画面を設定（External）
# 3. OAuth 2.0 Client IDを作成（iOS用）
# 4. CLIENT_IDをGoogleService-Info.plistに追加
# 5. Firebase ConsoleでGoogle Sign-Inを有効化

resource "google_identity_platform_default_supported_idp_config" "google" {
  count         = var.google_oauth_client_id != "" ? 1 : 0
  provider      = google-beta
  project       = var.project_id
  idp_id        = "google.com"
  enabled       = true
  client_id     = var.google_oauth_client_id
  client_secret = var.google_oauth_client_secret

  depends_on = [google_identity_platform_config.auth]
}
