-- Allow database instances to exist without an agent so direct and managed
-- integrations can share the same backend diagnosis model.

ALTER TABLE database_instances
    ALTER COLUMN agent_id DROP NOT NULL;
