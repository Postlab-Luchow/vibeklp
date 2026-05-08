#!/usr/bin/env bash
# Deploy vibeklp to the home server.
#
# - Cross-compiles cmd/server for linux/amd64
# - Stages binary + web/ + data/ + systemd unit on the remote
# - Installs everything under /opt/vibeklp owned by a system user `vibeklp`
# - (Re)starts the systemd service
#
# NOTE: this WILL overwrite /opt/vibeklp/data with the local data/ snapshot
# (rsync --delete). If the server has freshly-crawled data you want to keep,
# update the local data/ first or remove `data` from the rsync source list.

set -euo pipefail

# Resolve repo + deploy dir from this script's location
DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$DEPLOY_DIR/.." && pwd)"
cd "$REPO_DIR"

# Optional local config (gitignored). Put per-host settings here, e.g.:
#   REMOTE=user@your-host
#   INSTALL_DIR=/opt/vibeklp
if [[ -f "$DEPLOY_DIR/deploy.env" ]]; then
    # shellcheck disable=SC1091
    source "$DEPLOY_DIR/deploy.env"
fi

REMOTE="${REMOTE:?REMOTE not set. Export it or create deploy/deploy.env with REMOTE=user@host}"
INSTALL_DIR="${INSTALL_DIR:-/opt/vibeklp}"
SERVICE_NAME="${SERVICE_NAME:-vibeklp}"
STAGING_DIR="${STAGING_DIR:-/tmp/vibeklp-deploy}"

echo "==> [1/4] Cross-compiling server (linux/amd64)"
mkdir -p build
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o build/server ./cmd/server

echo "==> [2/4] Syncing files to ${REMOTE}:${STAGING_DIR}"
ssh "$REMOTE" "rm -rf ${STAGING_DIR} && mkdir -p ${STAGING_DIR}"
rsync -az --delete \
    build/server \
    web \
    data \
    deploy/${SERVICE_NAME}.service \
    "${REMOTE}:${STAGING_DIR}/"

echo "==> [3/4] Installing on remote (you will be prompted for sudo)"
ssh -t "$REMOTE" sudo INSTALL_DIR="$INSTALL_DIR" SERVICE_NAME="$SERVICE_NAME" \
    STAGING_DIR="$STAGING_DIR" bash -s <<'REMOTE_SCRIPT'
set -euo pipefail

# Create dedicated system user/group on first deploy
if ! id vibeklp >/dev/null 2>&1; then
    echo "    creating system user 'vibeklp'"
    useradd --system --home-dir "$INSTALL_DIR" --shell /usr/sbin/nologin vibeklp
fi

# Sync staged files into the install dir (excluding the service unit)
mkdir -p "$INSTALL_DIR"
rsync -a --delete \
    --exclude="${SERVICE_NAME}.service" \
    "${STAGING_DIR}/" "${INSTALL_DIR}/"
chown -R vibeklp:vibeklp "$INSTALL_DIR"
chmod +x "${INSTALL_DIR}/server"

# Install / refresh the systemd unit
install -m 0644 \
    "${STAGING_DIR}/${SERVICE_NAME}.service" \
    "/etc/systemd/system/${SERVICE_NAME}.service"

systemctl daemon-reload
systemctl enable "$SERVICE_NAME" >/dev/null
systemctl restart "$SERVICE_NAME"
systemctl --no-pager --lines=15 status "$SERVICE_NAME" || true

rm -rf "$STAGING_DIR"
REMOTE_SCRIPT

echo "==> [4/4] Done. Service '${SERVICE_NAME}' restarted on ${REMOTE}."
