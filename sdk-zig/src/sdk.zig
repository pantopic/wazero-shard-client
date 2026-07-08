const std = @import("std");
const abi = @import("abi.zig");
const errors = @import("errors.zig");

const max_shard_name_len = 256;
const max_data_len = 2 << 20;
const max_err_len = 1024;
const max_stream_name_len = 64;

const SharedState = struct {
    val: u64 = 0,
    shard_name: [max_shard_name_len]u8 = .{0} ** max_shard_name_len,
    shard_name_len: u32 = 0,
    data: [max_data_len]u8 = .{0} ** max_data_len,
    data_len: u32 = 0,
    err: [max_err_len]u8 = .{0} ** max_err_len,
    err_len: u32 = 0,
    stream_name: [max_stream_name_len]u8 = .{0} ** max_stream_name_len,
    stream_name_len: u32 = 0,
};

var shared_state = SharedState{};
var stream_recv: ?abi.StreamRecvFn = null;

pub const ReadResult = struct {
    val: u64 = 0,
    res: []u8 = &.{},
    err: ?[]const u8 = null,
};

pub const ApplyResult = struct {
    val: u64 = 0,
    res: []u8 = &.{},
    err: ?[]const u8 = null,
};

pub const Client = struct {
    shard_name: []const u8 = "",

    pub fn init(name: []const u8) Client {
        return .{ .shard_name = name };
    }

    pub fn Read(self: Client, query: []const u8, stale: bool) ReadResult {
        setShardName(self.shard_name);
        setData(query);
        if (stale) {
            abi.readLocal();
        } else {
            abi.read();
        }
        return .{
            .val = getVal(),
            .res = getData(),
            .err = getErr(),
        };
    }

    pub fn Apply(self: Client, cmd: []const u8) ApplyResult {
        setShardName(self.shard_name);
        setData(cmd);
        abi.apply();
        return .{
            .val = getVal(),
            .res = getData(),
            .err = getErr(),
        };
    }

    pub fn StreamOpen(self: Client, name: []const u8) !void {
        if (stream_recv == null) {
            return errors.ErrStreamRecvNotRegistered;
        }
        setShardName(self.shard_name);
        setStreamName(name);
        abi.streamOpen();
    }

    pub fn StreamOpenLocal(self: Client, name: []const u8) !void {
        if (stream_recv == null) {
            return errors.ErrStreamRecvNotRegistered;
        }
        setShardName(self.shard_name);
        setStreamName(name);
        abi.streamOpenLocal();
    }

    pub fn StreamSend(self: Client, name: []const u8, data: []const u8) !void {
        setShardName(self.shard_name);
        setStreamName(name);
        setData(data);
        abi.streamSend();
        return {};
    }

    pub fn StreamClose(self: Client, name: []const u8) !void {
        setShardName(self.shard_name);
        setStreamName(name);
        abi.streamClose();
        return {};
    }
};

pub fn New(name: []const u8) Client {
    return Client.init(name);
}

pub fn RegisterStreamRecv(callback: abi.StreamRecvFn) !void {
    if (stream_recv != null) {
        return errors.ErrStreamRecvAlreadyRegistered;
    }
    stream_recv = callback;
    return {};
}

pub fn DispatchStreamRecv() void {
    if (stream_recv) |recv| {
        recv(getShardName(), getStreamName(), getData(), getVal());
    }
}

fn setShardName(name: []const u8) void {
    const len = @min(name.len, shared_state.shard_name.len);
    @memcpy(shared_state.shard_name[0..len], name[0..len]);
    shared_state.shard_name_len = @as(u32, @intCast(len));
}

fn getShardName() []const u8 {
    const len: usize = @as(usize, @intCast(shared_state.shard_name_len));
    return shared_state.shard_name[0..len];
}

fn setData(v: []const u8) void {
    const len = @min(v.len, shared_state.data.len);
    @memcpy(shared_state.data[0..len], v[0..len]);
    shared_state.data_len = @as(u32, @intCast(len));
}

fn getData() []u8 {
    const len: usize = @as(usize, @intCast(shared_state.data_len));
    return shared_state.data[0..len];
}

fn setErr(e: []const u8) void {
    const len = @min(e.len, shared_state.err.len);
    @memcpy(shared_state.err[0..len], e[0..len]);
    shared_state.err_len = @as(u32, @intCast(len));
}

fn getErr() ?[]const u8 {
    if (shared_state.err_len == 0) {
        return null;
    }
    const len: usize = @as(usize, @intCast(shared_state.err_len));
    return shared_state.err[0..len];
}

fn getVal() u64 {
    return shared_state.val;
}

fn getStreamName() []const u8 {
    const len: usize = @as(usize, @intCast(shared_state.stream_name_len));
    return shared_state.stream_name[0..len];
}

fn setStreamName(name: []const u8) void {
    const len = @min(name.len, shared_state.stream_name.len);
    @memcpy(shared_state.stream_name[0..len], name[0..len]);
    shared_state.stream_name_len = @as(u32, @intCast(len));
}

test "client initialization preserves shard name" {
    const client = Client.init("alpha");
    try std.testing.expectEqualStrings(client.shard_name, "alpha");
}
