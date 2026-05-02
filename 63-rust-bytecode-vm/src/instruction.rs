pub const OP_ICONST: u8 = 0x01;
pub const OP_ADD: u8 = 0x02;
pub const OP_SUB: u8 = 0x03;
pub const OP_MUL: u8 = 0x04;
pub const OP_DIV: u8 = 0x05;
pub const OP_MOD: u8 = 0x06;
pub const OP_LOAD: u8 = 0x07;
pub const OP_STORE: u8 = 0x08;
pub const OP_JMP: u8 = 0x09;
pub const OP_JZ: u8 = 0x0a;
pub const OP_JNZ: u8 = 0x0b;
pub const OP_CALL: u8 = 0x0c;
pub const OP_RET: u8 = 0x0d;
pub const OP_PRINT: u8 = 0x0e;
pub const OP_HALT: u8 = 0x0f;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Instruction {
    IConst(i32),
    Add,
    Sub,
    Mul,
    Div,
    Mod,
    Load(usize),
    Store(usize),
    Jmp(usize),
    Jz(usize),
    Jnz(usize),
    Call(usize),
    Ret,
    Print,
    Halt,
}

impl Instruction {
    pub fn size(&self) -> usize {
        match self {
            Instruction::IConst(_) => 5,
            Instruction::Load(_) => 5,
            Instruction::Store(_) => 5,
            Instruction::Jmp(_) => 5,
            Instruction::Jz(_) => 5,
            Instruction::Jnz(_) => 5,
            Instruction::Call(_) => 5,
            _ => 1,
        }
    }
}

pub fn has_operand(opcode: u8) -> bool {
    matches!(opcode, OP_ICONST | OP_LOAD | OP_STORE | OP_JMP | OP_JZ | OP_JNZ | OP_CALL)
}
