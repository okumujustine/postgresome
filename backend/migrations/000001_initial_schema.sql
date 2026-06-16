-- Postgresome baseline schema.
-- This is intentionally squashed because the project is still pre-deploy and
-- we want the migration history to reflect the diagnosis-first product model
-- rather than every intermediate refactor.

CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    environment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS database_instances (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES agents (id),
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS metric_points (
    time TIMESTAMPTZ NOT NULL,
    metric_key TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit TEXT,
    database_instance_id TEXT,
    agent_id TEXT,
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb
);

SELECT create_hypertable('metric_points', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metric_points_metric_key_time
    ON metric_points (metric_key, time);

CREATE INDEX IF NOT EXISTS idx_metric_points_database_instance_id_time
    ON metric_points (database_instance_id, time);

CREATE INDEX IF NOT EXISTS idx_metric_points_agent_id_time
    ON metric_points (agent_id, time);

CREATE INDEX IF NOT EXISTS idx_metric_points_dimensions
    ON metric_points USING GIN (dimensions);

CREATE TABLE IF NOT EXISTS findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    severity TEXT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    recommendation TEXT NOT NULL,
    database_instance_id TEXT REFERENCES database_instances (id) ON DELETE CASCADE,
    agent_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rule_key TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL DEFAULT '',
    resource_name TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL,
    current_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    threshold_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'open',
    occurrence_count INTEGER NOT NULL DEFAULT 1,
    regression_count INTEGER NOT NULL DEFAULT 0,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    last_regressed_at TIMESTAMPTZ,
    improving_since TIMESTAMPTZ,
    verified_fixed_at TIMESTAMPTZ,
    verification_status TEXT NOT NULL DEFAULT 'pending',
    verification_summary TEXT NOT NULL DEFAULT '',
    problem_summary TEXT NOT NULL DEFAULT '',
    evidence_summary TEXT NOT NULL DEFAULT '',
    impact_summary TEXT NOT NULL DEFAULT '',
    suggested_action TEXT NOT NULL DEFAULT '',
    confidence_label TEXT NOT NULL DEFAULT '',
    confidence_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    baseline_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    baseline_label TEXT NOT NULL DEFAULT '',
    change_summary TEXT NOT NULL DEFAULT '',
    verification_hint TEXT NOT NULL DEFAULT '',
    CONSTRAINT findings_status_check
        CHECK (status IN ('open', 'resolved')),
    CONSTRAINT findings_verification_status_check
        CHECK (verification_status IN ('pending', 'regressed', 'improving', 'verified_fixed')),
    CONSTRAINT findings_database_instance_fingerprint_key
        UNIQUE (database_instance_id, fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_findings_status
    ON findings (database_instance_id, status);

CREATE INDEX IF NOT EXISTS idx_findings_last_seen_at
    ON findings (database_instance_id, status, last_seen_at);

CREATE TABLE IF NOT EXISTS table_stats (
    id BIGSERIAL PRIMARY KEY,
    database_instance_id TEXT NOT NULL REFERENCES database_instances (id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    live_rows BIGINT NOT NULL DEFAULT 0,
    dead_rows BIGINT NOT NULL DEFAULT 0,
    sequential_scans BIGINT NOT NULL DEFAULT 0,
    sequential_rows_read BIGINT NOT NULL DEFAULT 0,
    index_scans BIGINT NOT NULL DEFAULT 0,
    index_rows_fetched BIGINT NOT NULL DEFAULT 0,
    rows_inserted BIGINT NOT NULL DEFAULT 0,
    rows_updated BIGINT NOT NULL DEFAULT 0,
    rows_deleted BIGINT NOT NULL DEFAULT 0,
    last_vacuum_at TIMESTAMPTZ,
    last_autovacuum_at TIMESTAMPTZ,
    last_analyze_at TIMESTAMPTZ,
    last_autoanalyze_at TIMESTAMPTZ,
    vacuum_count BIGINT NOT NULL DEFAULT 0,
    autovacuum_count BIGINT NOT NULL DEFAULT 0,
    analyze_count BIGINT NOT NULL DEFAULT 0,
    autoanalyze_count BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT idx_table_stats_snapshot_row
        UNIQUE (database_instance_id, collected_at, schema_name, table_name)
);

CREATE INDEX IF NOT EXISTS idx_table_stats_instance
    ON table_stats (database_instance_id);

CREATE INDEX IF NOT EXISTS idx_table_stats_instance_collected_at
    ON table_stats (database_instance_id, collected_at DESC);

CREATE INDEX IF NOT EXISTS idx_table_stats_resource_history
    ON table_stats (database_instance_id, schema_name, table_name, collected_at DESC);

CREATE TABLE IF NOT EXISTS query_stats (
    id BIGSERIAL PRIMARY KEY,
    database_instance_id TEXT NOT NULL REFERENCES database_instances (id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL,
    query_id TEXT NOT NULL,
    database_name TEXT NOT NULL,
    user_name TEXT NOT NULL,
    query TEXT NOT NULL,
    calls BIGINT NOT NULL DEFAULT 0,
    total_exec_time_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    mean_exec_time_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    min_exec_time_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_exec_time_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    rows_returned BIGINT NOT NULL DEFAULT 0,
    shared_blocks_hit BIGINT NOT NULL DEFAULT 0,
    shared_blocks_read BIGINT NOT NULL DEFAULT 0,
    shared_blocks_dirtied BIGINT NOT NULL DEFAULT 0,
    shared_blocks_written BIGINT NOT NULL DEFAULT 0,
    temp_blocks_read BIGINT NOT NULL DEFAULT 0,
    temp_blocks_written BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT idx_query_stats_snapshot_row
        UNIQUE (database_instance_id, collected_at, query_id)
);

CREATE INDEX IF NOT EXISTS idx_query_stats_instance
    ON query_stats (database_instance_id);

CREATE INDEX IF NOT EXISTS idx_query_stats_instance_collected_at
    ON query_stats (database_instance_id, collected_at DESC);

CREATE INDEX IF NOT EXISTS idx_query_stats_resource_history
    ON query_stats (database_instance_id, query_id, collected_at DESC);

CREATE TABLE IF NOT EXISTS activity_stats (
    id BIGSERIAL PRIMARY KEY,
    database_instance_id TEXT NOT NULL REFERENCES database_instances (id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL,
    database_name TEXT NOT NULL,
    process_id BIGINT NOT NULL,
    user_name TEXT NOT NULL,
    application_name TEXT NOT NULL,
    client_address TEXT NOT NULL,
    state TEXT NOT NULL,
    query TEXT NOT NULL,
    wait_event_type TEXT NOT NULL,
    wait_event TEXT NOT NULL,
    backend_started_at TIMESTAMPTZ NOT NULL,
    query_started_at TIMESTAMPTZ,
    state_changed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_activity_stats_instance
    ON activity_stats (database_instance_id);

CREATE TABLE IF NOT EXISTS explain_plans (
    id BIGSERIAL PRIMARY KEY,
    database_instance_id TEXT NOT NULL REFERENCES database_instances (id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL,
    query_id TEXT NOT NULL,
    query TEXT NOT NULL,
    plan_json JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_explain_plans_instance
    ON explain_plans (database_instance_id);

CREATE TABLE IF NOT EXISTS finding_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    finding_id UUID NOT NULL REFERENCES findings (id) ON DELETE CASCADE,
    observed_at TIMESTAMPTZ NOT NULL,
    evidence_type TEXT NOT NULL,
    role TEXT NOT NULL,
    label TEXT NOT NULL,
    summary TEXT NOT NULL,
    metric_key TEXT NOT NULL DEFAULT '',
    current_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    baseline_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    change_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    confidence_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    reference_id TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_finding_evidence_finding_id
    ON finding_evidence (finding_id, position, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_finding_evidence_reference
    ON finding_evidence (evidence_type, reference_id);
