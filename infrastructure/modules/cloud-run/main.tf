# =============================================================================
# Cloud Run - バックエンドAPIサービス
# =============================================================================

# -----------------------------------------------------------------------------
# サービスアカウント
# -----------------------------------------------------------------------------

resource "google_service_account" "cloud_run" {
  project      = var.project_id
  account_id   = "${var.service_name}-sa"
  display_name = "Cloud Run Service Account for ${var.service_name}"
  description  = "Cloud Runサービス用のサービスアカウント"
}

# Firestoreアクセス権限
resource "google_project_iam_member" "firestore_user" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Cloud Storageアクセス権限（読み書きのみ、IAMポリシー管理は不要）
resource "google_project_iam_member" "storage_object_user" {
  project = var.project_id
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Firebase Auth検証権限（トークン検証のみ）
resource "google_project_iam_member" "firebase_auth" {
  project = var.project_id
  role    = "roles/firebaseauth.viewer"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Secret Managerアクセス権限
resource "google_project_iam_member" "secret_accessor" {
  count = length(var.secrets) > 0 ? 1 : 0

  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Artifact Registry書き込み権限（GitHub ActionsからのDockerイメージプッシュ用）
# WIF経由でこのサービスアカウントを使用してイメージをプッシュするため必要
resource "google_project_iam_member" "artifact_registry_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Cloud Runデプロイ権限（GitHub Actionsからのデプロイ用）
# WIF経由でこのサービスアカウントを使用してCloud Runにデプロイするため必要
resource "google_project_iam_member" "run_developer" {
  project = var.project_id
  role    = "roles/run.developer"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
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
    service_account = google_service_account.cloud_run.email

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
