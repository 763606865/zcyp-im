#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <release-tar.gz>" >&2
  exit 1
fi

PACKAGE_FILE="$1"
DEPLOY_ROOT="${DEPLOY_ROOT:-/home/data/go/zcyp-im}"
APP_NAME="${APP_NAME:-zcyp-im}"
RUN_MIGRATIONS="${RUN_MIGRATIONS:-true}"
MIGRATE_BIN="${MIGRATE_BIN:-migrate}"
SERVICE_API="${SERVICE_API:-zcyp-im-api}"
SERVICE_WS="${SERVICE_WS:-zcyp-im-ws}"
SHARED_ENV_FILE="${SHARED_ENV_FILE:-${DEPLOY_ROOT}/shared/.env}"
KEEP_RELEASES="${KEEP_RELEASES:-5}"
TIMESTAMP="$(date '+%Y%m%d%H%M%S')"
RELEASES_DIR="${DEPLOY_ROOT}/releases"
SHARED_DIR="${DEPLOY_ROOT}/shared"
CURRENT_LINK="${DEPLOY_ROOT}/current"
RELEASE_DIR="${RELEASES_DIR}/${TIMESTAMP}"

if [[ ! -f "${PACKAGE_FILE}" ]]; then
  echo "package not found: ${PACKAGE_FILE}" >&2
  exit 1
fi

if [[ ! -f "${SHARED_ENV_FILE}" ]]; then
  echo "shared env file not found: ${SHARED_ENV_FILE}" >&2
  exit 1
fi

mkdir -p "${RELEASES_DIR}" "${SHARED_DIR}"
tar -C "${RELEASE_DIR%/*}" -xzf "${PACKAGE_FILE}"

EXTRACTED_DIR="$(tar -tzf "${PACKAGE_FILE}" | head -1 | cut -d/ -f1)"
if [[ -z "${EXTRACTED_DIR}" ]]; then
  echo "invalid package layout: ${PACKAGE_FILE}" >&2
  exit 1
fi

mv "${RELEASES_DIR}/${EXTRACTED_DIR}" "${RELEASE_DIR}"
ln -sfn "${SHARED_ENV_FILE}" "${RELEASE_DIR}/.env"

 if [[ "${RUN_MIGRATIONS}" == "true" ]]; then
  (
    cd "${RELEASE_DIR}"
    set -a
    . ./.env
    set +a
    "${MIGRATE_BIN}" -path migrations -database "mysql://${ZCYP_IM_MYSQL_USERNAME}:${ZCYP_IM_MYSQL_PASSWORD}@tcp(${ZCYP_IM_MYSQL_HOST}:${ZCYP_IM_MYSQL_PORT})/${ZCYP_IM_MYSQL_DATABASE}" up
  )
fi

ln -sfn "${RELEASE_DIR}" "${CURRENT_LINK}"
systemctl restart "${SERVICE_API}"
systemctl restart "${SERVICE_WS}"
systemctl is-active --quiet "${SERVICE_API}"
systemctl is-active --quiet "${SERVICE_WS}"

find "${RELEASES_DIR}" -mindepth 1 -maxdepth 1 -type d | sort | head -n -"${KEEP_RELEASES}" | xargs -r rm -rf

echo "deployed ${APP_NAME} to ${RELEASE_DIR}"
