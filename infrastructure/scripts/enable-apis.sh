#!/bin/bash
# =============================================================================
# GCP API 有効化スクリプト
# =============================================================================
# 使用方法: ./enable-apis.sh <project_id>
# 例: ./enable-apis.sh utikomi-dev
# =============================================================================

set -euo pipefail

PROJECT_ID=${1:?'Usage: ./enable-apis.sh <project_id>'}

echo "Enabling APIs for project: $PROJECT_ID"

# 必要なAPIを有効化
gcloud services enable \
  firestore.googleapis.com \
  storage.googleapis.com \
  firebase.googleapis.com \
  identitytoolkit.googleapis.com \
  aiplatform.googleapis.com \
  serviceusage.googleapis.com \
  cloudbuild.googleapis.com \
  --project="$PROJECT_ID"

echo ""
echo "Successfully enabled APIs for project: $PROJECT_ID"
echo ""
echo "Enabled APIs:"
echo "  - firestore.googleapis.com (Firestore)"
echo "  - storage.googleapis.com (Cloud Storage)"
echo "  - firebase.googleapis.com (Firebase)"
echo "  - identitytoolkit.googleapis.com (Firebase Auth)"
echo "  - aiplatform.googleapis.com (Gemini API)"
echo "  - serviceusage.googleapis.com (Service Usage)"
echo "  - cloudbuild.googleapis.com (Cloud Build)"
