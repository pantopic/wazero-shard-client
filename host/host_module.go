package wazero_shard_client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/logbn/zongzi"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/pantopic/wazero-pool"
)

// Name is the name of this host module.
const Name = "pantopic/wazero-shard-client"

var (
	DefaultCtxKeyMeta       = `wazero_shard_client_meta`
	DefaultCtxKeyAgent      = `wazero_shard_client_agent`
	DefaultCtxKeyModPool    = `wazero_shard_client_mod_pool`
	DefaultCtxKeyStreamList = `wazero_shard_client_stream_list`
)

type meta struct {
	ptrData          uint32
	ptrDataCap       uint32
	ptrDataLen       uint32
	ptrErr           uint32
	ptrErrCap        uint32
	ptrErrLen        uint32
	ptrShardName     uint32
	ptrShardNameCap  uint32
	ptrShardNameLen  uint32
	ptrStreamName    uint32
	ptrStreamNameCap uint32
	ptrStreamNameLen uint32
	ptrVal           uint32
}

type hostModule struct {
	sync.RWMutex

	module           api.Module
	ctxKeyMeta       string
	ctxKeyAgent      string
	ctxKeyModPool    string
	ctxKeyStreamList string

	resolveNamespace func(context.Context) string
	resolveResource  func(context.Context) string
}

func New(opts ...Option) *hostModule {
	p := &hostModule{
		ctxKeyMeta:       DefaultCtxKeyMeta,
		ctxKeyAgent:      DefaultCtxKeyAgent,
		ctxKeyModPool:    DefaultCtxKeyModPool,
		ctxKeyStreamList: DefaultCtxKeyStreamList,
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
		"__shard_client_apply": func(ctx context.Context, client zongzi.ShardClient, cmd []byte) (val uint64, res []byte, err error) {
			return client.Apply(ctx, cmd)
		},
		"__shard_client_read": func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error) {
			return client.Read(ctx, query, false)
		},
		"__shard_client_read_local": func(ctx context.Context, client zongzi.ShardClient, query []byte) (val uint64, res []byte, err error) {
			return client.Read(ctx, query, true)
		},
		"__shard_client_stream_open": func(ctx context.Context, client zongzi.ShardClient, name []byte) (err error) {
			s, err := p.getStreamList(ctx).new(ctx, name)
			if err != nil {
				return
			}
			s.wg.Go(func() {
				if err := client.Stream(ctx, s.in, s.out, false); err != nil {
					slog.Error("Error opening stream", "name", s.name, "err", err)
				}
				s.close()
			})
			s.wg.Go(func() {
				for {
					select {
					case res := <-s.out:
						meta := get[*meta](ctx, p.ctxKeyMeta)
						wazeropool.Context(ctx).Run(func(mod api.Module) {
							setStreamName(mod, meta, name)
							setVal(mod, meta, res.Value)
							setData(mod, meta, res.Data)
							setErr(mod, meta, nil)
							if _, err = mod.ExportedFunction("__shard_client_stream_recv").Call(ctx); err != nil {
								return
							}
							if err = getErr(mod, meta); err != nil {
								slog.Error("Error receiving stream message", "name", s.name, "err", err.Error())
								s.close()
								return
							}
						})
					case <-s.ctx.Done():
						return
					}
				}
			})
			return
		},
		"__shard_client_stream_open_local": func(ctx context.Context, client zongzi.ShardClient, name []byte) (err error) {
			s, err := p.getStreamList(ctx).new(ctx, name)
			if err != nil {
				return
			}
			s.wg.Go(func() {
				if err := client.Stream(ctx, s.in, s.out, true); err != nil {
					slog.Error("Error opening stream local", "name", s.name, "err", err)
				}
				s.close()
			})
			return
		},
		"__shard_client_stream_send": func(ctx context.Context, name, data []byte) (err error) {
			s, err := p.getStreamList(ctx).find(name)
			if err != nil {
				return
			}
			select {
			case s.in <- data:
			case <-ctx.Done():
				err = ErrStreamClosed
			}
			return
		},
		"__shard_client_stream_close": func(ctx context.Context, name []byte) (err error) {
			s, err := p.getStreamList(ctx).find(name)
			if err != nil {
				return
			}
			if s == nil {
				err = ErrStreamNotFound
				return
			}
			s.close()
			return
		},
	} {
		switch fn := fn.(type) {
		case func(context.Context, zongzi.ShardClient, []byte) (uint64, []byte, error):
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
		case func(context.Context, zongzi.ShardClient, []byte) (err error):
			register(name, func(ctx context.Context, m api.Module, stack []uint64) {
				meta := get[*meta](ctx, p.ctxKeyMeta)
				client := p.agent(ctx).ClientByName(fmt.Sprintf(`%s.%s.%s`,
					p.resolveNamespace(ctx),
					p.resolveResource(ctx),
					getShardName(m, meta)))
				err := fn(ctx, client, getStreamName(m, meta))
				setErr(m, meta, err)
			})
		case func(context.Context, []byte, []byte) (err error):
			register(name, func(ctx context.Context, m api.Module, stack []uint64) {
				meta := get[*meta](ctx, p.ctxKeyMeta)
				err := fn(ctx, getStreamName(m, meta), getData(m, meta))
				setErr(m, meta, err)
			})
		case func(context.Context, []byte) (err error):
			register(name, func(ctx context.Context, m api.Module, stack []uint64) {
				meta := get[*meta](ctx, p.ctxKeyMeta)
				err := fn(ctx, getStreamName(m, meta))
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
		&meta.ptrStreamNameCap,
		&meta.ptrStreamNameLen,
		&meta.ptrStreamName,
	} {
		*v = readUint32(m, ptr+uint32(4*i))
	}
	ctx = context.WithValue(ctx, p.ctxKeyMeta, meta)
	ctx = context.WithValue(ctx, p.ctxKeyAgent, agent)
	return ctx, nil
}

// ContextCopy populates dst context with the meta page from src context.
func (h *hostModule) ContextCopy(dst, src context.Context) context.Context {
	dst = context.WithValue(dst, h.ctxKeyMeta, get[*meta](src, h.ctxKeyMeta))
	dst = context.WithValue(dst, h.ctxKeyAgent, h.agent(src))
	dst = context.WithValue(dst, h.ctxKeyStreamList, newStreamList())
	return dst
}

// ContextClose cleans up resources created during ContextCopy.
func (h *hostModule) ContextClose(ctx context.Context) {
	h.getStreamList(ctx).release()
}

func (p *hostModule) agent(ctx context.Context) *zongzi.Agent {
	return get[*zongzi.Agent](ctx, p.ctxKeyAgent)
}

func (p *hostModule) getStreamList(ctx context.Context) *streamList {
	return get[*streamList](ctx, p.ctxKeyStreamList)
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

func setShardName(m api.Module, meta *meta, name []byte) {
	buf := read(m, meta.ptrShardName, 0, meta.ptrShardNameCap)
	copy(buf[:len(name)], name)
	writeUint32(m, meta.ptrShardNameLen, uint32(len(name)))
}

func getStreamName(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrStreamName, meta.ptrStreamNameLen, meta.ptrStreamNameCap)
}

func setStreamName(m api.Module, meta *meta, name []byte) {
	buf := read(m, meta.ptrStreamName, 0, meta.ptrStreamNameCap)
	copy(buf[:len(name)], name)
	writeUint32(m, meta.ptrStreamNameLen, uint32(len(name)))
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

func getErr(m api.Module, meta *meta) (err error) {
	if b := read(m, meta.ptrErr, meta.ptrErrLen, meta.ptrErrCap); len(b) > 0 {
		err = errors.New(string(b))
	}
	return
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
