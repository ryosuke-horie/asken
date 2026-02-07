# Security Reviewer Memory

## Project Overview
- iOS + Go backend app (food nutrition tracker)
- Firestore database, Cloud Storage for images, Gemini API for food analysis
- Firebase Auth for authentication
- Infrastructure managed by Terraform, deployed on Cloud Run via GitHub Actions

## Key Security Architecture
- Secrets managed via 1Password (mise.toml uses `op read`) and GCP Secret Manager (Gemini API key)
- WIF (Workload Identity Federation) for GitHub Actions -> GCP (no SA key needed in CI)
- .gitignore properly excludes .env, *.tfvars, sa-key.json, service-account*.json, GoogleService-Info.plist
- Terraform state stored in GCS bucket (utikomi-dev-tfstate)
- Dev auth bypass via APP_ENV=development + hardcoded mock token

## Known Issues Found (2026-02-07)
- Cloud Storage CORS set to ["*"] in dev
- No rate limiting on API endpoints
- GCP Project Number exposed in iOS Info.plist (OAuth Client ID)
- Cloud Run allows unauthenticated requests (allUsers invoker) - by design, app-level auth
- debugPrint statements in iOS production code
- No backend .env.example file found (permission denied or missing)
- Dev mock token hardcoded as constant "dev-mock-token"

## File Patterns
- Backend: backend/ (Go, layered architecture: handler -> service -> repository)
- iOS: ios/Uchikomi/ (SwiftUI)
- Infrastructure: infrastructure/environments/dev/ + infrastructure/modules/
- CI/CD: .github/workflows/ci.yml + deploy.yml
