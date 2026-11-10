-- migrate:up
ALTER TABLE message_templates ADD COLUMN "order" int4 NOT NULL DEFAULT 1000;


-- migrate:down
ALTER TABLE message_templates DROP COLUMN "order";
