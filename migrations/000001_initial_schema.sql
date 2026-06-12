CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    environment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS database_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents (id),
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS metric_points (
    time TIMESTAMPTZ NOT NULL,
    metric_key TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit TEXT,
    database_instance_id UUID,
    agent_id UUID,
    dimensions JSONB
);

SELECT create_hypertable('metric_points', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metric_points_metric_key_time ON metric_points (metric_key, time);
CREATE INDEX IF NOT EXISTS idx_metric_points_database_instance_id_time ON metric_points (database_instance_id, time);
CREATE INDEX IF NOT EXISTS idx_metric_points_agent_id_time ON metric_points (agent_id, time);
CREATE INDEX IF NOT EXISTS idx_metric_points_dimensions ON metric_points USING GIN (dimensions);

CREATE TABLE IF NOT EXISTS findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    severity TEXT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    recommendation TEXT NOT NULL,
    database_instance_id UUID,
    agent_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
