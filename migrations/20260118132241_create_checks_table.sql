-- +goose Up
-- +goose StatementBegin
CREATE TABLE checks (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(127) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    has_alert BOLEEAN NOT NULL,
    created_at TIMESTAMP DEFAULT NOT()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE checks;
-- +goose StatementEnd
