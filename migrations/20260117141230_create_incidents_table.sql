-- +goose Up
-- +goose StatementBegin
CREATE TABLE incidents (
    id SERIAL PRIMARY KEY,
    name VARCHAR(127) NOT NULL,
    descr TEXT,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius_m DOUBLE PRECISION NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_incidents_is_active ON incidents(is_active);
CREATE INDEX idx_incidents_created_at ON incidents(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE incidents;
-- +goose StatementEnd
