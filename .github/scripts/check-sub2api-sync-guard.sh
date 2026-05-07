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
  "backend/ent/schema/supplier_profile.go"
  "backend/ent/supplierprofile.go"
  "backend/ent/supplierprofile"
  "backend/ent/supplierprofile_create.go"
  "backend/ent/supplierprofile_delete.go"
  "backend/ent/supplierprofile_query.go"
  "backend/ent/supplierprofile_update.go"
  "backend/internal/handler/admin/supplier_handler.go"
  "backend/internal/handler/auth_supplier_status_test.go"
  "backend/internal/handler/dto/mappers_supplier_pricing_test.go"
  "backend/internal/handler/supplier_handler.go"
  "backend/internal/repository/supplier_pricing_repo.go"
  "backend/internal/repository/supplier_repo.go"
  "backend/internal/server/middleware/supplier_auth.go"
  "backend/internal/server/routes/supplier.go"
  "backend/internal/service/supplier.go"
  "backend/internal/service/supplier_approval_test.go"
  "backend/internal/service/supplier_pricing_validation_test.go"
  "backend/migrations/134_supplier_center_mvp.sql"
  "backend/migrations/135_supplier_account_model_test_gate.sql"
  "backend/migrations/136_supplier_account_pricing_revisions.sql"
  "backend/migrations/137_supplier_account_pricing_revision_kind.sql"
  "frontend/src/api/__tests__/supplierPricing.spec.ts"
  "frontend/src/api/admin/suppliers.ts"
  "frontend/src/api/supplier.ts"
  "frontend/src/views/admin/SuppliersView.vue"
  "frontend/src/views/supplier"
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
grep -q 'path: .*/supplier' frontend/src/router/index.ts || fail "supplier routes are missing from frontend router"
grep -q 'supplier_profiles' backend/migrations/134_supplier_center_mvp.sql || fail "supplier center migration is missing"

printf 'Sub2API sync guard passed against %s.\n' "$base_ref"
