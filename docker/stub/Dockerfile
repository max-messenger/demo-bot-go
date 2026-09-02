FROM golang:1.26.1 AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o stub ./cmd/stub

FROM alpine:3.20
COPY --from=builder /app/stub /app/stub
ENTRYPOINT ["/app/stub"]
CMD ["-addr", ":19090"]
