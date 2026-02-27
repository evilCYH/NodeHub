#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/web"
GO_CHECK=false

if [[ "${1:-}" == "--with-go" ]]; then
  GO_CHECK=true
fi

info() {
  printf '[INFO] %s\n' "$1"
}

ok() {
  printf '[OK] %s\n' "$1"
}

fail() {
  printf '[ERROR] %s\n' "$1" >&2
  exit 1
}

run_step() {
  local title="$1"
  shift
  info "$title"
  "$@"
  ok "$title"
}

command -v node >/dev/null 2>&1 || fail "Node.js 未安装"
command -v npm >/dev/null 2>&1 || fail "npm 未安装"
[[ -d "$WEB_DIR" ]] || fail "未找到目录: $WEB_DIR"

cd "$WEB_DIR"

if [[ -f pnpm-lock.yaml ]]; then
  PM="pnpm"
  if ! command -v pnpm >/dev/null 2>&1; then
    run_step "安装 pnpm" npm i -g pnpm
  fi
  INSTALL_CMD=(pnpm install)
  LINT_CMD=(pnpm run lint)
  BUILD_CMD=(pnpm run build)
elif [[ -f package-lock.json ]]; then
  PM="npm"
  INSTALL_CMD=(npm ci)
  LINT_CMD=(npm run lint)
  BUILD_CMD=(npm run build)
else
  PM="npm"
  INSTALL_CMD=(npm install)
  LINT_CMD=(npm run lint)
  BUILD_CMD=(npm run build)
fi

info "使用包管理器: $PM"
run_step "安装前端依赖" "${INSTALL_CMD[@]}"
run_step "前端 Lint 检查" "${LINT_CMD[@]}"
run_step "前端构建检查" "${BUILD_CMD[@]}"

if [[ "$GO_CHECK" == "true" ]]; then
  command -v go >/dev/null 2>&1 || fail "未检测到 Go，无法执行 --with-go"
  cd "$ROOT_DIR"
  run_step "下载 Go 依赖" go mod download
  TMP_BIN="${TMPDIR:-/tmp}/nodehub-predeploy-check"
  run_step "后端编译检查" go build -o "$TMP_BIN" ./cmd/nodehub
  rm -f "$TMP_BIN"
  ok "已清理临时编译产物"
fi

ok "部署前检查通过"
