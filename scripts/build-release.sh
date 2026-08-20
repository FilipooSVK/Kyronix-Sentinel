#!/usr/bin/env bash

set -euo pipefail

MODULE="kyronix/sentinel"
GOOS_TARGET="linux"

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <version>"
    echo
    echo "Example:"
    echo "  $0 0.1.1"
    echo "  $0 v0.1.1"
    exit 1
fi

INPUT_VERSION="$1"

VERSION="${INPUT_VERSION#v}"
RELEASE_VERSION="v${VERSION}"

if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
    echo "Invalid version: ${INPUT_VERSION}"
    exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DIST_DIR="${ROOT_DIR}/dist/${RELEASE_VERSION}"

COMMIT="$(
    git -C "${ROOT_DIR}" rev-parse --short=12 HEAD 2>/dev/null ||
    printf 'unknown'
)"

BUILD_DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

LDFLAGS=(
    "-s"
    "-w"
    "-X" "${MODULE}/internal/version.Version=${VERSION}"
    "-X" "${MODULE}/internal/version.Commit=${COMMIT}"
    "-X" "${MODULE}/internal/version.BuildDate=${BUILD_DATE}"
)

echo "Kyronix Sentinel Release Builder"
echo
echo "Version:    ${RELEASE_VERSION}"
echo "Commit:     ${COMMIT}"
echo "Build date: ${BUILD_DATE}"
echo "Target OS:  ${GOOS_TARGET}"
echo

rm -rf "${DIST_DIR}"

mkdir -p "${DIST_DIR}"

build_release() {

    local arch="$1"

    local package_name
    local archive_path
    local checksum_path
    local stage_dir

    package_name="sentinel-${RELEASE_VERSION}-${GOOS_TARGET}-${arch}.tar.gz"

    archive_path="${DIST_DIR}/${package_name}"

    checksum_path="${archive_path}.sha256"

    stage_dir="$(mktemp -d)"

    echo "Building ${GOOS_TARGET}/${arch}..."

    GOOS="${GOOS_TARGET}" \
    GOARCH="${arch}" \
    CGO_ENABLED=0 \
    GOTOOLCHAIN=local \
    go build \
        -buildvcs=false \
        -trimpath \
        -ldflags="${LDFLAGS[*]}" \
        -o "${stage_dir}/sentineld" \
        "${ROOT_DIR}/cmd/sentineld"

    GOOS="${GOOS_TARGET}" \
    GOARCH="${arch}" \
    CGO_ENABLED=0 \
    GOTOOLCHAIN=local \
    go build \
        -buildvcs=false \
        -trimpath \
        -ldflags="${LDFLAGS[*]}" \
        -o "${stage_dir}/sentinelctl" \
        "${ROOT_DIR}/cmd/sentinelctl"

    cp \
        "${ROOT_DIR}/systemd/sentinel-update.service" \
        "${stage_dir}/sentinel-update.service"

    chmod 0755 \
        "${stage_dir}/sentineld" \
        "${stage_dir}/sentinelctl"

    chmod 0644 \
        "${stage_dir}/sentinel-update.service"

    local sentineld_sha
    local sentinelctl_sha
    local update_service_sha

    sentineld_sha="$(
        sha256sum "${stage_dir}/sentineld" |
        awk '{print $1}'
    )"

    sentinelctl_sha="$(
        sha256sum "${stage_dir}/sentinelctl" |
        awk '{print $1}'
    )"

    update_service_sha="$(
        sha256sum "${stage_dir}/sentinel-update.service" |
        awk '{print $1}'
    )"

    cat > "${stage_dir}/manifest.json" <<MANIFEST
{
  "version": "${RELEASE_VERSION}",
  "os": "${GOOS_TARGET}",
  "arch": "${arch}",
  "files": [
    {
      "name": "sentineld",
      "sha256": "${sentineld_sha}"
    },
    {
      "name": "sentinelctl",
      "sha256": "${sentinelctl_sha}"
    },
    {
      "name": "sentinel-update.service",
      "sha256": "${update_service_sha}"
    }
  ]
}
MANIFEST

    tar \
        -C "${stage_dir}" \
        -czf "${archive_path}" \
        manifest.json \
        sentineld \
        sentinelctl \
        sentinel-update.service

    (
        cd "${DIST_DIR}"

        sha256sum \
            "${package_name}" \
            > "${package_name}.sha256"
    )

    echo "Created:"
    echo "  ${archive_path}"
    echo "  ${checksum_path}"

    echo
    echo "Archive contents:"

    tar -tzf "${archive_path}" |
        sed 's/^/  /'

    echo
    echo "Outer SHA256:"

    (
        cd "${DIST_DIR}"

        sha256sum \
            --check \
            "${package_name}.sha256"
    )

    rm -rf "${stage_dir}"

    echo
}

cd "${ROOT_DIR}"

build_release "arm64"
build_release "amd64"

echo "Release build complete."
echo
echo "Artifacts:"

find "${DIST_DIR}" \
    -maxdepth 1 \
    -type f \
    -printf '  %f\n' |
    sort
