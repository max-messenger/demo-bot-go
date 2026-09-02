# GRACE

Пакет для управлення жизненным циклом сервисов с поддержкой graceful shutdown (корректного завершения).

## Обзор

Пакет `grace` предоставляет пул сервисов (`ServicePool`), который позволяет:

- Запускать множество сервисов одновременно
- Корректно останавливать все сервисы при завершении работы приложения
- Использовать параллельное завершение для быстрого shutdown

## Компоненты

### Service

Интерфейс, который должен реализовать каждый сервис, управляемый пулом:

```go
type Service interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

## Использование

### Регистрация сервиса

Чтобы сервис управлялся пулом, добавьте его в группу `"grace"`:

```go
func NewService() *MyService {
    return &MyService{}
}

// Регистрация в FX
fx.Provide(
    NewService,
    graceAdapter,
),

type graceOut struct {
    fx.Out

    Service grace.Service `group:"grace"`
}

func graceAdapter(srvc *MyService) graceOut {
    return graceOut{
        Service: srvc,
}
}

```

### Пример интеграции

```go
app := fx.New(
    grace.Module,
    fx.Provide(
        NewHTTPServer,
    ),
    fx.Invoke(func(lc fx.Lifecycle, pool *grace.ServicePool) {
        lc.Append(fx.Hook{
            OnStart: func(ctx context.Context) error {
                return pool.Start(ctx)
            },
            OnStop: func(ctx context.Context) error {
                return pool.Stop(ctx)
            },
        })
    }),
)
```
