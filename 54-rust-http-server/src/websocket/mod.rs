pub mod frame;

pub use frame::{Frame, OpCode, FrameParser, ParseFrameResult, encode_frame, compute_accept_key, WS_GUID};

pub const MAX_PAYLOAD_SIZE: usize = 4 * 1024 * 1024;
pub const PING_INTERVAL_SECS: u64 = 30;
pub const PONG_TIMEOUT_SECS: u64 = 60;
