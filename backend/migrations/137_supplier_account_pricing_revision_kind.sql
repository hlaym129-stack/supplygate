ALTER TABLE supplier_account_pricing_revisions
    ADD COLUMN IF NOT EXISTS revision_kind VARCHAR(20);

UPDATE supplier_account_pricing_revisions
SET revision_kind = COALESCE(NULLIF(revision_kind, ''), 'initial')
WHERE revision_kind IS NULL OR revision_kind = '';

ALTER TABLE supplier_account_pricing_revisions
    ALTER COLUMN revision_kind SET DEFAULT 'initial',
    ALTER COLUMN revision_kind SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_supplier_kind_created
    ON supplier_account_pricing_revisions(supplier_id, revision_kind, created_at DESC);
