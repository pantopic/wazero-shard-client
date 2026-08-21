pub const StreamRecvError = error{
    StreamRecvAlreadyRegistered,
    StreamRecvNotRegistered,
};

pub const AsyncRecvError = error{
    AsyncRecvAlreadyRegistered,
    AsyncRecvNotRegistered,
};

// Host indicates the host module reported an error; the message is available
// via sdk.HostError().
pub const HostError = error{
    Host,
};

pub const Error = StreamRecvError || AsyncRecvError || HostError;
pub const ErrStreamRecvAlreadyRegistered = StreamRecvError.StreamRecvAlreadyRegistered;
pub const ErrStreamRecvNotRegistered = StreamRecvError.StreamRecvNotRegistered;
pub const ErrAsyncRecvAlreadyRegistered = AsyncRecvError.AsyncRecvAlreadyRegistered;
pub const ErrAsyncRecvNotRegistered = AsyncRecvError.AsyncRecvNotRegistered;
pub const ErrHost = HostError.Host;
