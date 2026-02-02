#!/bin/bash
# =============================================================================
# GCP API 有効化スクリプト
# =============================================================================
# 使用方法: ./enable-apis.sh <project_id>
# 例: ./enable-apis.sh utikomi-dev
# =============================================================================

set -euo pipefail

# -----------------------------------------------------------------------------
# 前提条件チェック
# -----------------------------------------------------------------------------

# gcloud CLIの存在確認
if ! command -v gcloud &> /dev/null; then
  echo "ERROR: gcloud CLI is not installed"
  echo "Please install: https://cloud.google.com/sdk/docs/install"
  exit 1
fi

# gcloud認証状態の確認
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>/dev/null | head -n1 | grep -q .; then
  echo "ERROR: Not authenticated with gcloud"
  echo "Please run: gcloud auth login"
  exit 1
fi

# プロジェクトID引数チェック
PROJECT_ID=${1:?'Usage: ./enable-apis.sh <project_id>'}

# プロジェクト存在確認
if ! gcloud projects describe "$PROJECT_ID" &>/dev/null; then
  echo "ERROR: Project '$PROJECT_ID' does not exist or you don't have access"
  exit 1
fi

# -----------------------------------------------------------------------------
# API有効化
# -----------------------------------------------------------------------------

APIS=(
  "firestore.googleapis.com"
  "storage.googleapis.com"
  "firebase.googleapis.com"
  "identitytoolkit.googleapis.com"
  "aiplatform.googleapis.com"
  "serviceusage.googleapis.com"
  "cloudbuild.googleapis.com"
)

API_DESCRIPTIONS=(
  "Firestore"
  "Cloud Storage"
  "Firebase"
  "Firebase Auth"
  "Gemini API"
  "Service Usage"
  "Cloud Build"
)

echo "Enabling APIs for project: $PROJECT_ID"
echo ""

FAILED_APIS=()
SUCCEEDED_APIS=()

for i in "${!APIS[@]}"; do
  api="${APIS[$i]}"
  desc="${API_DESCRIPTIONS[$i]}"
  echo -n "Enabling $api ($desc)... "

  if gcloud services enable "$api" --project="$PROJECT_ID" 2>&1; then
    echo "OK"
    SUCCEEDED_APIS+=("$api")
  else
    echo "FAILED"
    FAILED_APIS+=("$api")
  fi
done

echo ""

# 結果サマリー
if [ ${#FAILED_APIS[@]} -gt 0 ]; then
  echo "=========================================="
  echo "ERROR: Failed to enable the following APIs:"
  echo "=========================================="
  for api in "${FAILED_APIS[@]}"; do
    echo "  - $api"
  done
  echo ""
  echo "Please check:"
  echo "  1. Billing account is linked to the project"
  echo "  2. You have sufficient permissions (roles/serviceusage.serviceUsageAdmin)"
  echo "  3. The API is available in your region"
  echo ""
  exit 1
fi

echo "=========================================="
echo "Successfully enabled all APIs for project: $PROJECT_ID"
echo "=========================================="
echo ""
echo "Enabled APIs:"
for i in "${!APIS[@]}"; do
  echo "  - ${APIS[$i]} (${API_DESCRIPTIONS[$i]})"
done
