ALTER TABLE database_instances DROP CONSTRAINT database_instances_agent_id_fkey;

ALTER TABLE agents ALTER COLUMN id DROP DEFAULT;
ALTER TABLE agents ALTER COLUMN id TYPE TEXT USING id::TEXT;

ALTER TABLE database_instances ALTER COLUMN id DROP DEFAULT;
ALTER TABLE database_instances ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE database_instances ALTER COLUMN agent_id TYPE TEXT USING agent_id::TEXT;

ALTER TABLE database_instances
    ADD CONSTRAINT database_instances_agent_id_fkey
    FOREIGN KEY (agent_id) REFERENCES agents (id);

ALTER TABLE metric_points ALTER COLUMN agent_id TYPE TEXT USING agent_id::TEXT;
ALTER TABLE metric_points ALTER COLUMN database_instance_id TYPE TEXT USING database_instance_id::TEXT;

ALTER TABLE findings ALTER COLUMN agent_id TYPE TEXT USING agent_id::TEXT;
ALTER TABLE findings ALTER COLUMN database_instance_id TYPE TEXT USING database_instance_id::TEXT;
