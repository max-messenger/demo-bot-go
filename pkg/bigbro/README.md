# BigBro MAX

Пакет для отправки ивентов в сервис аналитики.

## Config

Настройка происходит через конфигурацию в файле `config.yaml`.

```yaml
bigbro:
  enabled: true
  token: some-token
  url: https://test.ru
  extra_fields:
    app: demo_bot
```

**Поля:**

- `enabled` — включена ли отправка в сервис аналитики
- `token` — токен приложения, заводится отдельно на каждый сервис внутри аналитики
- `url` — хост, куда будет отправляться аналитика
- `extra_fields` — мапа дополнительных константных полей


## Использование

Подключить нужно через `fx`. Подключаем модуль в `internal/app/fx.go`

```go
fx.Options(
    bigbro.Module,
)
```

Для отправки ивента используется метод SendEvent

```go
SendEvent(ctx context.Context, eventName string, eventType EventType, userID int64, eventFields EventFields)
```

**Поля:**

- `eventName` — название события
- `eventType` — Тип события. Может быть только `click`, `action`, `navgo` или `view`
- `userID` — Идентификатор пользователя
- `eventFields` — Любой произвольный json-объект `map[string]any`

### Асинхронная отправка события

Асинхроннную отправку можно реализовать через фоновые задачи (например, `pkg/bgtasker`).
