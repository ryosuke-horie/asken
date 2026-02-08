# =============================================================================
# Cloud Run - バックエンドAPIサービス
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

# -----------------------------------------------------------------------------
# ランタイムサービスアカウント（Cloud Run実行用）
# -----------------------------------------------------------------------------

resource "google_service_account" "runtime" {
  project      = var.project_id
  account_id   = "${var.service_name}-sa"
  display_name = "Cloud Run Runtime Service Account for ${var.service_name}"
  description  = "Cloud Runサービスのランタイム用サービスアカウント"
}

# Firestoreアクセス権限
resource "google_project_iam_member" "runtime_firestore_user" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.runtime.email}"
}

# Cloud Storageアクセス権限（読み書きのみ、IAMポリシー管理は不要）
resource "google_project_iam_member" "runtime_storage_object_user" {
  project = var.project_id
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.runtime.email}"
}

# Firebase Auth検証権限（トークン検証のみ）
resource "google_project_iam_member" "runtime_firebase_auth" {
  project = var.project_id
  role    = "roles/firebaseauth.viewer"
  member  = "serviceAccount:${google_service_account.runtime.email}"
}

# Secret Managerアクセス権限
resource "google_project_iam_member" "runtime_secret_accessor" {
  count = length(var.secrets) > 0 ? 1 : 0

  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.runtime.email}"
}

# 署名付きURL生成権限（Cloud Storage SignedURL用）
# SA自身に対するToken Creator権限 - BucketHandle.SignedURL()がIAM signBlobで署名するために必要
resource "google_service_account_iam_member" "runtime_self_token_creator" {
  service_account_id = google_service_account.runtime.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_service_account.runtime.email}"
}

# -----------------------------------------------------------------------------
# デプロイサービスアカウント（CI/CD用）
# -----------------------------------------------------------------------------

resource "google_service_account" "deploy" {
  project      = var.project_id
  account_id   = "${var.service_name}-deploy-sa"
  display_name = "Cloud Run Deploy Service Account for ${var.service_name}"
  description  = "CI/CDパイプラインからのデプロイ用サービスアカウント"
}

# Artifact Registry書き込み権限（GitHub ActionsからのDockerイメージプッシュ用）
resource "google_project_iam_member" "deploy_artifact_registry_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.deploy.email}"
}

# Cloud Runデプロイ権限（GitHub Actionsからのデプロイ用）
resource "google_project_iam_member" "deploy_run_developer" {
  project = var.project_id
  role    = "roles/run.developer"
  member  = "serviceAccount:${google_service_account.deploy.email}"
}

# デプロイSAがランタイムSAを指定してCloud Runをデプロイする権限
# Cloud Runデプロイ時にランタイムSAをサービスアカウントとして設定するために必要
resource "google_service_account_iam_member" "deploy_acts_as_runtime" {
  service_account_id = google_service_account.runtime.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.deploy.email}"
}

# デプロイSAがランタイムSAに対するToken Creator権限を持つ
# E2EテストでFirebase CustomToken署名時に、デプロイSAがランタイムSAのsignBlob APIを呼び出すために必要
resource "google_service_account_iam_member" "deploy_token_creator_on_runtime" {
  service_account_id = google_service_account.runtime.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_service_account.deploy.email}"
}

# -----------------------------------------------------------------------------
# Cloud Runサービス
# -----------------------------------------------------------------------------

resource "google_cloud_run_v2_service" "api" {
  project  = var.project_id
  name     = var.service_name
  location = var.location
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.runtime.email

    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }

    containers {
      image = var.container_image

      resources {
        limits = {
          cpu    = var.cpu_limit
          memory = var.memory_limit
        }
        cpu_idle          = true
        startup_cpu_boost = true
      }

      # 環境変数
      dynamic "env" {
        for_each = var.env_vars
        content {
          name  = env.key
          value = env.value
        }
      }

      # Secret Managerからの環境変数
      dynamic "env" {
        for_each = var.secrets
        content {
          name = env.key
          value_source {
            secret_key_ref {
              secret  = env.value.secret_name
              version = env.value.version
            }
          }
        }
      }

      ports {
        container_port = var.container_port
      }

      # ヘルスチェック
      startup_probe {
        initial_delay_seconds = 10
        timeout_seconds       = 5
        period_seconds        = 10
        failure_threshold     = 3
        http_get {
          path = var.health_check_path
          port = var.container_port
        }
      }

      liveness_probe {
        timeout_seconds   = 5
        period_seconds    = 30
        failure_threshold = 3
        http_get {
          path = var.health_check_path
          port = var.container_port
        }
      }
    }

    # タイムアウト設定
    timeout = "${var.request_timeout_seconds}s"

    # 最大同時リクエスト数
    max_instance_request_concurrency = var.max_concurrent_requests

    # VPCコネクタ（必要に応じて）
    dynamic "vpc_access" {
      for_each = var.vpc_connector != "" ? [1] : []
      content {
        connector = var.vpc_connector
        egress    = "PRIVATE_RANGES_ONLY"
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = var.labels

  lifecycle {
    ignore_changes = [
      # デプロイ時に変更されるイメージタグを無視
      template[0].containers[0].image,
    ]
  }

  # Cloud Runサービスが起動時にSignedURL生成を試みるため、事前にIAM権限が必要
  depends_on = [
    google_service_account_iam_member.runtime_self_token_creator,
  ]
}

# -----------------------------------------------------------------------------
# 公開アクセス許可（未認証リクエストを許可）
# -----------------------------------------------------------------------------

resource "google_cloud_run_v2_service_iam_member" "public_access" {
  count = var.allow_unauthenticated ? 1 : 0

  project  = var.project_id
  location = var.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
