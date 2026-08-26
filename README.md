# go-ratelimit

分布式令牌桶限流：内存 + 分片 Store 抽象，配套 dashboard/ 管理台。

```text
go build ./...
go test ./... -count=1
go run ./cmd/limitd -addr :8224
```
