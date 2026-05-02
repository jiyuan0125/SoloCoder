use sha1::{Sha1, Digest};
use base64::engine::general_purpose;
use base64::Engine as _;

pub const MAX_PAYLOAD_SIZE: usize = 4 * 1024 * 1024;
pub const PING_INTERVAL_SECS: u64 = 30;
pub const PONG_TIMEOUT_SECS: u64 = 60;

pub const WS_MAGIC_GUID: &str = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11";

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
    pub fn from_u8(byte: u8) -> Option<Self> {
        match byte {
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
    pub fn new(opcode: OpCode, payload: Vec<u8>) -> Self {
        Frame {
            fin: true,
            opcode,
            payload,
        }
    }

    pub fn text(payload: String) -> Self {
        Frame::new(OpCode::Text, payload.into_bytes())
    }

    pub fn binary(payload: Vec<u8>) -> Self {
        Frame::new(OpCode::Binary, payload)
    }

    pub fn ping(payload: Vec<u8>) -> Self {
        Frame::new(OpCode::Ping, payload)
    }

    pub fn pong(payload: Vec<u8>) -> Self {
        Frame::new(OpCode::Pong, payload)
    }

    pub fn close(code: u16, reason: &str) -> Self {
        let mut payload = Vec::new();
        payload.extend_from_slice(&code.to_be_bytes());
        payload.extend_from_slice(reason.as_bytes());
        Frame::new(OpCode::Close, payload)
    }

    pub fn close_code(&self) -> Option<u16> {
        if self.opcode != OpCode::Close || self.payload.len() < 2 {
            return None;
        }
        Some(u16::from_be_bytes([self.payload[0], self.payload[1]]))
    }

    pub fn encode(&self) -> Vec<u8> {
        let mut result = Vec::new();
        
        let mut first_byte: u8 = 0;
        if self.fin {
            first_byte |= 0x80;
        }
        first_byte |= self.opcode as u8;
        result.push(first_byte);
        
        let payload_len = self.payload.len();
        if payload_len <= 125 {
            result.push(payload_len as u8);
        } else if payload_len <= 65535 {
            result.push(126);
            result.extend_from_slice(&(payload_len as u16).to_be_bytes());
        } else {
            result.push(127);
            result.extend_from_slice(&(payload_len as u64).to_be_bytes());
        }
        
        result.extend_from_slice(&self.payload);
        
        result
    }
}

pub enum FrameParseError {
    Incomplete,
    InvalidOpcode,
    PayloadTooLarge,
    InvalidFrame,
}

pub fn parse_frame(buffer: &[u8], is_server: bool) -> Result<(Frame, usize), FrameParseError> {
    if buffer.len() < 2 {
        return Err(FrameParseError::Incomplete);
    }
    
    let first_byte = buffer[0];
    let second_byte = buffer[1];
    
    let fin = (first_byte & 0x80) != 0;
    let opcode = OpCode::from_u8(first_byte & 0x0F).ok_or(FrameParseError::InvalidOpcode)?;
    
    let masked = (second_byte & 0x80) != 0;
    let mut payload_len = (second_byte & 0x7F) as usize;
    
    let mut offset = 2;
    
    if payload_len == 126 {
        if buffer.len() < offset + 2 {
            return Err(FrameParseError::Incomplete);
        }
        payload_len = u16::from_be_bytes([buffer[offset], buffer[offset + 1]]) as usize;
        offset += 2;
    } else if payload_len == 127 {
        if buffer.len() < offset + 8 {
            return Err(FrameParseError::Incomplete);
        }
        payload_len = u64::from_be_bytes([
            buffer[offset], buffer[offset + 1], buffer[offset + 2], buffer[offset + 3],
            buffer[offset + 4], buffer[offset + 5], buffer[offset + 6], buffer[offset + 7],
        ]) as usize;
        offset += 8;
    }
    
    if opcode.is_control() && payload_len > 125 {
        return Err(FrameParseError::InvalidFrame);
    }
    
    if payload_len > MAX_PAYLOAD_SIZE {
        return Err(FrameParseError::PayloadTooLarge);
    }
    
    let mut masking_key = [0u8; 4];
    if masked {
        if is_server {
            if buffer.len() < offset + 4 {
                return Err(FrameParseError::Incomplete);
            }
            masking_key.copy_from_slice(&buffer[offset..offset + 4]);
            offset += 4;
        } else {
            return Err(FrameParseError::InvalidFrame);
        }
    }
    
    if buffer.len() < offset + payload_len {
        return Err(FrameParseError::Incomplete);
    }
    
    let mut payload = buffer[offset..offset + payload_len].to_vec();
    
    if masked {
        for i in 0..payload.len() {
            payload[i] ^= masking_key[i % 4];
        }
    }
    
    let frame = Frame {
        fin,
        opcode,
        payload,
    };
    
    Ok((frame, offset + payload_len))
}

pub fn compute_accept_key(sec_websocket_key: &str) -> String {
    let mut hasher = Sha1::new();
    hasher.update(sec_websocket_key.as_bytes());
    hasher.update(WS_MAGIC_GUID.as_bytes());
    let result = hasher.finalize();
    general_purpose::STANDARD.encode(result)
}

pub fn create_handshake_response(sec_websocket_key: &str) -> Vec<u8> {
    let accept_key = compute_accept_key(sec_websocket_key);
    format!(
        "HTTP/1.1 101 Switching Protocols\r\n\
         Upgrade: websocket\r\n\
         Connection: Upgrade\r\n\
         Sec-WebSocket-Accept: {}\r\n\
         \r\n",
        accept_key
    ).into_bytes()
}
