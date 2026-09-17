# Zig SDK

This directory contains a Zig port of the guest SDK for the Wazero shard client.

## Layout

- src/abi.zig - Low-level extern declarations for the shard client ABI.
- src/errors.zig - Error values mirrored from the Go SDK.
- src/sdk.zig - Public client API and state handling.
- src/root.zig - Root module exports.

## Example

```zig
const shard_client = @import("shard_client");

pub fn main() void {}

pub fn get() u64 {
    const client = shard_client.New("alpha");
    const result = client.Read("GET /test", false);
    _ = result;
    return 0;
}
```

## Build

```sh
zig build test
```
