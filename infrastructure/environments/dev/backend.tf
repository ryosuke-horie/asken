# =============================================================================
# Terraform Backend Configuration
# =============================================================================

terraform {
  backend "gcs" {
    bucket = "utikomi-dev-tfstate"
    prefix = "terraform/state/dev"
  }
}
