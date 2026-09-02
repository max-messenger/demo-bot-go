# Redis Pool

Клиент для redis на основе `redis-go/v9`

## Использование

Подключить можно через `fx`

```go
fx.Options(
    rediscli.Module,
)
```

После в сервисе вызываем получение коннекта к кластеру

```go
func NewService(redisPool *rediscli.Pool) (*Service, error) {
    redis, _ := redisPool.Get("some_service")
    
    return &Service{
        redis: redis,
    }
}
```

Дальше уже можно использовать коннект для работы с redis

Для форматирования ключей используется функция `rediscli.KeyFormatter` которая нужна для того чтобы при вызове к redis указывать ключ с префиксом и разделителем,
требуется для ипользования одного кластера на нескольких дев средах.

```go
func (s *Service) DoSomething(ctx context.Context) error {
    return s.redis.Call(ctx, "call_name", func(ctx context.Context, clu *redis.ClusterClient, kf rediscli.KeyFormatter) error {
        key := kf.FormatKey("some_key")
        return clu.Set(ctx, key, "some_value", 0).Err()
    })
}
```
