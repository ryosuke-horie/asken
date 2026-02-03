# =============================================================================
# Workload Identity Federation Outputs
# =============================================================================

output "pool_id" {
  description = "Workload Identity Pool ID"
  value       = google_iam_workload_identity_pool.github.workload_identity_pool_id
}

output "pool_name" {
  description = "Workload Identity Pool フルパス"
  value       = google_iam_workload_identity_pool.github.name
}

output "provider_id" {
  description = "Workload Identity Provider ID"
  value       = google_iam_workload_identity_pool_provider.github.workload_identity_pool_provider_id
}

output "provider_name" {
  description = "Workload Identity Provider フルパス（GitHub Actionsで使用）"
  value       = google_iam_workload_identity_pool_provider.github.name
}
