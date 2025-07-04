# Wazero Shard Client

A [wazero](https://pkg.go.dev/github.com/tetratelabs/wazero) host module, ABI and guest SDK providing a pantopic cluster shard client for WASI modules.

## Host Module

<!-- [![Go Reference](https://godoc.org/github.com/pantopic/wazero-shard-client/host?status.svg)](https://godoc.org/github.com/pantopic/wazero-shard-client/host)
[![Go Report Card](https://goreportcard.com/badge/github.com/pantopic/wazero-shard-client/host)](https://goreportcard.com/report/github.com/pantopic/wazero-shard-client/host)
[![Go Coverage](https://github.com/pantopic/wazero-shard-client/wiki/host/coverage.svg)](https://raw.githack.com/wiki/pantopic/wazero-shard-client/host/coverage.html) -->

First register the host module with the runtime

```go
import (
    "github.com/tetratelabs/wazero"
    "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

    "github.com/pantopic/wazero-shard-client/host"
)

func main() {
    ctx := context.Background()
    r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig())
    wasi_snapshot_preview1.MustInstantiate(ctx, r)

    module := wazero_shard_client.New()
    module.Register(ctx, r)

    // ...
}
```

## Guest SDK (Go)

<!-- [![Go Reference](https://godoc.org/github.com/pantopic/wazero-shard-client/shard-client-go?status.svg)](https://godoc.org/github.com/pantopic/wazero-shard-client/shard-client-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/pantopic/wazero-shard-client/shard-client-go)](https://goreportcard.com/report/github.com/pantopic/wazero-shard-client/shard-client-go) -->

Then you can import the guest SDK into your WASI module to interact with pantopic cluster shards from WASM.

```go
package main

import (
    "unsafe"

    "github.com/pantopic/wazero-shard-client/shard-client-go"
)

func main() {}

var shardID uint64 = 123

//export get
func get() uint64 {
    _, data, _ := shard_client.Read(shardID, []byte(`GET /test`))
    return uint64(uintptr(unsafe.Pointer(&data[0])))<<32 + uint64(len(data))
}

//export getstale
func getstale() uint64 {
    _, data, _ := shard_client.ReadLocal(shardID, []byte(`GET /test`))
    return uint64(uintptr(unsafe.Pointer(&data[0])))<<32 + uint64(len(data))
}

//export put
func put() uint64 {
    val, _, _ := shard_client.Apply(shardID, []byte(`PUT /test {"a": 1}`))
    return val
}

//export index
func index() uint64 {
    val, _ := shard_client.Index(shardID)
    return val
}
```

The [guest SDK](https://pkg.go.dev/github.com/pantopic/wazero-shard-client/shard-client) has no dependencies outside the Go std lib.

## Roadmap

This project is in alpha. Breaking API changes should be expected until Beta.

- `v0.0.x` - Alpha
  - [ ] Stabilize API
- `v0.x.x` - Beta
  - [ ] Finalize API
  - [ ] Test in production
- `v1.x.x` - General Availability
  - [ ] Proven long term stability in production
