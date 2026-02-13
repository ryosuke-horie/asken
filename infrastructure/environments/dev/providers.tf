# =============================================================================
# Terraform & Provider Configuration
# =============================================================================

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.45.2"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "5.45.2"
    }
    github = {
      source  = "integrations/github"
      version = "6.10.2"
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

provider "google-beta" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

provider "github" {
  token = var.github_token
}
