package shard_client

type Client struct {
	componentName []byte
	shardName     []byte
}

func New(componentName, shardName []byte) Client {
	return Client{componentName, shardName}
}

func (c Client) Read(query []byte, stale bool) (val uint64, res []byte, err error) {
	setComponentName(c.componentName)
	setShardName(c.shardName)
	setData(query)
	if stale {
		_read_local()
	} else {
		_read()
	}
	return getVal(), getData(), getErr()
}

func (c Client) Apply(cmd []byte) (val uint64, res []byte, err error) {
	setComponentName(c.componentName)
	setShardName(c.shardName)
	setData(cmd)
	_apply()
	return getVal(), getData(), getErr()
}

func (c Client) AsyncApply(cmd, name []byte) {
	setComponentName(c.componentName)
	setShardName(c.shardName)
	setStreamName(name)
	setData(cmd)
	_async_apply()
}

func (c Client) AsyncRead(query, name []byte, stale bool) {
	setComponentName(c.componentName)
	setShardName(c.shardName)
	setStreamName(name)
	setData(query)
	if stale {
		_async_read_local()
	} else {
		_async_read()
	}
}

func (c Client) StreamOpen(name []byte) (err error) {
	if streamRecv == nil {
		return ErrStreamRecvNotRegistered
	}
	setComponentName(c.componentName)
	setShardName(c.shardName)
	setStreamName(name)
	_streamOpen()
	return getErr()
}

func (c Client) StreamOpenLocal(name []byte) (err error) {
	if streamRecv == nil {
		return ErrStreamRecvNotRegistered
	}
	setComponentName(c.componentName)
	setShardName(c.shardName)
	setStreamName(name)
	_streamOpenLocal()
	return getErr()
}

func (c Client) StreamSend(name, data []byte) (err error) {
	setStreamName(name)
	setData(data)
	_streamSend()
	return getErr()
}

func (c Client) StreamClose(name []byte) (err error) {
	setStreamName(name)
	_streamClose()
	return getErr()
}

func RegisterStreamRecv(fn streamRecvFunc) (err error) {
	if streamRecv != nil {
		return ErrStreamRecvAlreadyRegistered
	}
	streamRecv = fn
	return
}

func RegisterAsyncRecv(fn asyncRecvFunc) (err error) {
	if asyncRecv != nil {
		return ErrAsyncRecvAlreadyRegistered
	}
	asyncRecv = fn
	return
}
