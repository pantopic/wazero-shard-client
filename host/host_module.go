package wazero_shard_client

import (
	"context"
	"log"
	"sync"

	"github.com/logbn/zongzi"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

var (
	DefaultCtxKeyMeta  = `wazero_shard_client_meta_key`
	DefaultCtxKeyAgent = `wazero_shard_client_meta_agent`
)

type meta struct {
	ptrShardID uint32
	ptrVal     uint32
	ptrDataMax uint32
	ptrDataLen uint32
	ptrData    uint32
	ptrErrCode uint32
}

type hostModule struct {
	sync.RWMutex

	module      api.Module
	ctxKeyMeta  string
	ctxKeyAgent string
}

func New(opts ...Option) *hostModule {
	p := &hostModule{
		ctxKeyMeta:  DefaultCtxKeyMeta,
		ctxKeyAgent: DefaultCtxKeyAgent,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *hostModule) Uri() string {
	return "github.com/pantopic/wazero-shard-client"
}

// Register instantiates the host module, making it available to all module instances in this runtime
func (p *hostModule) Register(ctx context.Context, r wazero.Runtime) (err error) {
	builder := r.NewHostModuleBuilder("shard_client")
	register := func(name string, fn func(ctx context.Context, m api.Module, stack []uint64)) {
		builder = builder.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(fn), nil, nil).Export(name)
	}
	for name, fn := range map[string]any{
		"Read": func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error) {
			return client.Read(ctx, query, true)
		},
		"ReadLocal": func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error) {
			return client.Read(ctx, query, false)
		},
		"Apply": func(ctx context.Context, client zongzi.ShardClient, cmd []byte) (val uint64, res []byte, err error) {
			return client.Apply(ctx, cmd)
		},
	} {
		switch fn := fn.(type) {
		case func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error):
			register(name, func(ctx context.Context, m api.Module, stack []uint64) {
				meta := get[*meta](ctx, p.ctxKeyMeta)
				client := p.agent(ctx).Client(shardID(m, meta))
				val, data, err := fn(ctx, client, data(m, meta))
				setVal(val)
				setData(data)
				writeError(m, meta, err)
			})
		default:
			log.Panicf("Method signature implementation missing: %#v", fn)
		}
	}
	p.module, err = builder.Instantiate(ctx)
	return
}

// InitContext retrieves the meta page from the wasm module
func (p *hostModule) InitContext(ctx context.Context, m api.Module) (context.Context, *meta, error) {
	stack, err := m.ExportedFunction(`__shard_client`).Call(ctx)
	if err != nil {
		return ctx, nil, err
	}
	meta := &meta{}
	ptr := uint32(stack[0])
	for i, v := range []*uint32{
		&meta.ptrShardID,
		&meta.ptrVal,
		&meta.ptrDataMax,
		&meta.ptrDataLen,
		&meta.ptrData,
		&meta.ptrErrCode,
	} {
		*v = readUint32(m, ptr+uint32(4*(i+2)))
	}
	return context.WithValue(ctx, p.ctxKeyMeta, meta), meta, nil
}

func (p *hostModule) agent(ctx context.Context) *zongzi.Agent {
	return get[*zongzi.Agent](ctx, p.ctxKeyAgent)
}

func get[T any](ctx context.Context, key string) T {
	v := ctx.Value(key)
	if v == nil {
		log.Panicf("Context item missing %s", key)
	}
	return v.(T)
}

func shardID(m api.Module, meta *meta) uint64 {
	return readUint64(m, meta.ptrShardID)
}

func readUint32(m api.Module, ptr uint32) (val uint32) {
	val, ok := m.Memory().ReadUint32Le(ptr)
	if !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
	return
}

func data(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrVal, meta.ptrDataLen, meta.ptrDataMax)
}

func read(m api.Module, ptrData, ptrLen, ptrMax uint32) (buf []byte) {
	buf, ok := m.Memory().Read(ptrData, readUint32(m, ptrMax))
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", ptrData, ptrLen)
	}
	return buf[:readUint32(m, ptrLen)]
}

func readUint64(m api.Module, ptr uint32) (val uint64) {
	val, ok := m.Memory().ReadUint64Le(ptr)
	if !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
	return
}

func writeUint32(m api.Module, ptr uint32, val uint32) {
	if ok := m.Memory().WriteUint32Le(ptr, val); !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
}
