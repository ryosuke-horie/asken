# =============================================================================
# GitHub Module - Outputs
# =============================================================================

output "environment_name" {
  description = "GitHub Environment名"
  value       = github_repository_environment.env.environment
}

output "repository_name" {
  description = "GitHubリポジトリ名"
  value       = data.github_repository.main.name
}
