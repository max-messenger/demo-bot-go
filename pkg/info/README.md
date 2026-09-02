# Info

Модуль для получения информации о приложении через HTTP endpoint.

## Обзор

Пакет `info` предоставляет HTTP handler, который возвращает метаинформацию о приложении в формате JSON, включая:

- Информацию о Git коммите (автор, SHA, timestamp, сообщение, ветка)
- Информацию о сборке (artifact name, version)

Переменные заполняются через ldflags при сборке приложения.

## Структуры ответа

### Response

```json
{
  "git": {
    "commit": {
      "user": {
        "name": "John Doe"
      },
      "id": "a1b2c3d",
      "time": "2024-01-15T10:30:00Z",
      "message": {
        "full": "Fix critical bug in payment processing"
      }
    },
    "branch": "main"
  },
  "build": {
    "artifact": "my-service",
    "time": "2024-01-15T10:30:00Z",
    "version": "v1.2.3"
  }
}
```

Если тег отсутствует, версия будет равна короткому SHA:

```json
{
  "build": {
    "artifact": "my-service",
    "time": "2024-01-15T10:30:00Z",
    "version": "a1b2c3d"
  }
}
```

## Использование

### Настройка ldflags

При сборке приложения необходимо передать Git информацию через ldflags:

```bash
go build \
  -ldflags="\
    -X 'demo_bot/pkg/info.CommitAuthor=John Doe'\
    -X 'demo_bot/pkg/info.CommitShortSHA=a1b2c3d'\
    -X 'demo_bot/pkg/info.CommitTimestamp=2024-01-15T10:30:00Z'\
    -X 'demo_bot/pkg/info.CommitMessage=Fix critical bug'\
    -X 'demo_bot/pkg/info.CommitTag=v1.2.3'\
    -X 'demo_bot/pkg/info.CommitRefName=main'\
  " \
  -o app ./cmd/app
```
