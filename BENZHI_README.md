本项目为 Go 库/组件（含配套管理页）：令牌桶限流（可插拔 Store/分片），配套 limitd。

# RateLimit

```text
go build ./...
go test ./... -count=1
go run ./cmd/limitd -addr :8224
```
