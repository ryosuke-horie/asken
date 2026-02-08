# =============================================================================
# Cloud Run Outputs
# =============================================================================

output "service_name" {
  description = "Cloud Runサービス名"
  value       = google_cloud_run_v2_service.api.name
}

output "service_url" {
  description = "Cloud RunサービスのURL"
  value       = google_cloud_run_v2_service.api.uri
}

# ランタイムSA
output "service_account_email" {
  description = "ランタイムサービスアカウントのメールアドレス（Cloud Run実行ID。E2EテストではFirebase CustomToken署名のServiceAccountIDとしても使用）"
  value       = google_service_account.runtime.email
}

output "service_account_id" {
  description = "ランタイムサービスアカウントのフルパス"
  value       = google_service_account.runtime.id
}

# デプロイSA
output "deploy_service_account_email" {
  description = "デプロイサービスアカウントのメールアドレス"
  value       = google_service_account.deploy.email
}

output "deploy_service_account_id" {
  description = "デプロイサービスアカウントのフルパス"
  value       = google_service_account.deploy.id
}

output "location" {
  description = "Cloud Runのリージョン"
  value       = google_cloud_run_v2_service.api.location
}
