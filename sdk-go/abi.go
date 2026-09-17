package shard_client

import (
	"unsafe"
)

type streamRecvFunc func(name, data []byte, val uint64)
type asyncRecvFunc func(name, data []byte, val uint64, err error)

var (
	meta             = make([]uint32, 16)
	val              uint64
	shardNameCap     uint32 = 64
	shardNameLen     uint32
	componentNameCap uint32 = 64
	componentNameLen uint32
	dataCap          uint32 = 2 << 20 // 2 MiB
	dataLen          uint32
	errCap           uint32 = 16 << 10 // 16 KiB
	errLen           uint32
	streamNameCap    uint32 = 64
	streamNameLen    uint32

	componentName = make([]byte, int(componentNameCap))
	data          = make([]byte, int(dataCap))
	err           = make([]byte, int(errCap))
	shardName     = make([]byte, int(shardNameCap))
	streamName    = make([]byte, int(streamNameCap))

	streamRecv streamRecvFunc
	asyncRecv  asyncRecvFunc
)

//export __shard_client
func __shard_client() uint32 {
	for i, p := range []unsafe.Pointer{
		unsafe.Pointer(&val),
		unsafe.Pointer(&componentNameCap),
		unsafe.Pointer(&componentNameLen),
		unsafe.Pointer(&componentName[0]),
		unsafe.Pointer(&shardNameCap),
		unsafe.Pointer(&shardNameLen),
		unsafe.Pointer(&shardName[0]),
		unsafe.Pointer(&dataCap),
		unsafe.Pointer(&dataLen),
		unsafe.Pointer(&data[0]),
		unsafe.Pointer(&errCap),
		unsafe.Pointer(&errLen),
		unsafe.Pointer(&err[0]),
		unsafe.Pointer(&streamNameCap),
		unsafe.Pointer(&streamNameLen),
		unsafe.Pointer(&streamName[0]),
	} {
		meta[i] = uint32(uintptr(p))
	}
	return uint32(uintptr(unsafe.Pointer(&meta[0])))
}

//export __shard_client_stream_recv
func __shard_client_stream_recv() {
	streamRecv(getStreamName(), getData(), getVal())
}

//export __shard_client_async_recv
func __shard_client_async_recv() {
	asyncRecv(getStreamName(), getData(), getVal(), getErr())
}

func setComponentName(name []byte) {
	copy(componentName[:len(name)], name)
	componentNameLen = uint32(len(name))
}

func getComponentName() []byte {
	return componentName[:componentNameLen]
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
	// if dataLen > dataCap {
	// 	res := make([]byte, dataLen)
	// 	var i uint32
	// 	for i = 0; i < dataLen; {
	// 		copy(res[i*dataCap:], data[:min(dataCap, dataLen-(i*dataCap))])
	// 		i += dataLen
	// 		_buffer_continue()
	// 	}
	// 	return res
	// }
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
func _read()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_read_local
func _read_local()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_apply
func _apply()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_async_read
func _async_read()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_async_read_local
func _async_read_local()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_async_apply
func _async_apply()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_open
func _streamOpen()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_open_local
func _streamOpenLocal()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_send
func _streamSend()

//go:wasm-module pantopic/wazero-shard-client
//export __shard_client_stream_close
func _streamClose()

var _ = __shard_client
var _ = __shard_client_stream_recv
var _ = __shard_client_async_recv
var _ = getShardName
var _ = setData
