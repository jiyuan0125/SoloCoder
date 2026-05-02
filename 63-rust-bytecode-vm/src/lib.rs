pub mod bytecode;
pub mod call_stack;
pub mod instruction;
pub mod vm;

pub use bytecode::{Bytecode, VmError};
pub use vm::VirtualMachine;
