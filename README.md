# Architecture TODO Backend (MVP Skeleton)

Учебный backend на Go в стиле модульного монолита с попыткой в чистую архитектуру с DDD

В этой версии добавлено:

- Go HTTP API
- PostgreSQL
- Docker / docker-compose

## Структура проекта

```text
backend/
  cmd/server
  internal/
    bootstrap
    config
    modules/
      identity/
      todo/
    platform/database
    shared
docker/
  docker-compose.yml
  docker-compose.debug.yml
  docker-compose.prod.yml
```

## Быстрый старт

1. Запуск:

```bash
docker compose -f docker/docker-compose.yml -f docker/docker-compose.debug.yml up --build
```

Прод-режим:

```bash
docker compose -f docker/docker-compose.yml -f docker/docker-compose.prod.yml up --build
```

Или через Makefile:

```bash
make up-debug
```

2. Healthcheck:

```bash
curl http://localhost:8080/healthz
```

## Команды (оформлены через Makefile)

```bash
make up-debug      # запуск debug в фоне с пересборкой
make down-debug    # остановка debug окружения
make logs-debug    # логи debug окружения
make config-debug  # проверка debug compose-конфига

make up-prod       # запуск prod в фоне с пересборкой
make down-prod     # остановка prod окружения
make logs-prod     # логи prod окружения
make config-prod   # проверка prod compose-конфига

make test-api      # запуск python e2e тестов
```

## Файлы окружения

- backend/.env.debug
- backend/.env.prod

## API (текущий swagger)

### Регистрация

```bash
curl -X POST http://localhost:8080/api/v1/identity/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
```

Ответ содержит `user_id`. Для TODO-эндпоинтов передавайте его в заголовке `X-User-ID`.

### Создание TODO

```bash
curl -X POST http://localhost:8080/api/v1/todos/ \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user_id>" \
  -d '{"title":"Learn DDD","description":"Read aggregate chapter","priority":3}'
```

### Список TODO

```bash
curl -X GET http://localhost:8080/api/v1/todos/ \
  -H "X-User-ID: <user_id>"
```

### Получить инфо по конкретной TODO

```bash
curl -X GET http://localhost:8080/api/v1/todos/<todo_id> \
  -H "X-User-ID: <user_id>"
```

### Изменить статус TODO

```bash
curl -X PATCH http://localhost:8080/api/v1/todos/<todo_id>/status \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user_id>" \
  -d '{"status":"done"}'
```

### Изменить приоритет TODO

```bash
curl -X PATCH http://localhost:8080/api/v1/todos/<todo_id>/priority \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user_id>" \
  -d '{"priority":5}'
```

### Удаление TODO

```bash
curl -X DELETE http://localhost:8080/api/v1/todos/<todo_id> \
  -H "X-User-ID: <user_id>"
```