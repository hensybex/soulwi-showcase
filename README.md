# Soulwi Backend Showcase (Go)

Публичный showcase Go-бэкенда проекта Soulwi с фокусом на инженерное качество: архитектура, модульность, читаемость и поддерживаемость.

## Что здесь важно

- Слоистая архитектура: `handler -> usecase -> repository`.
- Явное разделение ответственности по пакетам и фичам.
- Ручной DI-контейнер для прозрачного wiring зависимостей.
- Контракты через интерфейсы (удобно для тестирования и замены реализаций).
- Централизованные middleware (auth, rate limit, лимиты сообщений).
- Чистая структура `internal/*` без смешивания бизнес-логики и транспорта.

## Стек

- Go 1.23+
- Gin (`github.com/gin-gonic/gin`)
- GORM + PostgreSQL
- Firebase Admin SDK (опционально для локального запуска)

## Структура проекта

- `cmd/server/main.go` - точка входа приложения.
- `internal/transport/router` - регистрация роутов и группировка endpoint-ов.
- `internal/handler` - HTTP-слой (валидация/ответы).
- `internal/usecase` - бизнес-логика и orchestration.
- `internal/repository` - работа с БД.
- `internal/model` и `internal/migration` - модели и миграции.
- `internal/di/di.go` - сборка зависимостей приложения.

## Быстрый запуск

1. Подготовить окружение:

```bash
cp .env.example .env
```

2. Поднять PostgreSQL:

```bash
docker compose -f compose.yaml up -d db
```

3. Запустить API:

```bash
go run ./cmd/server
```

API стартует на `http://localhost:${API_PORT}` (`8000` в `.env.example`).

4. Проверить health:

```bash
curl http://localhost:8000/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## Полезные команды

```bash
make deps        # скачать зависимости
make run         # запустить сервер
make build       # собрать бинарник
make lint        # go fmt + go vet
go test ./...    # прогнать тесты
```

## Конфигурация и безопасность

- Секреты в репозиторий не коммитятся, только `.env.example`.
- Для полного набора auth/integration сценариев нужны:
  - `FIREBASE_CREDS_FILE` или `FIREBASE_CREDS_JSON`
  - `FIREBASE_WEB_API_KEY`
  - `OPENAI_KEY`
  - `TG_BOT_TOKEN`
- Если Firebase не настроен, сервер всё равно стартует; Firebase-зависимые endpoint-ы отвечают `503`.
- Dev/debug роуты выключены по умолчанию. Для локальной отладки:

```bash
ENABLE_DEV_ROUTES=true
```
