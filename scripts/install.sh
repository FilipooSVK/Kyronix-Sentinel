#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_OWNER="FilipooSVK"
REPOSITORY_NAME="Kyronix-Sentinel"

GITHUB_REPO="${REPOSITORY_OWNER}/${REPOSITORY_NAME}"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}"
RAW_BASE="https://raw.githubusercontent.com/${GITHUB_REPO}"

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/sentinel"
CONFIG_PATH="${CONFIG_DIR}/sentinel.yaml"
SYSTEMD_DIR="/etc/systemd/system"

SENTINELD_SERVICE="${SYSTEMD_DIR}/sentineld.service"
UPDATE_SERVICE="${SYSTEMD_DIR}/sentinel-update.service"

STATE_DIR="/var/lib/sentinel"

log() {
    printf '==> %s\n' "$*"
}

fail() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

if [[ "${EUID}" -ne 0 ]]; then
    fail "installer must run as root (use sudo)"
fi

for command in curl tar sha256sum install systemctl uname sed awk grep sort; do
    command -v "${command}" >/dev/null 2>&1 ||
        fail "required command not found: ${command}"
done

case "$(uname -m)" in
    aarch64|arm64)
        ARCH="arm64"
        ;;
    x86_64|amd64)
        ARCH="amd64"
        ;;
    *)
        fail "unsupported architecture: $(uname -m)"
        ;;
esac

log "Detected architecture: ${ARCH}"

log "Discovering latest stable GitHub release"

VERSION="$(
    curl -fsSL \
        -H "Accept: application/vnd.github+json" \
        "${GITHUB_API}/releases/latest" |
    sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' |
    head -n 1
)"

[[ -n "${VERSION}" ]] ||
    fail "could not determine latest GitHub release"

[[ "${VERSION}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] ||
    fail "unexpected release version: ${VERSION}"

log "Latest stable release: ${VERSION}"

PACKAGE="sentinel-${VERSION}-linux-${ARCH}.tar.gz"
CHECKSUM="${PACKAGE}.sha256"

RELEASE_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}"

WORK_DIR="$(mktemp -d -t sentinel-install-XXXXXXXX)"

cleanup() {
    rm -rf "${WORK_DIR}"
}

trap cleanup EXIT

cd "${WORK_DIR}"

log "Downloading ${PACKAGE}"

curl -fL \
    --retry 3 \
    --retry-delay 2 \
    -o "${PACKAGE}" \
    "${RELEASE_URL}/${PACKAGE}"

log "Downloading ${CHECKSUM}"

curl -fL \
    --retry 3 \
    --retry-delay 2 \
    -o "${CHECKSUM}" \
    "${RELEASE_URL}/${CHECKSUM}"

log "Verifying release SHA256"

sha256sum --check "${CHECKSUM}"

log "Validating archive contents"

EXPECTED_CONTENTS="$(
    printf '%s\n' \
        manifest.json \
        sentinel-update.service \
        sentinelctl \
        sentineld |
    sort
)"

ACTUAL_CONTENTS="$(
    tar -tzf "${PACKAGE}" |
    sed 's#^\./##' |
    sort
)"

if [[ "${ACTUAL_CONTENTS}" != "${EXPECTED_CONTENTS}" ]]; then
    printf '%s\n' "Unexpected archive contents:" >&2
    printf '%s\n' "${ACTUAL_CONTENTS}" >&2
    fail "release archive validation failed"
fi

if ! tar -tvzf "${PACKAGE}" |
    awk '
        {
            if (substr($1, 1, 1) != "-")
                exit 1
        }
        END {
            if (NR != 4)
                exit 1
        }
    '
then
    fail "release archive contains unsupported filesystem objects"
fi

mkdir release

tar \
    --no-same-owner \
    --no-same-permissions \
    -xzf "${PACKAGE}" \
    -C release

cd release

[[ -f manifest.json ]] ||
    fail "manifest.json missing"

[[ -f sentineld ]] ||
    fail "sentineld missing"

[[ -f sentinelctl ]] ||
    fail "sentinelctl missing"

[[ -f sentinel-update.service ]] ||
    fail "sentinel-update.service missing"

grep -F "\"version\": \"${VERSION}\"" manifest.json >/dev/null ||
    fail "manifest version mismatch"

grep -F "\"os\": \"linux\"" manifest.json >/dev/null ||
    fail "manifest OS mismatch"

grep -F "\"arch\": \"${ARCH}\"" manifest.json >/dev/null ||
    fail "manifest architecture mismatch"

manifest_sha() {
    local name="$1"

    awk -v target="${name}" '
        $0 ~ "\"name\"" && $0 ~ "\"" target "\"" {
            found=1
            next
        }

        found && $0 ~ "\"sha256\"" {
            gsub(/.*"sha256"[[:space:]]*:[[:space:]]*"/, "")
            gsub(/".*/, "")
            print
            exit
        }
    ' manifest.json
}

verify_manifest_file() {
    local name="$1"
    local expected
    local actual

    expected="$(manifest_sha "${name}")"

    [[ "${expected}" =~ ^[0-9a-fA-F]{64}$ ]] ||
        fail "invalid manifest SHA256 for ${name}"

    actual="$(
        sha256sum "${name}" |
        awk '{print $1}'
    )"

    [[ "${actual,,}" == "${expected,,}" ]] ||
        fail "manifest SHA256 mismatch for ${name}"

    log "Verified ${name}"
}

log "Verifying manifest file hashes"

verify_manifest_file sentineld
verify_manifest_file sentinelctl
verify_manifest_file sentinel-update.service

log "Installing Sentinel binaries"

install \
    -m 0755 \
    sentineld \
    "${INSTALL_DIR}/sentineld"

install \
    -m 0755 \
    sentinelctl \
    "${INSTALL_DIR}/sentinelctl"

log "Installing update worker"

install \
    -m 0644 \
    sentinel-update.service \
    "${UPDATE_SERVICE}"

log "Installing sentineld systemd service"

curl -fsSL \
    "${RAW_BASE}/${VERSION}/systemd/sentineld.service" \
    -o "${SENTINELD_SERVICE}"

chmod 0644 "${SENTINELD_SERVICE}"

mkdir -p \
    "${CONFIG_DIR}" \
    "${STATE_DIR}"

if [[ ! -f "${CONFIG_PATH}" ]]; then

    log "Installing default configuration"

    curl -fsSL \
        "${RAW_BASE}/${VERSION}/configs/sentinel.yaml" \
        -o "${CONFIG_PATH}"

    sed -i \
        's/owner: ""/owner: "FilipooSVK"/' \
        "${CONFIG_PATH}"

    sed -i \
        's/repository: ""/repository: "Kyronix-Sentinel"/' \
        "${CONFIG_PATH}"

else

    log "Existing configuration preserved: ${CONFIG_PATH}"
fi

chmod 0644 "${CONFIG_PATH}"

log "Reloading systemd"

systemctl daemon-reload

log "Enabling and starting sentineld"

systemctl enable sentineld.service >/dev/null

systemctl restart sentineld.service

sleep 2

if ! systemctl is-active --quiet sentineld.service; then

    systemctl status \
        sentineld.service \
        --no-pager || true

    fail "sentineld failed to start"
fi

log "Verifying Sentinel runtime"

"${INSTALL_DIR}/sentinelctl" version
"${INSTALL_DIR}/sentinelctl" status

log "Verifying update worker"

if [[ "$(
    systemctl show \
        sentinel-update.service \
        -p LoadState \
        --value
)" != "loaded" ]]; then

    fail "sentinel-update.service is not loaded"
fi

printf '\n'
printf '%s\n' "Kyronix Sentinel installation completed successfully."
printf '\n'
printf '%s\n' "Useful commands:"
printf '%s\n' "  sentinelctl status"
printf '%s\n' "  sentinelctl prediction"
printf '%s\n' "  sentinelctl update check"
printf '%s\n' "  systemctl status sentineld"
printf '\n'
