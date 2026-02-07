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

## CI/CD・インフラ診断結果 (2026-02-07)
- CRITICAL: Cloud Run SAにランタイム+CI/CD権限が集約（SA分離が必要）
- HIGH: サードパーティActionsがSHAピンニングされていない
- HIGH: WIFにブランチ制限なし（任意ブランチからGCP認証可能）
- HIGH: IAM権限がプロジェクトレベル（リソースレベルに変更推奨）
- HIGH: SwiftLint/SwiftFormatダウンロード時のチェックサム検証なし
- HIGH: deploy.ymlにトップレベルpermissions未定義
- MEDIUM: Cloud Run ingress=INGRESS_TRAFFIC_ALL
- MEDIUM: Terraform stateに機密データ格納（CMEK/アクセス制限確認必要）
- MEDIUM: persist-credentials: false未設定
- Dockerfile: distroless+nonroot+マルチステージで良好

## File Patterns
- Backend: backend/ (Go, layered architecture: handler -> service -> repository)
- iOS: ios/Uchikomi/ (SwiftUI)
- Infrastructure: infrastructure/environments/dev/ + infrastructure/modules/
- CI/CD: .github/workflows/ci.yml + deploy.yml
