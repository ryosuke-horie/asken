# =============================================================================
# Firebase Authentication Module - Outputs
# =============================================================================

output "firebase_project_id" {
  description = "FirebaseプロジェクトID"
  value       = google_firebase_project.main.project
}

output "auth_domain" {
  description = "Firebase認証ドメイン"
  value       = "${var.project_id}.firebaseapp.com"
}
