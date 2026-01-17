include .env
export

DB_URL = postgres://$(PG_DB_USER):$(PG_DB_PASSWORD)@$(PG_DB_HOST):$(PG_DB_PORT)/$(PG_DB_NAME)?sslmode=disable

.PHONY: run build migrate-up migrate-down migrate-create clean

run:
	go run cmd/main.go

build:
	go build -o bin/geonotify cmd/main.go

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

docs : clean
	swag init -g ./cmd/main.go --output ./docs --parseDependency --parseInternal

clean : 
	rm -rf docs/