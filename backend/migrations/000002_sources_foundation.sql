-- Add a first-class source model so Postgresome is not structurally bound
-- to agent-only ingestion. Existing agent-based databases become one source
-- kind among several future connector types.

CREATE TABLE IF NOT EXISTS sources (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    provider TEXT NOT NULL,
    name TEXT NOT NULL,
    agent_id TEXT REFERENCES agents (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sources_kind_check
        CHECK (kind IN ('agent', 'direct', 'managed_integration')),
    CONSTRAINT sources_provider_check
        CHECK (provider IN ('postgres', 'supabase', 'rds', 'cloudsql', 'neon'))
);

ALTER TABLE database_instances
    ADD COLUMN IF NOT EXISTS source_id TEXT;

UPDATE database_instances
SET source_id = id
WHERE source_id IS NULL;

INSERT INTO sources (id, kind, provider, name, agent_id)
SELECT
    di.id,
    'agent',
    'postgres',
    di.name,
    di.agent_id
FROM database_instances di
LEFT JOIN sources s ON s.id = di.id
WHERE s.id IS NULL;

ALTER TABLE database_instances
    ALTER COLUMN source_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE constraint_name = 'database_instances_source_id_fkey'
          AND table_name = 'database_instances'
    ) THEN
        ALTER TABLE database_instances
            ADD CONSTRAINT database_instances_source_id_fkey
            FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_database_instances_source_id
    ON database_instances (source_id);
