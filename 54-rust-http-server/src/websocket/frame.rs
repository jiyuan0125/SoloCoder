#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OpCode {
    Continuation = 0x0,
    Text = 0x1,
    Binary = 0x2,
    Close = 0x8,
    Ping = 0x9,
    Pong = 0xA,
}

impl OpCode {
    pub fn from_u8(value: u8) -> Option<Self> {
        match value {
            0x0 => Some(OpCode::Continuation),
            0x1 => Some(OpCode::Text),
            0x2 => Some(OpCode::Binary),
            0x8 => Some(OpCode::Close),
            0x9 => Some(OpCode::Ping),
            0xA => Some(OpCode::Pong),
            _ => None,
        }
    }

    pub fn is_control(&self) -> bool {
        matches!(self, OpCode::Close | OpCode::Ping | OpCode::Pong)
    }
}

#[derive(Debug, Clone)]
pub struct Frame {
    pub fin: bool,
    pub opcode: OpCode,
    pub payload: Vec<u8>,
}

impl Frame {
    pub fn new(fin: bool, opcode: OpCode, payload: Vec<u8>) -> Self {
        Frame { fin, opcode, payload }
    }

    pub fn text(payload: String) -> Self {
        Frame::new(true, OpCode::Text, payload.into_bytes())
    }

    pub fn binary(payload: Vec<u8>) -> Self {
        Frame::new(true, OpCode::Binary, payload)
    }

    pub fn ping(payload: Vec<u8>) -> Self {
        Frame::new(true, OpCode::Ping, payload)
    }

    pub fn pong(payload: Vec<u8>) -> Self {
        Frame::new(true, OpCode::Pong, payload)
    }

    pub fn close(code: u16, reason: Option<&str>) -> Self {
        let mut payload = Vec::new();
        payload.extend_from_slice(&code.to_be_bytes());
        if let Some(reason) = reason {
            payload.extend_from_slice(reason.as_bytes());
        }
        Frame::new(true, OpCode::Close, payload)
    }

    pub fn close_code(&self) -> Option<u16> {
        if self.payload.len() >= 2 {
            Some(u16::from_be_bytes([self.payload[0], self.payload[1]]))
        } else {
            None
        }
    }
}

pub struct FrameParser {
    buffer: Vec<u8>,
}

pub enum ParseFrameResult {
    Complete(Frame),
    Partial,
    Error,
    TooLarge,
}

impl FrameParser {
    pub fn new() -> Self {
        FrameParser { buffer: Vec::new() }
    }

    pub fn consume(&mut self, data: &[u8]) {
        self.buffer.extend_from_slice(data);
    }

    pub fn parse(&mut self, max_payload_size: usize) -> ParseFrameResult {
        if self.buffer.len() < 2 {
            return ParseFrameResult::Partial;
        }

        let first_byte = self.buffer[0];
        let second_byte = self.buffer[1];

        let fin = (first_byte & 0x80) != 0;
        let opcode = match OpCode::from_u8(first_byte & 0x0F) {
            Some(op) => op,
            None => return ParseFrameResult::Error,
        };

        let masked = (second_byte & 0x80) != 0;
        let mut payload_len = (second_byte & 0x7F) as usize;

        let mut header_size = 2;

        if payload_len == 126 {
            if self.buffer.len() < 4 {
                return ParseFrameResult::Partial;
            }
            payload_len = u16::from_be_bytes([self.buffer[2], self.buffer[3]]) as usize;
            header_size = 4;
        } else if payload_len == 127 {
            if self.buffer.len() < 10 {
                return ParseFrameResult::Partial;
            }
            let len_bytes: [u8; 8] = self.buffer[2..10].try_into().unwrap();
            payload_len = u64::from_be_bytes(len_bytes) as usize;
            header_size = 10;
        }

        if payload_len > max_payload_size {
            return ParseFrameResult::TooLarge;
        }

        let mask_key = if masked {
            if self.buffer.len() < header_size + 4 {
                return ParseFrameResult::Partial;
            }
            let key: [u8; 4] = self.buffer[header_size..header_size + 4].try_into().unwrap();
            header_size += 4;
            Some(key)
        } else {
            None
        };

        if self.buffer.len() < header_size + payload_len {
            return ParseFrameResult::Partial;
        }

        let mut payload: Vec<u8> = self.buffer.drain(..header_size + payload_len).skip(header_size).collect();

        if let Some(mask) = mask_key {
            for (i, byte) in payload.iter_mut().enumerate() {
                *byte ^= mask[i % 4];
            }
        }

        ParseFrameResult::Complete(Frame { fin, opcode, payload })
    }
}

pub fn encode_frame(frame: &Frame) -> Vec<u8> {
    let mut bytes = Vec::new();

    let first_byte = if frame.fin { 0x80 } else { 0x00 } | (frame.opcode as u8);
    bytes.push(first_byte);

    let payload_len = frame.payload.len();
    if payload_len <= 125 {
        bytes.push(payload_len as u8);
    } else if payload_len <= 65535 {
        bytes.push(126);
        bytes.extend_from_slice(&(payload_len as u16).to_be_bytes());
    } else {
        bytes.push(127);
        bytes.extend_from_slice(&(payload_len as u64).to_be_bytes());
    }

    bytes.extend_from_slice(&frame.payload);
    bytes
}

pub const WS_GUID: &str = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11";

pub fn compute_accept_key(sec_ws_key: &str) -> String {
    use sha1::{Sha1, Digest};
    
    let mut hasher = Sha1::new();
    hasher.update(sec_ws_key.as_bytes());
    hasher.update(WS_GUID.as_bytes());
    let result = hasher.finalize();
    
    base64::encode(result.as_slice())
}

mod base64 {
    const BASE64_CHARS: &[u8] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

    pub fn encode(data: &[u8]) -> String {
        let mut result = Vec::new();
        let mut i = 0;

        while i < data.len() {
            let b1 = data[i];
            let b2 = if i + 1 < data.len() { data[i + 1] } else { 0 };
            let b3 = if i + 2 < data.len() { data[i + 2] } else { 0 };

            let chunk = ((b1 as u32) << 16) | ((b2 as u32) << 8) | (b3 as u32);

            result.push(BASE64_CHARS[((chunk >> 18) & 0x3F) as usize]);
            result.push(BASE64_CHARS[((chunk >> 12) & 0x3F) as usize]);

            if i + 1 < data.len() {
                result.push(BASE64_CHARS[((chunk >> 6) & 0x3F) as usize]);
            }
            if i + 2 < data.len() {
                result.push(BASE64_CHARS[(chunk & 0x3F) as usize]);
            }

            i += 3;
        }

        while result.len() % 4 != 0 {
            result.push(b'=');
        }

        String::from_utf8(result).unwrap()
    }
}
