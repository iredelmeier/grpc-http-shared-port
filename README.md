# grpc-http-shared-port

## Demo of `grpc.Server#ServeHTTP`

This is a small demo of `grpc.Server#ServeHTTP`, which allows you to run a gRPC and non-gRPC HTTP server on the same port.

### Installation

```bash
go get github.com/iredelmeier/grpc-http-shared-port
```

### With TLS

```bash
go run h2/main.go
```

### With TLS

```bash
go run h2c/main.go
```
