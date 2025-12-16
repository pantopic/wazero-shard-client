package wazero_shard_client

import (
	"context"
)

type Option func(*hostModule)

func WithCtxKeyMeta(key string) Option {
	return func(p *hostModule) {
		p.ctxKeyMeta = key
	}
}

// WithNamespace specifies a static namespace for all invocations
func WithNamespace(name string) Option {
	return WithNamespaceResolver(func(ctx context.Context) string {
		return name
	})
}

// WithResource specifies a static resource name for all invocations
func WithResource(name string) Option {
	return WithResourceResolver(func(ctx context.Context) string {
		return name
	})
}

// WithNamespaceResolver specifies a callback for resolving namespace on each invocation
func WithNamespaceResolver(r func(ctx context.Context) string) Option {
	return func(p *hostModule) {
		p.resolveNamespace = r
	}
}

// WithNamespaceResolver specifies a callback for resolving resource name on each invocation
func WithResourceResolver(r func(ctx context.Context) string) Option {
	return func(p *hostModule) {
		p.resolveResource = r
	}
}
