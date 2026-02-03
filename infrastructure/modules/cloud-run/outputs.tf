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

output "service_account_email" {
  description = "Cloud Runサービスアカウントのメールアドレス"
  value       = google_service_account.cloud_run.email
}

output "service_account_id" {
  description = "Cloud Runサービスアカウントのフルパス"
  value       = google_service_account.cloud_run.id
}

output "location" {
  description = "Cloud Runのリージョン"
  value       = google_cloud_run_v2_service.api.location
}
