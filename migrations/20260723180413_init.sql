-- migrate:up
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status_enum') THEN
            CREATE TYPE task_status_enum AS ENUM ('pending', 'success', 'failure');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS message_templates
(
    code        VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT          DEFAULT NULL,
    params      JSONB         DEFAULT NULL,
    subject     VARCHAR(1024) DEFAULT NULL,
    body        TEXT         NOT NULL
);

CREATE TABLE IF NOT EXISTS messages
(
    id            UUID PRIMARY KEY,
    sender_id     VARCHAR(255)  NOT NULL,
    transport     VARCHAR(255)  NOT NULL,
    status        VARCHAR(255)  NOT NULL,
    template_code VARCHAR(255)             DEFAULT NULL REFERENCES message_templates (code) ON DELETE SET NULL,
    recipients    JSONB         NOT NULL   DEFAULT '[]'::JSONB,
    meta          JSONB                    DEFAULT NULL,
    params        JSONB                    DEFAULT NULL,
    retry         JSONB                    DEFAULT NULL,
    schedule      TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    deadline      TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    subject       VARCHAR(1024) NOT NULL,
    body          TEXT          NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_template_code ON messages (template_code);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages (status);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_transport ON messages(transport);
CREATE INDEX IF NOT EXISTS idx_messages_recipients_gin ON messages USING GIN (recipients);


CREATE TABLE IF NOT EXISTS tasks
(
    id             UUID PRIMARY KEY,
    message_id     UUID                              DEFAULT NULL REFERENCES messages (id) ON DELETE CASCADE,
    worker         VARCHAR(255)             NOT NULL,

    status         task_status_enum         NOT NULL DEFAULT 'pending',

    is_processed boolean NOT NULL DEFAULT false,

    retries        INT                      NOT NULL DEFAULT 0,
    max_retries    INT                      NOT NULL DEFAULT 3,
    backoff_code   VARCHAR(255)             NOT NULL,
    backoff_params JSONB                    NOT NULL DEFAULT '{}'::JSONB,
    deadline       TIMESTAMP WITH TIME ZONE          DEFAULT NULL,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_run       TIMESTAMP WITH TIME ZONE          DEFAULT NULL,
    next_run       TIMESTAMP WITH TIME ZONE          DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_next_run_id_pending
    ON tasks (next_run, id)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_tasks_message_id ON tasks (message_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);

CREATE TABLE IF NOT EXISTS task_execution_results
(
    id             UUID PRIMARY KEY,
    task_id        UUID                     NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,

    status         task_status_enum         NOT NULL,

    run_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NULL,
    result         BYTEA                             DEFAULT NULL,
    is_critical    BOOLEAN                  NOT NULL DEFAULT FALSE,
    execution_time INTERVAL                 NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_task_execution_results_task_id ON task_execution_results (task_id);
CREATE INDEX IF NOT EXISTS idx_task_execution_results_status ON task_execution_results (status);


-- migrate:down

DROP INDEX IF EXISTS idx_task_execution_results_status;
DROP INDEX IF EXISTS idx_task_execution_results_task_id;
DROP TABLE IF EXISTS task_execution_results;

DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_message_id;
DROP INDEX IF EXISTS idx_tasks_next_run_id_pending;
DROP TABLE IF EXISTS tasks;

DROP INDEX IF EXISTS idx_messages_status;
DROP INDEX IF EXISTS idx_messages_template_code;
DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS message_templates;
DROP TYPE IF EXISTS task_status_enum;
