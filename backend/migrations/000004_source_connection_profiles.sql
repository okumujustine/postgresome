CREATE TABLE IF NOT EXISTS source_connection_profiles (
    source_id TEXT PRIMARY KEY REFERENCES sources (id) ON DELETE CASCADE,
    connection_uri TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE sources
    ADD COLUMN IF NOT EXISTS last_check_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_check_completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_check_status TEXT NOT NULL DEFAULT 'not_run',
    ADD COLUMN IF NOT EXISTS last_check_error TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE constraint_name = 'sources_last_check_status_check'
          AND table_name = 'sources'
    ) THEN
        ALTER TABLE sources
            ADD CONSTRAINT sources_last_check_status_check
            CHECK (last_check_status IN ('not_run', 'running', 'succeeded', 'failed'));
    END IF;
END $$;
