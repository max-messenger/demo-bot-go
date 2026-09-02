# Background Tasker

Это модуль для запуска фоновых задачь в приложении

## Использование

Для начала подключить модуль в файле fx

```go
fx.Options(
    bgtasker.Module,
)
```

Конфиг для запуска задачи в фоне, по умолчанию используется тип drop

**drop** - если буффер переполнен, то задача будет отклонена
**block** - если буффер переполнен, то задача будет блокирована и ждет пока не сможет добавить ее в буффер

```yaml
background:
  tasks:
    %name%:
      async_workers_num: 16
      async_buffer_size: 4096
      async_worker_timeout: "2s"
      runner_type: drop # block, drop
```

## Пример запуска задачи в фоне

```go
func Run(ctx ctx context.Context bgTaskerPool *bgtasker.Pool, poolName string) error {
    bgRunner, err := bgTaskerPool.Get(poolName)
    if err != nil {
        return nil, err
    }

    bgRunner.Run(ctx, func(ctx context.Context) error {
        // do something in background
    
        return nil
    })
    return nil
}
```
