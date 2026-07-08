pub const StreamRecvError = error{
    StreamRecvAlreadyRegistered,
    StreamRecvNotRegistered,
};

pub const Error = StreamRecvError;
pub const ErrStreamRecvAlreadyRegistered = StreamRecvError.StreamRecvAlreadyRegistered;
pub const ErrStreamRecvNotRegistered = StreamRecvError.StreamRecvNotRegistered;
