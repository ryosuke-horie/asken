# =============================================================================
# Cloud Storage Module
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

# 画像保存用バケット
resource "google_storage_bucket" "images" {
  project       = var.project_id
  name          = "${var.project_id}-images"
  location      = var.location
  storage_class = "STANDARD"

  # バージョニング（devは無効、prodは有効）
  versioning {
    enabled = var.enable_versioning
  }

  # ライフサイクルルール（古い画像の自動削除）
  dynamic "lifecycle_rule" {
    for_each = var.lifecycle_delete_days > 0 ? [1] : []
    content {
      condition {
        age = var.lifecycle_delete_days
      }
      action {
        type = "Delete"
      }
    }
  }

  # CORS設定（iOSアプリからのアクセス用）
  cors {
    origin          = var.cors_origins
    method          = ["GET", "PUT", "POST", "DELETE", "HEAD"]
    response_header = ["Content-Type", "Content-Length", "Content-MD5"]
    max_age_seconds = 3600
  }

  uniform_bucket_level_access = true

  # 削除時にバケット内のオブジェクトも削除（devのみ）
  force_destroy = var.force_destroy
}

# バケットへのIAMバインディング（サービスアカウント用）
resource "google_storage_bucket_iam_member" "storage_admin" {
  count  = var.service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.images.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.service_account_email}"
}
