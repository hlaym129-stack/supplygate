ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS supplier_profile_id BIGINT;

ALTER TABLE groups
    DROP CONSTRAINT IF EXISTS fk_groups_supplier_profile;

ALTER TABLE groups
    ADD CONSTRAINT fk_groups_supplier_profile
    FOREIGN KEY (supplier_profile_id)
    REFERENCES supplier_profiles(id)
    ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_groups_supplier_profile_id
    ON groups(supplier_profile_id)
    WHERE supplier_profile_id IS NOT NULL AND deleted_at IS NULL;
