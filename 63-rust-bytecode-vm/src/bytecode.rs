use std::fs;
use std::io;

pub struct Bytecode {
    pub data: Vec<u8>,
}

impl Bytecode {
    pub fn from_file(path: &str) -> io::Result<Self> {
        let data = fs::read(path)?;
        Ok(Self { data })
    }

    pub fn len(&self) -> usize {
        self.data.len()
    }

    pub fn is_empty(&self) -> bool {
        self.data.is_empty()
    }

    pub fn read_u8(&self, offset: usize) -> Option<u8> {
        self.data.get(offset).copied()
    }

    pub fn read_i32(&self, offset: usize) -> Option<i32> {
        if offset + 4 > self.data.len() {
            return None;
        }
        let bytes: [u8; 4] = self.data[offset..offset + 4].try_into().ok()?;
        Some(i32::from_le_bytes(bytes))
    }
}

#[derive(Debug, PartialEq, Eq)]
pub enum VmError {
    DivisionByZero,
    InvalidAddress(usize),
    StackOverflow,
    StackUnderflow,
    InvalidOpcode(u8),
    Halted,
}

impl std::fmt::Display for VmError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            VmError::DivisionByZero => write!(f, "division by zero"),
            VmError::InvalidAddress(addr) => write!(f, "invalid address: {}", addr),
            VmError::StackOverflow => write!(f, "call stack overflow"),
            VmError::StackUnderflow => write!(f, "operand stack underflow"),
            VmError::InvalidOpcode(op) => write!(f, "invalid opcode: 0x{:02x}", op),
            VmError::Halted => write!(f, "vm halted"),
        }
    }
}

impl std::error::Error for VmError {}
