package shard_client

var (
	ErrStreamRecvAlreadyRegistered = strErr(`StreamRecv Already Registered`)
	ErrStreamRecvNotRegistered     = strErr(`StreamRecv Not Registered`)
	ErrAsyncRecvAlreadyRegistered  = strErr(`AsyncRecv Already Registered`)
	ErrAsyncRecvNotRegistered      = strErr(`AsyncRecv Not Registered`)
)

type strErr string

func (e strErr) Error() string {
	return string(e)
}
