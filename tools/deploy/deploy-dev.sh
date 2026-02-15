#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
TF_DEV_DIR="${ROOT_DIR}/infrastructure/environments/dev"

usage() {
  cat <<USAGE
Usage: tools/deploy/deploy-dev.sh [options]

Options:
  --project-id <id>                GCP project ID
  --region <region>                GCP region (default: GCP_REGION or asia-northeast1)
  --artifact-registry-url <url>    Artifact Registry base URL
  --service-name <name>            Cloud Run service name
  --image-tag <tag>                Image tag suffix (default: git short SHA)
  --help                           Show this help

Environment variable fallback:
  GCP_PROJECT_ID, GCP_REGION, ARTIFACT_REGISTRY_URL, CLOUD_RUN_SERVICE_NAME
USAGE
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: required command not found: $1" >&2
    exit 1
  fi
}

tf_output() {
  local output_name="$1"
  if ! command -v terraform >/dev/null 2>&1; then
    return 0
  fi
  if [ ! -d "${TF_DEV_DIR}" ]; then
    return 0
  fi

  terraform -chdir="${TF_DEV_DIR}" output -raw "${output_name}" 2>/dev/null || true
}

PROJECT_ID="${GCP_PROJECT_ID:-$(tf_output project_id)}"
REGION="${GCP_REGION:-asia-northeast1}"
ARTIFACT_REGISTRY_URL="${ARTIFACT_REGISTRY_URL:-$(tf_output artifact_registry_url)}"
CLOUD_RUN_SERVICE_NAME="${CLOUD_RUN_SERVICE_NAME:-$(tf_output cloud_run_service_name)}"
CUSTOM_IMAGE_TAG="${IMAGE_TAG:-}"

while [ $# -gt 0 ]; do
  case "$1" in
    --project-id)
      PROJECT_ID="$2"
      shift 2
      ;;
    --region)
      REGION="$2"
      shift 2
      ;;
    --artifact-registry-url)
      ARTIFACT_REGISTRY_URL="$2"
      shift 2
      ;;
    --service-name)
      CLOUD_RUN_SERVICE_NAME="$2"
      shift 2
      ;;
    --image-tag)
      CUSTOM_IMAGE_TAG="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

require_cmd git
require_cmd gcloud
require_cmd docker

if [ -z "${PROJECT_ID}" ]; then
  echo "ERROR: PROJECT_ID is required. Set GCP_PROJECT_ID or pass --project-id." >&2
  exit 1
fi

if [ -z "${ARTIFACT_REGISTRY_URL}" ]; then
  echo "ERROR: ARTIFACT_REGISTRY_URL is required. Set ARTIFACT_REGISTRY_URL or pass --artifact-registry-url." >&2
  exit 1
fi

if [ -z "${CLOUD_RUN_SERVICE_NAME}" ]; then
  echo "ERROR: CLOUD_RUN_SERVICE_NAME is required. Set CLOUD_RUN_SERVICE_NAME or pass --service-name." >&2
  exit 1
fi

ACTIVE_ACCOUNT="$(gcloud auth list --filter=status:ACTIVE --format='value(account)' | head -n 1)"
if [ -z "${ACTIVE_ACCOUNT}" ]; then
  echo "ERROR: no active gcloud account. Run: gcloud auth login" >&2
  exit 1
fi

if [ -z "${CUSTOM_IMAGE_TAG}" ]; then
  CUSTOM_IMAGE_TAG="$(git -C "${ROOT_DIR}" rev-parse --short=7 HEAD)"
fi

IMAGE_TAG="${ARTIFACT_REGISTRY_URL}/backend:${CUSTOM_IMAGE_TAG}"
LATEST_TAG="${ARTIFACT_REGISTRY_URL}/backend:latest"

echo "=== Deploy configuration ==="
echo "Project: ${PROJECT_ID}"
echo "Region: ${REGION}"
echo "Service: ${CLOUD_RUN_SERVICE_NAME}"
echo "Image: ${IMAGE_TAG}"
echo "Account: ${ACTIVE_ACCOUNT}"

echo "=== Configure Docker auth ==="
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet

echo "=== Build Docker image ==="
docker build \
  --platform linux/amd64 \
  -t "${IMAGE_TAG}" \
  -t "${LATEST_TAG}" \
  "${BACKEND_DIR}"

echo "=== Push Docker image ==="
docker push "${IMAGE_TAG}"
docker push "${LATEST_TAG}"

echo "=== Deploy to Cloud Run ==="
gcloud run deploy "${CLOUD_RUN_SERVICE_NAME}" \
  --image "${IMAGE_TAG}" \
  --region "${REGION}" \
  --platform managed \
  --project "${PROJECT_ID}" \
  --quiet

URL="$(gcloud run services describe "${CLOUD_RUN_SERVICE_NAME}" \
  --region "${REGION}" \
  --project "${PROJECT_ID}" \
  --format 'value(status.url)')"

echo "=== Deploy complete ==="
echo "URL: ${URL}"
