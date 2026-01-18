
# geonotify-service

Сервис для рыссылки геооповещений, асинхронно интегрированный с новостным порталом.

## Quick Start
- скопировать содержание файла `.env.example` в `.env`:
```sh
mv .env.example .env
```
- для локального запуска Web-приложения необходимо поднять контейнер с сервисами:
```sh
make dev
```
- накатить миграции:
```sh
make migrate-up
```
- собрать swagger-документацию:
```sh
make migrate-up
```
- запустить приложение:
```sh
make run
```

## Testing

Тесты можно запустить, импортировав `/tests/postman_collection.json` в `Postman GUI` (на большее не хватило времени)

## API SPEC

Детально с API сервиса можно ознакомиться, обратившись к `swagger-документации`: http://localhost:8081/swagger/index.html#/

## Enviroment
```txt
LOG_LEVEL=debug

HTTP_PORT=8081

PG_DB_USER=postgres
PG_DB_PASSWORD=password
PG_DB_HOST=localhost
PG_DB_NAME=geonotify_db
PG_DB_PORT=5434

REDIS_URL=redis://localhost:6379/0

SECRET_API_KEY=secret-api-key-required

STATS_TIME_WINDOWS_MINUTES=30
CACHE_TTL_MINUTES=10

WEBHOOK_URL=http://localhost:9090/webhook
WEBHOOK_MAX_RETRIES=3
WEBHOOK_RETRY_DELAY_SECONDS=60
```