package wazero_shard_client

type Option func(*hostModule)

func WithCtxKeyMeta(key string) Option {
	return func(p *hostModule) {
		p.ctxKeyMeta = key
	}
}
