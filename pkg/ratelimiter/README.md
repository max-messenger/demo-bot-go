# Rate limiter

Tут добавляется 2 модуля для рейтлимитера, local и shared, конфиг для рейт лимита общий.
Подключить и использовать можно оба

Лимитеры можно подменять, у них общий интерфейс

```go
var errLimitExceeded = errors.New("limit exceeded")

type Limiter interface{
    Limit(context.Context, Key, Action) (err error)
}
```

## Config

Для кофигурирования лимитов можно добавить следующие значения

```yaml
rate_limit:
    local: # настройки для локального конфига
        cache_size: 1024 # размер кеша, если не установленно, то 1024 * 1024, должно быть больше 2
        ttl: "2s" # настройка кеша default 2s
        custom: # кастомные кеши для рейтлимита, где требуется более тонкая настройка
            %custom_name%: 
                cache_size: 1024
    shared: # настройка shared лимитера, используется redis
        prefix: "service_name" # префикс для рейтлимитера
        pool_name: "pool_name" # имя redisпула для рейтлимитера
    rate:
        action_name: # action для рейт лимита
            limit: 100 # лимит
            keys: # cпец лимиты для определённых ключей
                key_name: 200 #
```

Для настройки кастомных лимит гетеров, для прокидывания лимита в зависимости от функции или внешних API, экспортим через fx функцию с гуппой `rate_limit_custom`

```go
type CustomLimiterOut struct {
    fx.Out

    LimitGetter ratelimiter.CustomLimitFunc `group:"rate_limit_custom"`
}

func adapterLimitGetter(rd *Redis) CustomLimiterOut {
    return HealthComponentOut{
        LimitGetter: func() (ratelimiter.Action, ratelimiter.LimitGet) {
            return "your_action", func(key ratelimiter.Key, action ratelimiter.Action) ratelimiter.Limit {
                return 1
            }
        }
    }
}

```

## Local

Локальный рейтлимитер, подключить можно через `fx`

```go
fx.Options(
    ratelimiter.ModuleLocal,
)
```

Ипортим себе в проект и используем.

```go
func NewSomething(rl *ratelimiter.LocalLimiter) *Something {
    err := sw.Limit(ctx, "key", "action")
    if err != nil {
        if errors.Is(err, ratelimiter.ErrLimitExceeded) {
            // exeed
        }
        // error, not ok
    }
    // ok
}
```

Под капотом создаётся 10 кешей, в которых храним лимиты, если требуется отдельный кеш для вашего лимита, то созадётся через конфиг

## Shared

Распределённый рейтлимитер, подключить можно через `fx`
Для работы нужно указать пул redis в конфиге который будет использоваться для хранения лимитов

```go
fx.Options(
    ratelimiter.ModuleShared,
)
```

```go
func NewSomething(rl *ratelimiter.SharedLimiter) *Something {
    err := sw.Limit(ctx, "key", "action")
    if err != nil {
        if errors.Is(err, ratelimiter.ErrLimitExceeded) {
            // exeed
        }
        // error, not ok
    }
    // ok
}
```
