#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

run() {
  printf '\n==> %s\n' "$*"
  "$@"
}

golangci_lint_bin="${GOLANGCI_LINT:-golangci-lint}"
if ! command -v "$golangci_lint_bin" >/dev/null 2>&1; then
  gopath="$(go env GOPATH)"
  if [[ -x "$gopath/bin/golangci-lint" ]]; then
    golangci_lint_bin="$gopath/bin/golangci-lint"
  fi
fi

python_bin="${PYTHON:-python3.11}"
if ! command -v "$python_bin" >/dev/null 2>&1; then
  python_bin="python3"
fi

run make -C backend test-unit
run make -C backend test-integration
printf '\n==> %s\n' "$golangci_lint_bin run --timeout=30m (backend)"
(cd backend && "$golangci_lint_bin" run --timeout=30m)
run pnpm --dir frontend run typecheck
run pnpm --dir frontend run test:run

audit_file="$(mktemp)"
trap 'rm -f "$audit_file"' EXIT
printf '\n==> pnpm --dir frontend audit --prod --audit-level=high --json\n'
pnpm --dir frontend audit --prod --audit-level=high --json > "$audit_file" || true
run "$python_bin" tools/check_pnpm_audit_exceptions.py \
  --audit "$audit_file" \
  --exceptions .github/audit-exceptions.yml
