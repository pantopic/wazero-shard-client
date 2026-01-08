package wazero_shard_client

import (
	"fmt"
)

var (
	ErrStreamNotFound = fmt.Errorf(`Stream not found`)
	ErrStreamExists   = fmt.Errorf(`Stream exists`)
	ErrStreamClosed   = fmt.Errorf(`Stream closed`)
)
