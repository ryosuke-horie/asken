#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
TF_DEV_DIR="${ROOT_DIR}/infrastructure/environments/dev"

usage() {
  cat <<USAGE
Usage: tools/e2e/run-backend-e2e-dev.sh [options]

Options:
  --base-url <url>         E2E target base URL (e.g. https://...a.run.app)
  --project-id <id>        GCP project ID (base URL auto-resolveに使用)
  --region <region>        GCP region (default: GCP_REGION or asia-northeast1)
  --service-name <name>    Cloud Run service name (base URL auto-resolveに使用)
  --test-uid <uid>         Test UID (default: E2E_TEST_UID or e2e-test-user)
  --run-gemini             Enable Gemini API tests (default: skipped)
  --help                   Show this help

Environment variable fallback:
  E2E_BASE_URL, GCP_PROJECT_ID, GCP_REGION, CLOUD_RUN_SERVICE_NAME,
  E2E_FIREBASE_API_KEY, SERVICE_ACCOUNT_EMAIL, E2E_TEST_UID, E2E_RUN_GEMINI

Note: Gemini API tests are skipped by default to avoid API costs and rate limits.
      Set E2E_RUN_GEMINI=true or pass --run-gemini to enable them.
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

E2E_BASE_URL="${E2E_BASE_URL:-$(tf_output cloud_run_url)}"
PROJECT_ID="${GCP_PROJECT_ID:-$(tf_output project_id)}"
REGION="${GCP_REGION:-asia-northeast1}"
CLOUD_RUN_SERVICE_NAME="${CLOUD_RUN_SERVICE_NAME:-$(tf_output cloud_run_service_name)}"
TEST_UID="${E2E_TEST_UID:-e2e-test-user}"
RUN_GEMINI="${E2E_RUN_GEMINI:-}"

while [ $# -gt 0 ]; do
  case "$1" in
    --base-url)
      E2E_BASE_URL="$2"
      shift 2
      ;;
    --project-id)
      PROJECT_ID="$2"
      shift 2
      ;;
    --region)
      REGION="$2"
      shift 2
      ;;
    --service-name)
      CLOUD_RUN_SERVICE_NAME="$2"
      shift 2
      ;;
    --test-uid)
      TEST_UID="$2"
      shift 2
      ;;
    --run-gemini)
      RUN_GEMINI="true"
      shift
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

require_cmd go

if [ -z "${E2E_BASE_URL}" ]; then
  if [ -z "${PROJECT_ID}" ] || [ -z "${CLOUD_RUN_SERVICE_NAME}" ]; then
    echo "ERROR: E2E_BASE_URL is required. Set E2E_BASE_URL or pass --base-url." >&2
    echo "       Alternatively set GCP_PROJECT_ID and CLOUD_RUN_SERVICE_NAME for auto resolution." >&2
    exit 1
  fi

  require_cmd gcloud

  E2E_BASE_URL="$(gcloud run services describe "${CLOUD_RUN_SERVICE_NAME}" \
    --region "${REGION}" \
    --project "${PROJECT_ID}" \
    --format 'value(status.url)')"
fi

if [ -z "${E2E_BASE_URL}" ]; then
  echo "ERROR: failed to resolve E2E_BASE_URL" >&2
  exit 1
fi

echo "=== Backend E2E configuration ==="
echo "Base URL: ${E2E_BASE_URL}"
echo "Test UID: ${TEST_UID}"
if [ -n "${E2E_FIREBASE_API_KEY:-}" ]; then
  echo "Auth E2E: enabled (E2E_FIREBASE_API_KEY is set)"
else
  echo "Auth E2E: partial mode (E2E_FIREBASE_API_KEY is not set)"
fi
if [ "${RUN_GEMINI}" = "true" ]; then
  echo "Gemini E2E: enabled (E2E_RUN_GEMINI=true)"
else
  echo "Gemini E2E: skipped (set E2E_RUN_GEMINI=true or --run-gemini to enable)"
fi

echo "=== Run backend E2E tests ==="
(
  cd "${BACKEND_DIR}"
  E2E_BASE_URL="${E2E_BASE_URL}" \
  E2E_FIREBASE_API_KEY="${E2E_FIREBASE_API_KEY:-}" \
  SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_EMAIL:-}" \
  E2E_TEST_UID="${TEST_UID}" \
  E2E_RUN_GEMINI="${RUN_GEMINI}" \
  go test -v -tags=e2e ./e2e/...
)
