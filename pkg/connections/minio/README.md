# Minio

Клиент для S3 на основе minio

## Использование

Подключение через `fx`

```go
fx.Options(
    minio.Module,
)
```

После в сервисе получаем клиента S3

```go
func NewService(minio *minio.Client) (*Service, error) {
    return &Service{
        s3: minio,
    }
}
```
