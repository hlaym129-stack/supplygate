ALTER TABLE supplier_profiles
    ADD COLUMN IF NOT EXISTS account_submission_enabled BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE supplier_profiles
SET account_submission_enabled = TRUE
WHERE status = 'approved';

CREATE INDEX IF NOT EXISTS supplier_profiles_account_submission_enabled_idx
    ON supplier_profiles(account_submission_enabled);
