#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# Arrow deploy script
# Usage:
#   ./deploy.sh            — Docker Compose deploy (recommended)
#   ./deploy.sh --native   — Build & run natively (requires Go + Node on host)
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

info()  { echo -e "${CYAN}▶ $*${NC}"; }
ok()    { echo -e "${GREEN}✔ $*${NC}"; }
warn()  { echo -e "${YELLOW}⚠ $*${NC}"; }
die()   { echo -e "${RED}✖ $*${NC}" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

MODE="docker"
[[ "${1:-}" == "--native" ]] && MODE="native"

# ─────────────────────────────────────────────────────────────────────────────
# Shared: create .env if missing
# ─────────────────────────────────────────────────────────────────────────────
setup_env() {
  if [[ ! -f .env ]]; then
    cp .env.example .env
    warn ".env not found — created from .env.example. Edit it if needed, then re-run."
    echo
    cat .env
    echo
  else
    ok ".env already exists"
  fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Docker Compose deploy
# ─────────────────────────────────────────────────────────────────────────────
deploy_docker() {
  info "Checking Docker..."
  command -v docker >/dev/null 2>&1 || die "Docker not found. Install it from https://docs.docker.com/get-docker/"
  docker compose version >/dev/null 2>&1 || die "'docker compose' not available. Upgrade Docker to v2.x or later."
  ok "Docker OK"

  setup_env

  info "Building and starting containers..."
  docker compose up --build -d

  echo
  ok "Arrow is running!"
  PORT="$(grep -E '^PORT=' .env | cut -d= -f2 | tr -d '[:space:]')"
  PORT="${PORT:-8080}"
  echo -e "  ${CYAN}http://$(hostname -I | awk '{print $1}'):${PORT}${NC}"
  echo
  echo "Useful commands:"
  echo "  docker compose logs -f        — live logs"
  echo "  docker compose restart arrow  — restart the app"
  echo "  docker compose down           — stop everything"
}

# ─────────────────────────────────────────────────────────────────────────────
# Native deploy (no Docker)
# ─────────────────────────────────────────────────────────────────────────────
deploy_native() {
  # ── Dependency checks ──────────────────────────────────────────────────────
  info "Checking dependencies..."

  command -v go >/dev/null 2>&1 || die "Go not found. Install from https://go.dev/dl/"
  GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
  ok "Go $GO_VER"

  command -v node >/dev/null 2>&1 || die "Node.js not found. Install from https://nodejs.org"
  ok "Node $(node --version)"

  command -v npm >/dev/null 2>&1 || die "npm not found"
  ok "npm $(npm --version)"

  # Check MongoDB is reachable (warn, don't fail — URI may be remote)
  MONGO_URI="${MONGODB_URI:-mongodb://localhost:27017}"
  if command -v mongosh >/dev/null 2>&1; then
    mongosh --quiet --eval "db.runCommand({ping:1})" "$MONGO_URI" >/dev/null 2>&1 \
      && ok "MongoDB reachable at $MONGO_URI" \
      || warn "MongoDB not reachable at $MONGO_URI — start it before running Arrow"
  else
    warn "mongosh not found — skipping MongoDB connectivity check"
  fi

  setup_env
  # Load .env
  set -a; source .env; set +a

  # ── Build frontend ─────────────────────────────────────────────────────────
  info "Building frontend..."
  (cd ui && npm ci --silent && npm run build)
  ok "Frontend built → ui/dist/"

  # ── Build Go binary ────────────────────────────────────────────────────────
  info "Building Go server..."
  CGO_ENABLED=0 go build -ldflags="-s -w" -o arrow ./cmd/server
  ok "Binary built → ./arrow"

  # ── Install as systemd service (optional) ─────────────────────────────────
  INSTALL_DIR="$(pwd)"
  SERVICE_FILE="/etc/systemd/system/arrow.service"

  if command -v systemctl >/dev/null 2>&1 && [[ $EUID -eq 0 ]]; then
    info "Installing systemd service..."
    PORT="${PORT:-8080}"
    MONGODB_URI="${MONGODB_URI:-mongodb://localhost:27017}"
    MONGODB_DB="${MONGODB_DB:-alerts_db}"
    CORS_ORIGIN="${CORS_ORIGIN:-}"

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Arrow Alert Server
After=network.target mongod.service

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/arrow
Restart=on-failure
RestartSec=5

Environment=ADDR_INFO=0.0.0.0:${PORT}
Environment=MONGODB_URI=${MONGODB_URI}
Environment=MONGODB_DB=${MONGODB_DB}
Environment=CORS_ORIGIN=${CORS_ORIGIN}

# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=${INSTALL_DIR}

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable arrow
    systemctl restart arrow
    ok "systemd service installed and started"

    echo
    ok "Arrow is running as a system service!"
    echo -e "  ${CYAN}http://$(hostname -I | awk '{print $1}'):${PORT}${NC}"
    echo
    echo "Useful commands:"
    echo "  systemctl status arrow   — service status"
    echo "  journalctl -u arrow -f   — live logs"
    echo "  systemctl restart arrow  — restart"
    echo "  systemctl stop arrow     — stop"

  elif command -v systemctl >/dev/null 2>&1; then
    warn "Not running as root — skipping systemd install."
    warn "Re-run with sudo to install as a service, or start manually:"
    echo
    run_foreground
  else
    run_foreground
  fi
}

run_foreground() {
  PORT="${PORT:-8080}"
  info "Starting Arrow on port ${PORT}..."
  echo "(Press Ctrl+C to stop)"
  echo
  ADDR_INFO="0.0.0.0:${PORT}" \
  MONGODB_URI="${MONGODB_URI:-mongodb://localhost:27017}" \
  MONGODB_DB="${MONGODB_DB:-alerts_db}" \
  CORS_ORIGIN="${CORS_ORIGIN:-}" \
    ./arrow
}

# ─────────────────────────────────────────────────────────────────────────────
echo
echo -e "${CYAN}Arrow Deploy Script${NC}"
echo "────────────────────"
echo "Mode: $MODE"
echo

if [[ "$MODE" == "docker" ]]; then
  deploy_docker
else
  deploy_native
fi
