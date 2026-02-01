# =============================================================================
# Cloud Storage Module - Outputs
# =============================================================================

output "bucket_name" {
  description = "バケット名"
  value       = google_storage_bucket.images.name
}

output "bucket_url" {
  description = "バケットURL"
  value       = google_storage_bucket.images.url
}

output "bucket_self_link" {
  description = "バケットのself_link"
  value       = google_storage_bucket.images.self_link
}
