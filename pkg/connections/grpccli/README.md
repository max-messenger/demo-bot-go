# GRPC Clients

## Использование

подключить можно через `fx`

```go
fx.Options(
    grpccli.Module,
)
```

После в сервисе вызываем получение коннекта

```go
func NewService(grpcPool *grpccli.Pool) (*Service, error) {
    conn, err := grpcPool.Get("some_service")
    if err != nil {
        return nil, err
    }
    return &Service{
        client: NewGrpcClient(conn.RawConn()), // используем коннект для создания grpc клиента
    }
}
```
