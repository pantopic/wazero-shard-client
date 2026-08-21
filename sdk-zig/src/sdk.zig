const std = @import("std");
const abi = @import("abi.zig");
const errors = @import("errors.zig");

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
    component_name: []const u8 = "",
    shard_name: []const u8 = "",

    pub fn init(component_name: []const u8, shard_name: []const u8) Client {
        return .{ .component_name = component_name, .shard_name = shard_name };
    }

    pub fn Read(self: Client, query: []const u8, stale: bool) ReadResult {
        abi.setComponentName(self.component_name);
        abi.setShardName(self.shard_name);
        abi.setData(query);
        if (stale) {
            abi.read_local();
        } else {
            abi.read();
        }
        return .{
            .val = abi.getVal(),
            .res = abi.getData(),
            .err = abi.getErr(),
        };
    }

    pub fn Apply(self: Client, cmd: []const u8) ApplyResult {
        abi.setComponentName(self.component_name);
        abi.setShardName(self.shard_name);
        abi.setData(cmd);
        abi.apply();
        return .{
            .val = abi.getVal(),
            .res = abi.getData(),
            .err = abi.getErr(),
        };
    }

    pub fn AsyncRead(self: Client, query: []const u8, name: []const u8, stale: bool) errors.Error!void {
        abi.setComponentName(self.component_name);
        abi.setShardName(self.shard_name);
        abi.setStreamName(name);
        abi.setData(query);
        if (stale) {
            abi.async_read_local();
        } else {
            abi.async_read();
        }
    }

    pub fn AsyncApply(self: Client, cmd: []const u8, name: []const u8) errors.Error!void {
        abi.setComponentName(self.component_name);
        abi.setShardName(self.shard_name);
        abi.setStreamName(name);
        abi.setData(cmd);
        abi.async_apply();
    }

    pub fn StreamOpen(self: Client, name: []const u8) errors.Error!void {
        if (abi.stream_recv == null) {
            return errors.ErrStreamRecvNotRegistered;
        }
        abi.setComponentName(self.component_name);
        abi.setShardName(self.shard_name);
        abi.setStreamName(name);
        abi.stream_open();
        return checkErr();
    }

    pub fn StreamOpenLocal(self: Client, name: []const u8) errors.Error!void {
        if (abi.stream_recv == null) {
            return errors.ErrStreamRecvNotRegistered;
        }
        abi.setComponentName(self.component_name);
        abi.setShardName(self.shard_name);
        abi.setStreamName(name);
        abi.stream_open_local();
        return checkErr();
    }

    pub fn StreamSend(self: Client, name: []const u8, data: []const u8) errors.Error!void {
        _ = self;
        abi.setStreamName(name);
        abi.setData(data);
        abi.stream_send();
        return checkErr();
    }

    pub fn StreamClose(self: Client, name: []const u8) errors.Error!void {
        _ = self;
        abi.setStreamName(name);
        abi.stream_close();
        return checkErr();
    }
};

pub fn New(component_name: []const u8, shard_name: []const u8) Client {
    return Client.init(component_name, shard_name);
}

pub fn RegisterStreamRecv(callback: abi.StreamRecvFn) errors.Error!void {
    if (abi.stream_recv != null) {
        return errors.ErrStreamRecvAlreadyRegistered;
    }
    abi.stream_recv = callback;
}

pub fn RegisterAsyncRecv(callback: abi.AsyncRecvFn) errors.Error!void {
    if (abi.async_recv != null) {
        return errors.ErrAsyncRecvAlreadyRegistered;
    }
    abi.async_recv = callback;
}

pub fn HostError() ?[]const u8 {
    return abi.getErr();
}

fn checkErr() errors.Error!void {
    if (abi.getErr() != null) {
        return errors.ErrHost;
    }
}

test "client initialization preserves component and shard name" {
    const client = Client.init("counter", "alpha");
    try std.testing.expectEqualStrings(client.component_name, "counter");
    try std.testing.expectEqualStrings(client.shard_name, "alpha");
}
