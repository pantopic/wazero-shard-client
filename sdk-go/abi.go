package shard_client

import (
	"unsafe"
)

var (
	meta         = make([]uint32, 10)
	val          uint64
	shardNameCap uint32 = 256
	shardNameLen uint32
	dataCap      uint32 = 2 << 20 // 2 MiB
	dataLen      uint32
	errCap       uint32 = 1024
	errLen       uint32

	shardName = make([]byte, int(shardNameCap))
	data      = make([]byte, int(dataCap))
	err       = make([]byte, int(errCap))
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
	return uint32(uintptr(unsafe.Pointer(&meta[0])))
}

func setShardName(v string) {
	b := []byte(v)
	copy(shardName[:len(b)], b)
	shardNameLen = uint32(len(v))
}

func getShardName() string {
	return string(shardName[:shardNameLen])
}

func setData(v []byte) {
	copy(data[:len(v)], v)
	dataLen = uint32(len(v))
}

func getData() []byte {
	return data[:dataLen]
}

func setErr(v string) {
	b := []byte(v)
	copy(err[:len(b)], b)
	errLen = uint32(len(b))
}

func getErr() (e error) {
	if errLen > 0 {
		e = strErr(string(err[:errLen]))
	}
	return
}

type strErr string

func (e strErr) Error() string {
	return string(e)
}

func getVal() uint64 {
	return val
}

//go:wasm-module pantopic/wazero-shard-client
//export Read
func read()

//go:wasm-module pantopic/wazero-shard-client
//export ReadLocal
func readlocal()

//go:wasm-module pantopic/wazero-shard-client
//export Apply
func apply()

var _ = __shard_client
var _ = setShardName
var _ = getShardName
var _ = setData
var _ = getData
var _ = setErr
var _ = getErr
