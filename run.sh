#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${COSIGHT_ENV_FILE:-${ROOT_DIR}/.env}"
API_PID=""
WEB_PID=""

cleanup() {
  trap - INT TERM EXIT
  if [[ -n "${API_PID}" ]] && kill -0 "${API_PID}" 2>/dev/null; then
    kill "${API_PID}" 2>/dev/null || true
  fi
  if [[ -n "${WEB_PID}" ]] && kill -0 "${WEB_PID}" 2>/dev/null; then
    kill "${WEB_PID}" 2>/dev/null || true
  fi
  wait 2>/dev/null || true
}

trap cleanup INT TERM EXIT

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "환경 설정 파일을 찾을 수 없습니다: ${ENV_FILE}" >&2
  echo "먼저 'cp .env.example .env'를 실행하고 Keycloak 정보를 입력하세요." >&2
  exit 1
fi

for command_name in go pnpm; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "필수 명령을 찾을 수 없습니다: ${command_name}" >&2
    exit 1
  fi
done

set -a
# shellcheck disable=SC1090
source "${ENV_FILE}"
set +a

required_variables=(
  COSIGHT_KEYCLOAK_URL
  COSIGHT_KEYCLOAK_REALM
  COSIGHT_KEYCLOAK_CLIENT_ID
)

for variable_name in "${required_variables[@]}"; do
  if [[ -z "${!variable_name:-}" ]]; then
    echo "필수 환경변수가 비어 있습니다: ${variable_name}" >&2
    exit 1
  fi
done

if [[ ! -d "${ROOT_DIR}/web/node_modules" ]]; then
  echo "프런트엔드 의존성을 설치합니다."
  (cd "${ROOT_DIR}/web" && pnpm install)
fi

echo "Cosight API를 시작합니다: ${COSIGHT_HTTP_ADDRESS:-:8080}"
(cd "${ROOT_DIR}" && go run ./cmd/api) &
API_PID=$!

echo "Cosight Web을 시작합니다: http://localhost:${COSIGHT_WEB_PORT:-5173}"
(cd "${ROOT_DIR}/web" && pnpm dev --host "${COSIGHT_WEB_HOST:-127.0.0.1}" --port "${COSIGHT_WEB_PORT:-5173}") &
WEB_PID=$!

while kill -0 "${API_PID}" 2>/dev/null && kill -0 "${WEB_PID}" 2>/dev/null; do
  sleep 1
done

echo "개발 서버 중 하나가 종료되어 나머지 프로세스를 정리합니다." >&2
exit 1
