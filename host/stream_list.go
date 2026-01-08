package wazero_shard_client

import (
	"context"
	"encoding/base64"
	"sync"

	"github.com/logbn/zongzi"
)

var streamListPool = sync.Pool{
	New: func() any {
		return &streamList{
			items: make(map[string]*stream),
		}
	},
}

func newStreamList() *streamList {
	return streamListPool.Get().(*streamList)
}

type streamList struct {
	items map[string]*stream
	mutex sync.Mutex
}

func (sl *streamList) release() {
	if sl == nil {
		return
	}
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	for k, s := range sl.items {
		s.cancel()
		delete(sl.items, k)
	}
	defer streamListPool.Put(sl)
}

func (sl *streamList) new(ctx context.Context, name []byte) (s *stream, err error) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	s, ok := sl.items[base64.URLEncoding.EncodeToString(name)]
	if ok {
		return s, ErrStreamExists
	}
	s = &stream{
		name: name,
		in:   make(chan []byte),
		out:  make(chan *zongzi.Result),
		list: sl,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	return
}

func (sl *streamList) find(name []byte) (s *stream, err error) {
	s, ok := sl.items[base64.URLEncoding.EncodeToString(name)]
	if !ok {
		err = ErrStreamNotFound
	}
	return
}

func (s *stream) close() {
	s.list.mutex.Lock()
	defer s.list.mutex.Unlock()
	s.cancel()
	delete(s.list.items, base64.URLEncoding.EncodeToString(s.name))
}

type stream struct {
	cancel context.CancelFunc
	ctx    context.Context
	in     chan []byte
	list   *streamList
	name   []byte
	out    chan *zongzi.Result
	wg     sync.WaitGroup
}
