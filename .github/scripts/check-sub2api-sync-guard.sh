#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

base_ref="${1:-origin/main}"
git rev-parse --verify "$base_ref^{commit}" >/dev/null 2>&1 || fail "base ref not found: $base_ref"

protected_paths=(
  "README.md"
  "CLA.md"
  "backend/cmd/server/VERSION"
  ".github/workflows/release.yml"
  ".github/workflows/cla.yml"
  "frontend/src/views/HomeView.vue"
  "frontend/src/views/KeyUsageView.vue"
)

changed="$(
  {
    git diff --name-only "$base_ref"...HEAD -- "${protected_paths[@]}" || true
    git diff --name-only -- "${protected_paths[@]}" || true
    git diff --cached --name-only -- "${protected_paths[@]}" || true
  } | sed '/^$/d' | sort -u
)"

if [[ -n "$changed" ]]; then
  printf 'Protected SupplyGate files changed during upstream sync:\n%s\n' "$changed" >&2
  fail "restore these files from $base_ref or move the change to a non-sync PR"
fi

grep -q '^# SupplyGate$' README.md || fail "README.md no longer identifies this project as SupplyGate"
grep -q 'hlaym129-stack/supplygate' README.md || fail "README.md no longer points to the SupplyGate repository"
grep -q 'SupplyGate' .github/workflows/release.yml || fail "release workflow no longer uses SupplyGate release branding"

printf 'Sub2API sync guard passed against %s.\n' "$base_ref"
