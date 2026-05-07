#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: .github/scripts/sync-sub2api.sh [--dry-run] [--run-tests]

Creates a codex/sync-sub2api-YYYYMMDD branch from origin/main, fetches the
Wei-Shaw/sub2api upstream without tags, and prepares a candidate upstream patch.

Options:
  --dry-run    Report the candidate diff without changing files.
  --run-tests  Run the SupplyGate guard and full sync test suite after applying.
EOF
}

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

run() {
  printf '==> %s\n' "$*" >&2
  "$@"
}

dry_run=0
run_tests=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      dry_run=1
      ;;
    --run-tests)
      run_tests=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      fail "unknown option: $1"
      ;;
  esac
  shift
done

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

state_file=".github/upstream-sub2api.env"
[[ -f "$state_file" ]] || fail "missing $state_file"
# shellcheck source=/dev/null
source "$state_file"

: "${UPSTREAM_REMOTE:?missing UPSTREAM_REMOTE}"
: "${UPSTREAM_URL:?missing UPSTREAM_URL}"
: "${UPSTREAM_BRANCH:?missing UPSTREAM_BRANCH}"
: "${UPSTREAM_BASELINE_SHA:?missing UPSTREAM_BASELINE_SHA}"
: "${UPSTREAM_BASELINE_LABEL:?missing UPSTREAM_BASELINE_LABEL}"

if [[ -n "$(git status --porcelain)" ]]; then
  fail "working tree is not clean; commit or stash local changes before syncing"
fi

if git remote get-url "$UPSTREAM_REMOTE" >/dev/null 2>&1; then
  current_url="$(git remote get-url "$UPSTREAM_REMOTE")"
  [[ "$current_url" == "$UPSTREAM_URL" ]] || fail "$UPSTREAM_REMOTE points to $current_url, expected $UPSTREAM_URL"
else
  run git remote add "$UPSTREAM_REMOTE" "$UPSTREAM_URL"
fi
run git remote set-url --push "$UPSTREAM_REMOTE" DISABLED

run git fetch --no-tags origin "+refs/heads/main:refs/remotes/origin/main"
run git fetch --no-tags "$UPSTREAM_REMOTE" "+refs/heads/*:refs/remotes/$UPSTREAM_REMOTE/*" --prune

upstream_ref="refs/remotes/$UPSTREAM_REMOTE/$UPSTREAM_BRANCH"
git rev-parse --verify "$upstream_ref^{commit}" >/dev/null 2>&1 || fail "upstream ref not found: $upstream_ref"
git rev-parse --verify "$UPSTREAM_BASELINE_SHA^{commit}" >/dev/null 2>&1 || fail "baseline commit not found locally: $UPSTREAM_BASELINE_SHA"
git merge-base --is-ancestor "$UPSTREAM_BASELINE_SHA" "$upstream_ref" || fail "$UPSTREAM_BASELINE_LABEL is not an ancestor of $upstream_ref"

upstream_sha="$(git rev-parse "$upstream_ref")"
if [[ "$upstream_sha" == "$UPSTREAM_BASELINE_SHA" ]]; then
  printf 'No upstream changes: %s already matches %s.\n' "$UPSTREAM_BRANCH" "$UPSTREAM_BASELINE_LABEL"
  exit 0
fi

sync_date="${SYNC_DATE:-$(date +%Y%m%d)}"
branch="codex/sync-sub2api-$sync_date"
suffix=2
while git show-ref --verify --quiet "refs/heads/$branch" || git ls-remote --exit-code --heads origin "$branch" >/dev/null 2>&1; do
  branch="codex/sync-sub2api-$sync_date-$suffix"
  suffix=$((suffix + 1))
done

protected_pathspecs=(
  ":(exclude)README.md"
  ":(exclude)CLA.md"
  ":(exclude)backend/cmd/server/VERSION"
  ":(exclude).github/workflows/release.yml"
  ":(exclude).github/workflows/cla.yml"
  ":(exclude)frontend/src/views/HomeView.vue"
  ":(exclude)frontend/src/views/KeyUsageView.vue"
  ":(exclude)backend/ent/schema/supplier_profile.go"
  ":(exclude)backend/ent/supplierprofile.go"
  ":(exclude)backend/ent/supplierprofile/**"
  ":(exclude)backend/ent/supplierprofile_create.go"
  ":(exclude)backend/ent/supplierprofile_delete.go"
  ":(exclude)backend/ent/supplierprofile_query.go"
  ":(exclude)backend/ent/supplierprofile_update.go"
  ":(exclude)backend/internal/handler/admin/supplier_handler.go"
  ":(exclude)backend/internal/handler/auth_supplier_status_test.go"
  ":(exclude)backend/internal/handler/dto/mappers_supplier_pricing_test.go"
  ":(exclude)backend/internal/handler/supplier_handler.go"
  ":(exclude)backend/internal/repository/supplier_pricing_repo.go"
  ":(exclude)backend/internal/repository/supplier_repo.go"
  ":(exclude)backend/internal/server/middleware/supplier_auth.go"
  ":(exclude)backend/internal/server/routes/supplier.go"
  ":(exclude)backend/internal/service/supplier.go"
  ":(exclude)backend/internal/service/supplier_approval_test.go"
  ":(exclude)backend/internal/service/supplier_pricing_validation_test.go"
  ":(exclude)backend/migrations/134_supplier_center_mvp.sql"
  ":(exclude)backend/migrations/135_supplier_account_model_test_gate.sql"
  ":(exclude)backend/migrations/136_supplier_account_pricing_revisions.sql"
  ":(exclude)backend/migrations/137_supplier_account_pricing_revision_kind.sql"
  ":(exclude)frontend/src/api/__tests__/supplierPricing.spec.ts"
  ":(exclude)frontend/src/api/admin/suppliers.ts"
  ":(exclude)frontend/src/api/supplier.ts"
  ":(exclude)frontend/src/views/admin/SuppliersView.vue"
  ":(exclude)frontend/src/views/supplier/**"
)

patch_file="$(mktemp "${TMPDIR:-/tmp}/sub2api-sync.XXXXXX.patch")"
trap 'rm -f "$patch_file"' EXIT

run git diff --binary "$UPSTREAM_BASELINE_SHA" "$upstream_ref" -- . "${protected_pathspecs[@]}" > "$patch_file"

changed_count="$(git diff --name-only "$UPSTREAM_BASELINE_SHA" "$upstream_ref" -- . "${protected_pathspecs[@]}" | sed '/^$/d' | wc -l | tr -d ' ')"
printf 'Upstream candidate: %s..%s (%s files after protected exclusions)\n' "$UPSTREAM_BASELINE_LABEL" "$upstream_sha" "$changed_count"

if [[ "$dry_run" -eq 1 ]]; then
  git diff --name-status "$UPSTREAM_BASELINE_SHA" "$upstream_ref" -- . "${protected_pathspecs[@]}" | sed -n '1,240p'
  printf '\nDry run only. No branch or files were changed.\n'
  exit 0
fi

run git switch -c "$branch" "origin/main"

if [[ ! -s "$patch_file" ]]; then
  printf 'No unprotected upstream changes to apply.\n'
  exit 0
fi

if ! git apply --3way --check "$patch_file"; then
  fail "candidate patch does not apply cleanly; inspect upstream diff and cherry-pick manually on $branch"
fi

run git apply --3way "$patch_file"

tmp_state="$(mktemp)"
awk -v sha="$upstream_sha" -v label="$UPSTREAM_BRANCH@$upstream_sha" '
  /^UPSTREAM_BASELINE_LABEL=/ { print "UPSTREAM_BASELINE_LABEL=" label; next }
  /^UPSTREAM_BASELINE_SHA=/ { print "UPSTREAM_BASELINE_SHA=" sha; next }
  { print }
' "$state_file" > "$tmp_state"
mv "$tmp_state" "$state_file"

run .github/scripts/check-sub2api-sync-guard.sh origin/main

if [[ "$run_tests" -eq 1 ]]; then
  run .github/scripts/run-sub2api-sync-checks.sh
else
  printf '\nNext steps:\n'
  printf '  1. Review the diff: git diff --stat origin/main && git diff origin/main\n'
  printf '  2. Run checks: .github/scripts/run-sub2api-sync-checks.sh\n'
  printf '  3. Commit and open a draft PR only after the diff is acceptable.\n'
fi
