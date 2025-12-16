package wazero_shard_client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/logbn/zongzi"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Name is the name of this host module.
const Name = "pantopic/wazero-shard-client"

var (
	DefaultCtxKeyMeta  = `wazero_shard_client_meta`
	DefaultCtxKeyAgent = `wazero_shard_client_agent`
)

type meta struct {
	ptrData         uint32
	ptrDataCap      uint32
	ptrDataLen      uint32
	ptrErr          uint32
	ptrErrCap       uint32
	ptrErrLen       uint32
	ptrShardName    uint32
	ptrShardNameCap uint32
	ptrShardNameLen uint32
	ptrVal          uint32
}

type hostModule struct {
	sync.RWMutex

	module      api.Module
	ctxKeyMeta  string
	ctxKeyAgent string

	resolveNamespace func(context.Context) string
	resolveResource  func(context.Context) string
}

func New(opts ...Option) *hostModule {
	p := &hostModule{
		ctxKeyMeta:  DefaultCtxKeyMeta,
		ctxKeyAgent: DefaultCtxKeyAgent,
		resolveNamespace: func(ctx context.Context) string {
			return `default`
		},
		resolveResource: func(ctx context.Context) string {
			return `default`
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *hostModule) Name() string {
	return Name
}

// Register instantiates the host module, making it available to all module instances in this runtime
func (p *hostModule) Register(ctx context.Context, r wazero.Runtime) (err error) {
	builder := r.NewHostModuleBuilder(Name)
	register := func(name string, fn func(ctx context.Context, m api.Module, stack []uint64)) {
		builder = builder.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(fn), nil, nil).Export(name)
	}
	for name, fn := range map[string]any{
		"Read": func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error) {
			// TODO [Auth] - Validate that invocation is authorized to read shard
			return client.Read(ctx, query, false)
		},
		"ReadLocal": func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error) {
			// TODO [Auth] - Validate that invocation is authorized to read shard
			return client.Read(ctx, query, true)
		},
		"Apply": func(ctx context.Context, client zongzi.ShardClient, cmd []byte) (val uint64, res []byte, err error) {
			// TODO [Auth] - Validate that invocation is authorized to write to shard
			return client.Apply(ctx, cmd)
			// val, res, err = client.Apply(ctx, cmd)
			// slog.Info(`Apply`, `client`, client, `cmd`, string(cmd), `val`, val, `res`, res, `err`, err)
			// return
		},
	} {
		switch fn := fn.(type) {
		case func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error):
			register(name, func(ctx context.Context, m api.Module, stack []uint64) {
				meta := get[*meta](ctx, p.ctxKeyMeta)
				client := p.agent(ctx).ClientByName(fmt.Sprintf(`%s.%s.%s`,
					p.resolveNamespace(ctx),
					p.resolveResource(ctx),
					getShardName(m, meta)))
				val, data, err := fn(ctx, client, getData(m, meta))
				setVal(m, meta, val)
				setData(m, meta, data)
				setErr(m, meta, err)
			})
		default:
			log.Panicf("Method signature implementation missing: %#v", fn)
		}
	}
	p.module, err = builder.Instantiate(ctx)
	return
}

// InitContext retrieves the meta page from the wasm module
func (p *hostModule) InitContext(ctx context.Context, m api.Module, agent *zongzi.Agent) (context.Context, error) {
	stack, err := m.ExportedFunction(`__shard_client`).Call(ctx)
	if err != nil {
		return ctx, err
	}
	meta := &meta{}
	ptr := uint32(stack[0])
	for i, v := range []*uint32{
		&meta.ptrVal,
		&meta.ptrShardNameCap,
		&meta.ptrShardNameLen,
		&meta.ptrShardName,
		&meta.ptrDataCap,
		&meta.ptrDataLen,
		&meta.ptrData,
		&meta.ptrErrCap,
		&meta.ptrErrLen,
		&meta.ptrErr,
	} {
		*v = readUint32(m, ptr+uint32(4*i))
	}
	ctx = context.WithValue(ctx, p.ctxKeyMeta, meta)
	ctx = context.WithValue(ctx, p.ctxKeyAgent, agent)
	return ctx, nil
}

// ContextCopy populates dst context with the meta page from src context.
func (h *hostModule) ContextCopy(src, dst context.Context) context.Context {
	dst = context.WithValue(dst, h.ctxKeyMeta, get[*meta](src, h.ctxKeyMeta))
	dst = context.WithValue(dst, h.ctxKeyAgent, h.agent(src))
	return dst
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

func getShardName(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrShardName, meta.ptrShardNameLen, meta.ptrShardNameCap)
}

func setShardID(m api.Module, meta *meta, val uint64) {
	writeUint64(m, meta.ptrVal, val)
}

func readUint32(m api.Module, ptr uint32) (val uint32) {
	val, ok := m.Memory().ReadUint32Le(ptr)
	if !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
	return
}

func getData(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrData, meta.ptrDataLen, meta.ptrDataCap)
}

func dataBuf(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrData, 0, meta.ptrDataCap)
}

func setVal(m api.Module, meta *meta, val uint64) {
	writeUint64(m, meta.ptrVal, val)
}

func setData(m api.Module, meta *meta, b []byte) {
	copy(dataBuf(m, meta)[:len(b)], b)
	writeUint32(m, meta.ptrDataLen, uint32(len(b)))
}

func errBuf(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrErr, 0, meta.ptrErrCap)
}

func setErr(m api.Module, meta *meta, err error) {
	var msg string
	if err != nil {
		msg = err.Error()
		copy(errBuf(m, meta)[:len(msg)], msg)
	}
	writeUint32(m, meta.ptrErrLen, uint32(len(msg)))
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

func writeUint64(m api.Module, ptr uint32, val uint64) {
	if ok := m.Memory().WriteUint64Le(ptr, val); !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
}
