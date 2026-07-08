pub const StreamRecvFn = *const fn (shard_name: []const u8, stream_name: []const u8, data: []const u8, val: u64) void;

extern fn __shard_client() usize;
extern fn __shard_client_stream_recv() void;
extern fn __shard_client_read() void;
extern fn __shard_client_read_local() void;
extern fn __shard_client_apply() void;
extern fn __shard_client_stream_open() void;
extern fn __shard_client_stream_open_local() void;
extern fn __shard_client_stream_send() void;
extern fn __shard_client_stream_close() void;

pub fn init() usize {
    return __shard_client();
}

pub fn streamRecv() void {
    __shard_client_stream_recv();
}

pub fn read() void {
    __shard_client_read();
}

pub fn readLocal() void {
    __shard_client_read_local();
}

pub fn apply() void {
    __shard_client_apply();
}

pub fn streamOpen() void {
    __shard_client_stream_open();
}

pub fn streamOpenLocal() void {
    __shard_client_stream_open_local();
}

pub fn streamSend() void {
    __shard_client_stream_send();
}

pub fn streamClose() void {
    __shard_client_stream_close();
}
