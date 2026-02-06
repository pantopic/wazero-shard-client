package shard_client

import (
	"unsafe"
)

type streamRecvFunc func(shardName, streamName, data []byte, val uint64)

var (
	meta          = make([]uint32, 13)
	val           uint64
	shardNameCap  uint32 = 256
	shardNameLen  uint32
	dataCap       uint32 = 2 << 20 // 2 MiB
	dataLen       uint32
	errCap        uint32 = 1024
	errLen        uint32
	streamNameCap uint32 = 64
	streamNameLen uint32

	streamName = make([]byte, 64)
	shardName  = make([]byte, int(shardNameCap))
	data       = make([]byte, int(dataCap))
	err        = make([]byte, int(errCap))

	streamRecv streamRecvFunc
)

//export __shard_client
func __shard_client() uint32 {
	meta[0] = uint32(uintptr(unsafe.Pointer(&val)))
	meta[1] = uint32(uintptr(unsafe.Pointer(&shardNameCap)))
	meta[2] = uint32(uintptr(unsafe.Pointer(&shardNameLen)))
	meta[3] = uint32(uintptr(unsafe.Pointer(&shardName[0])))
	meta[4] = uint32(uintptr(unsafe.Pointer(&dataCap)))
	meta[5] = uint32(uintptr(unsafe.Pointer(&dataLen)))
	meta[6] = uint32(uintptr(unsafe.Pointer(&data[0])))
	meta[7] = uint32(uintptr(unsafe.Pointer(&errCap)))
	meta[8] = uint32(uintptr(unsafe.Pointer(&errLen)))
	meta[9] = uint32(uintptr(unsafe.Pointer(&err[0])))
	meta[10] = uint32(uintptr(unsafe.Pointer(&streamNameCap)))
	meta[11] = uint32(uintptr(unsafe.Pointer(&streamNameLen)))
	meta[12] = uint32(uintptr(unsafe.Pointer(&streamName[0])))
	return uint32(uintptr(unsafe.Pointer(&meta[0])))
}

//export __shard_client_stream_recv
func __shard_client_stream_recv() {
	streamRecv(getShardName(), getStreamName(), getData(), getVal())
}

func setShardName(name []byte) {
	copy(shardName[:len(name)], name)
	shardNameLen = uint32(len(name))
}

func getShardName() []byte {
	return shardName[:shardNameLen]
}

func setData(v []byte) {
	copy(data[:len(v)], v)
	dataLen = uint32(len(v))
}

func getData() []byte {
	return data[:dataLen]
}

func setErr(e error) {
	b := []byte(e.Error())
	copy(err[:len(b)], b)
	errLen = uint32(len(b))
}

func getErr() (e error) {
	if errLen > 0 {
		e = strErr(string(err[:errLen]))
	}
	return
}

func getVal() uint64 {
	return val
}

func getStreamName() []byte {
	return streamName[:streamNameLen]
}

func setStreamName(name []byte) {
	copy(streamName[:len(name)], name)
	streamNameLen = uint32(len(name))
}

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_read
func read()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_read_local
func readlocal()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_apply
func apply()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_open
func streamOpen()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_open_local
func streamOpenLocal()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_send
func streamSend()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_close
func streamClose()

var _ = __shard_client
var _ = __shard_client_stream_recv
var _ = getShardName
var _ = setData
