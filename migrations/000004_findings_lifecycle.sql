-- Findings lifecycle: turn findings into stateful open/resolved issues with
-- fingerprint-based deduplication instead of an append-only event log.

ALTER TABLE findings
    ADD COLUMN rule_key TEXT NOT NULL DEFAULT '',
    ADD COLUMN resource_type TEXT NOT NULL DEFAULT '',
    ADD COLUMN resource_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN fingerprint TEXT,
    ADD COLUMN current_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN threshold_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN status TEXT NOT NULL DEFAULT 'open',
    ADD COLUMN occurrence_count INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN first_seen_at TIMESTAMPTZ,
    ADD COLUMN last_seen_at TIMESTAMPTZ,
    ADD COLUMN resolved_at TIMESTAMPTZ;

-- Backfill existing rows with a unique-per-row fingerprint (their id) so the
-- new unique constraint can be added safely. These rows won't match any
-- fingerprint a rule computes going forward, so they'll be auto-resolved by
-- the next resolve sweep.
UPDATE findings
SET fingerprint = id::text,
    first_seen_at = created_at,
    last_seen_at = created_at
WHERE fingerprint IS NULL;

ALTER TABLE findings
    ALTER COLUMN fingerprint SET NOT NULL,
    ALTER COLUMN first_seen_at SET NOT NULL,
    ALTER COLUMN last_seen_at SET NOT NULL;

ALTER TABLE findings
    ADD CONSTRAINT findings_status_check CHECK (status IN ('open', 'resolved'));

-- Global (not partial) uniqueness per database instance: when a resolved
-- finding's fingerprint reappears, ON CONFLICT reopens the same row instead
-- of inserting a duplicate.
ALTER TABLE findings
    ADD CONSTRAINT findings_database_instance_fingerprint_key UNIQUE (database_instance_id, fingerprint);

CREATE INDEX IF NOT EXISTS idx_findings_status ON findings (database_instance_id, status);
CREATE INDEX IF NOT EXISTS idx_findings_last_seen_at ON findings (database_instance_id, status, last_seen_at);
