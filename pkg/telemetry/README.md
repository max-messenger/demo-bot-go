# Telemetry

Пакет для настройки отправки трейсов, а также хелперов для сбора метрик.

## Tracer

Настройка tracer происходит через конфигурацию в файле `config.yaml`.

```yaml
telemetry:
  collector:
    enabled: true
    endpoint: localhost:4317
    sampling_ratio: 100
```

**Поля:**

- `enabled` — включен ли tracer
- `endpoint` — адрес сервера для отправки трейсов (GRPC)
- `sampling_ratio` — доля трейсов, которые будут отправляться в сервер
