-- +goose Up
-- +goose StatementBegin
CREATE TABLE checks (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(127) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    has_alert BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_checks_user_id ON checks(user_id);
CREATE INDEX idx_checks_has_alert ON checks(has_alert);
CREATE INDEX idx_checks_created_at ON checks(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE checks;
-- +goose StatementEnd
