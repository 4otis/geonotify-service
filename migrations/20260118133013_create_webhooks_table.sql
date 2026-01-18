-- +goose Up
-- +goose StatementBegin
CREATE TABLE webhooks (
    id SERIAL PRIMARY KEY,
    check_id INTEGER REFERENCES checks(id) ON DELETE CASCADE,
    state VARCHAR(50) DEFAULT 'in progress',
    retry_cnt INTEGER NOT NULL DEFAULT 0,
    payload BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    scheduled_at TIMESTAMP DEFAULT NOW(),
);

CREATE INDEX idx_webhooks_status ON webhooks(status);
CREATE INDEX idx_webhooks_scheduled_at ON webhooks(scheduled_at);
CREATE INDEX idx_webhooks_check_id ON webhooks(check_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE webhooks;
-- +goose StatementEnd
