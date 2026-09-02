# Kafka

Клиент для kafka на основе `kgo`

## Использование

подключить можно через `fx`

```go
fx.Options(
    kafka.Module,
)
```

## Producer

Для использования требуется получить конфиг из пула и создать producer.
При остановке происходит закрытие коннекта, поэтому необходимо вызывать close в сервисе.

```go
func NewService(kafkaPool *kafka.Pool) (*Service, error) {
    kCfg, _ := kafkaPool.Get("some_service")
   
    producer, _ := kafka.NewProducer(kCfg)

    return &Service{
        producer: producer,
    }
}

func (s *Service) Stop(ctx context.Context) error {
    return s.producer.Close(ctx)
}
```

И дальше уже пишем в топики

```go
err = producer.Send(context.Background(), topic, payload)
```

## Consumer

Для исмользования консюмера, его надо создать.
Consumer вытягивает данные со всех партиций, для разделения, используте `ConsumerGroup`

```go
    consumer, _ := kafka.NewConsumer(kk)
    consumer.Subscribe("topic", func(ctx context.Context, p kafka.Payload) error{
        fmt.Println(p)
        return nil
    })

    consumer.Consume()

    consumer.Close()
```

Нужно учитывать что по умолчанию стоит оффсет с самого начала.
Для чтения только новых сообщений

```go
    consumer, _ := kafka.NewConsumer(kk, kafka.WithConsumeResetOffset(kgo.NewOffset().AtEnd())
    consumer.Subscribe("topic", func(ctx context.Context, p kafka.Payload) error{
        fmt.Println(p)
        return nil
    })

    consumer.Consume()

    consumer.Close()
```

Для чтения от определённого времени

```go
    n := time.Now()
    consumer, _ := kafka.NewConsumer(kk, kafka.WithConsumeResetOffset(kgo.NewOffset().AfterMilli(n.UnixMilli())))
    consumer.Subscribe("topic", func(ctx context.Context, p kafka.Payload) error{
        fmt.Println(p)
        return nil
    })

    consumer.Consume()

    consumer.Close()
```

Для чтения из нескольких топиков одной функцией

```go
    n := time.Now()
    consumer, _ := kafka.NewConsumer(kk, kafka.WithConsumeResetOffset(kgo.NewOffset().AfterMilli(n.UnixMilli())))
    handler := func(ctx context.Context, p kafka.Payload) error{
        fmt.Println(p)
        return nil
    }

    consumer.Subscribe("topic", handler)
    consumer.Subscribe("topic2", handler)

    consumer.Consume()

    consumer.Close()
```

## Consumer Group

Для создания consumer группы используйте
в функцию `consumer.Consume(n)` n - это кол-во записей для обработки

```go
    consumer, _ := kafka.NewConsumerGroup(kk, "group")
    consumer.Subscribe("topic", func(ctx context.Context, p kafka.Payload) error{
        fmt.Println(p)
        return nil
    })

    consumer.Consume(1)

    consumer.Close()
```

Для чтения из нескольких топиков одной функцией

```go
    consumer, _ := kafka.NewConsumerGroup(kk, "group")
    handler := func(ctx context.Context, p kafka.Payload) error{
        fmt.Println(p)
        return nil
    }

    consumer.Subscribe("topic", handler)
    consumer.Subscribe("topic2", handler)

    consumer.Consume(0)

    consumer.Close()
```
