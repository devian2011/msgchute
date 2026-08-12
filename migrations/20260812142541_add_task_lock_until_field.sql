-- migrate:up
ALTER TABLE tasks ADD COLUMN lock_until TIMESTAMP WITH TIME ZONE DEFAULT NULL;
CREATE INDEX idx_tasks_lock_until ON tasks (lock_until) WHERE lock_until IS NOT NULL;

UPDATE tasks SET lock_until='0001-01-01 00:00:00.000000 +00:00';

-- migrate:down
DROP INDEX IF EXISTS idx_tasks_lock_until;
ALTER TABLE tasks DROP COLUMN IF EXISTS lock_until;
