# =============================================================================
# Artifact Registry Outputs
# =============================================================================

output "repository_id" {
  description = "Artifact RegistryリポジトリID"
  value       = google_artifact_registry_repository.docker.repository_id
}

output "repository_name" {
  description = "Artifact Registryリポジトリ名（フルパス）"
  value       = google_artifact_registry_repository.docker.name
}

output "repository_url" {
  description = "DockerイメージのベースURL"
  value       = "${var.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker.repository_id}"
}
