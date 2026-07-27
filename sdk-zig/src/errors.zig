pub const StreamRecvError = error{
    StreamRecvAlreadyRegistered,
    StreamRecvNotRegistered,
};

// Host indicates the host module reported an error; the message is available
// via sdk.HostError().
pub const HostError = error{
    Host,
};

pub const Error = StreamRecvError || HostError;
pub const ErrStreamRecvAlreadyRegistered = StreamRecvError.StreamRecvAlreadyRegistered;
pub const ErrStreamRecvNotRegistered = StreamRecvError.StreamRecvNotRegistered;
pub const ErrHost = HostError.Host;
