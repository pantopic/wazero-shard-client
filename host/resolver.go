package wazero_shard_client

import (
	"context"
)

type resolver struct {
	namespace string
	resource  string
}

func (r resolver) ContextCopy(dst, src context.Context) context.Context {
	dst = context.WithValue(dst, ctxKeyNamespace, r.namespace)
	dst = context.WithValue(dst, ctxKeyResource, r.resource)
	return dst
}

func NewResolver(namespace, resource string) resolver {
	return resolver{
		namespace: namespace,
		resource:  resource,
	}
}
