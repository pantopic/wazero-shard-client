package shard_client

type Client struct {
	shardName string
}

func New(name string) Client {
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
