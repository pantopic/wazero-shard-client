package shard_client

type Client struct {
	shardName []byte
}

func New(name []byte) Client {
	return Client{name}
}

func (c Client) Read(query []byte, stale bool) (val uint64, res []byte, err error) {
	setShardName(c.shardName)
	setData(query)
	if stale {
		readlocal()
	} else {
		read()
	}
	return getVal(), getData(), getErr()
}

func (c Client) Apply(cmd []byte) (val uint64, res []byte, err error) {
	setShardName(c.shardName)
	setData(cmd)
	apply()
	return getVal(), getData(), getErr()
}

func (c Client) StreamOpen(name []byte) (err error) {
	if streamRecv == nil {
		return ErrStreamRecvNotRegistered
	}
	setShardName(c.shardName)
	setStreamName(name)
	streamOpen()
	return getErr()
}

func (c Client) StreamOpenLocal(name []byte) (err error) {
	if streamRecv == nil {
		return ErrStreamRecvNotRegistered
	}
	setShardName(c.shardName)
	setStreamName(name)
	streamOpenLocal()
	return getErr()
}

func (c Client) StreamSend(name, data []byte) (err error) {
	setShardName(c.shardName)
	setStreamName(name)
	setData(data)
	streamSend()
	return getErr()
}

func (c Client) StreamClose(name []byte) (err error) {
	setShardName(c.shardName)
	setStreamName(name)
	streamClose()
	return getErr()
}

func RegisterStreamRecv(fn streamRecvFunc) (err error) {
	if streamRecv != nil {
		return ErrStreamRecvAlreadyRegistered
	}
	streamRecv = fn
	return
}
