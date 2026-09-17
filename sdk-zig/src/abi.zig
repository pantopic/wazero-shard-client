pub const StreamRecvFn = *const fn (name: []const u8, data: []const u8, val: u64) void;
pub const AsyncRecvFn = *const fn (name: []const u8, data: []const u8, val: u64, err: ?[]const u8) void;

const component_name_cap = 64;
const shard_name_cap = 64;
const data_cap = 2 << 20; // 2 MiB
const err_cap = 1024;
const stream_name_cap = 64;

var meta: [16]u32 = .{0} ** 16;
var val: u64 = 0;
var componentNameCap: u32 = component_name_cap;
var componentNameLen: u32 = 0;
var shardNameCap: u32 = shard_name_cap;
var shardNameLen: u32 = 0;
var dataCap: u32 = data_cap;
var dataLen: u32 = 0;
var errCap: u32 = err_cap;
var errLen: u32 = 0;
var streamNameCap: u32 = stream_name_cap;
var streamNameLen: u32 = 0;

var componentName: [component_name_cap]u8 = .{0} ** component_name_cap;
var shardName: [shard_name_cap]u8 = .{0} ** shard_name_cap;
var data: [data_cap]u8 = .{0} ** data_cap;
var err: [err_cap]u8 = .{0} ** err_cap;
var streamName: [stream_name_cap]u8 = .{0} ** stream_name_cap;

pub var stream_recv: ?StreamRecvFn = null;
pub var async_recv: ?AsyncRecvFn = null;

export fn __shard_client() u32 {
    meta[0] = @intCast(@intFromPtr(&val));
    meta[1] = @intCast(@intFromPtr(&componentNameCap));
    meta[2] = @intCast(@intFromPtr(&componentNameLen));
    meta[3] = @intCast(@intFromPtr(&componentName[0]));
    meta[4] = @intCast(@intFromPtr(&shardNameCap));
    meta[5] = @intCast(@intFromPtr(&shardNameLen));
    meta[6] = @intCast(@intFromPtr(&shardName[0]));
    meta[7] = @intCast(@intFromPtr(&dataCap));
    meta[8] = @intCast(@intFromPtr(&dataLen));
    meta[9] = @intCast(@intFromPtr(&data[0]));
    meta[10] = @intCast(@intFromPtr(&errCap));
    meta[11] = @intCast(@intFromPtr(&errLen));
    meta[12] = @intCast(@intFromPtr(&err[0]));
    meta[13] = @intCast(@intFromPtr(&streamNameCap));
    meta[14] = @intCast(@intFromPtr(&streamNameLen));
    meta[15] = @intCast(@intFromPtr(&streamName[0]));
    return @intCast(@intFromPtr(&meta[0]));
}

export fn __shard_client_stream_recv() void {
    stream_recv.?(getStreamName(), getData(), getVal());
}

export fn __shard_client_async_recv() void {
    async_recv.?(getStreamName(), getData(), getVal(), getErr());
}

pub fn setComponentName(name: []const u8) void {
    const len = @min(name.len, componentName.len);
    @memcpy(componentName[0..len], name[0..len]);
    componentNameLen = @intCast(len);
}

pub fn getComponentName() []const u8 {
    return componentName[0..componentNameLen];
}

pub fn setShardName(name: []const u8) void {
    const len = @min(name.len, shardName.len);
    @memcpy(shardName[0..len], name[0..len]);
    shardNameLen = @intCast(len);
}

pub fn getShardName() []const u8 {
    return shardName[0..shardNameLen];
}

pub fn setData(v: []const u8) void {
    const len = @min(v.len, data.len);
    @memcpy(data[0..len], v[0..len]);
    dataLen = @intCast(len);
}

pub fn getData() []u8 {
    return data[0..dataLen];
}

pub fn setErr(e: []const u8) void {
    const len = @min(e.len, err.len);
    @memcpy(err[0..len], e[0..len]);
    errLen = @intCast(len);
}

pub fn getErr() ?[]const u8 {
    if (errLen == 0) {
        return null;
    }
    return err[0..errLen];
}

pub fn getVal() u64 {
    return val;
}

pub fn setStreamName(name: []const u8) void {
    const len = @min(name.len, streamName.len);
    @memcpy(streamName[0..len], name[0..len]);
    streamNameLen = @intCast(len);
}

pub fn getStreamName() []const u8 {
    return streamName[0..streamNameLen];
}

extern "pantopic/wazero-shard-client" fn __shard_client_read() void;
extern "pantopic/wazero-shard-client" fn __shard_client_read_local() void;
extern "pantopic/wazero-shard-client" fn __shard_client_apply() void;
extern "pantopic/wazero-shard-client" fn __shard_client_async_read() void;
extern "pantopic/wazero-shard-client" fn __shard_client_async_read_local() void;
extern "pantopic/wazero-shard-client" fn __shard_client_async_apply() void;
extern "pantopic/wazero-shard-client" fn __shard_client_stream_open() void;
extern "pantopic/wazero-shard-client" fn __shard_client_stream_open_local() void;
extern "pantopic/wazero-shard-client" fn __shard_client_stream_send() void;
extern "pantopic/wazero-shard-client" fn __shard_client_stream_close() void;

pub fn read() void {
    __shard_client_read();
}

pub fn read_local() void {
    __shard_client_read_local();
}

pub fn apply() void {
    __shard_client_apply();
}
pub fn async_read() void {
    __shard_client_async_read();
}

pub fn async_read_local() void {
    __shard_client_async_read_local();
}

pub fn async_apply() void {
    __shard_client_async_apply();
}

pub fn stream_open() void {
    __shard_client_stream_open();
}

pub fn stream_open_local() void {
    __shard_client_stream_open_local();
}

pub fn stream_send() void {
    __shard_client_stream_send();
}

pub fn stream_close() void {
    __shard_client_stream_close();
}
