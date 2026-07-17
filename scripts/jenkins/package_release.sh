#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ARTIFACTS_DIR="${ARTIFACTS_DIR:-${ROOT_DIR}/artifacts}"
APP_NAME="${APP_NAME:-zcyp-im}"
BUILD_VERSION="${BUILD_VERSION:-${BUILD_NUMBER:-$(git -C "${ROOT_DIR}" rev-parse --short HEAD)}}"
PACKAGE_NAME="${APP_NAME}-${BUILD_VERSION}"
PACKAGE_DIR="${ARTIFACTS_DIR}/${PACKAGE_NAME}"
PACKAGE_FILE="${ARTIFACTS_DIR}/${PACKAGE_NAME}.tar.gz"
RUN_TESTS="${RUN_TESTS:-true}"

mkdir -p "${ARTIFACTS_DIR}"
rm -rf "${PACKAGE_DIR}" "${PACKAGE_FILE}"
mkdir -p "${PACKAGE_DIR}/bin" "${PACKAGE_DIR}/configs" "${PACKAGE_DIR}/migrations"

if [[ "${RUN_TESTS}" == "true" ]]; then
  (cd "${ROOT_DIR}" && make test)
fi

(cd "${ROOT_DIR}" && make build)

cp "${ROOT_DIR}/zcyp-im" "${PACKAGE_DIR}/bin/"
cp "${ROOT_DIR}/zcyp-im-ws" "${PACKAGE_DIR}/bin/"
cp -R "${ROOT_DIR}/configs/." "${PACKAGE_DIR}/configs/"
cp -R "${ROOT_DIR}/migrations/." "${PACKAGE_DIR}/migrations/"
cp "${ROOT_DIR}/.env.example" "${PACKAGE_DIR}/.env.example"

cat > "${PACKAGE_DIR}/REVISION" <<EOF
build_version=${BUILD_VERSION}
git_commit=$(git -C "${ROOT_DIR}" rev-parse HEAD)
build_time=$(date '+%Y-%m-%d %H:%M:%S %z')
EOF

tar -C "${ARTIFACTS_DIR}" -czf "${PACKAGE_FILE}" "${PACKAGE_NAME}"
echo "created ${PACKAGE_FILE}"
