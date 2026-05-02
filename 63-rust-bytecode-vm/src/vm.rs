use crate::bytecode::{Bytecode, VmError};
use crate::call_stack::{CallStack, StackFrame};
use crate::instruction::*;

pub struct VirtualMachine {
    bytecode: Bytecode,
    pc: usize,
    operand_stack: Vec<i32>,
    call_stack: CallStack,
    running: bool,
}

impl VirtualMachine {
    pub fn new(bytecode: Bytecode) -> Self {
        Self {
            bytecode,
            pc: 0,
            operand_stack: Vec::new(),
            call_stack: CallStack::new(),
            running: true,
        }
    }

    pub fn run(&mut self) -> Result<(), VmError> {
        while self.running {
            if self.pc >= self.bytecode.len() {
                return Err(VmError::InvalidAddress(self.pc));
            }
            self.execute_one()?;
        }
        Ok(())
    }

    fn execute_one(&mut self) -> Result<(), VmError> {
        let opcode = self
            .bytecode
            .read_u8(self.pc)
            .ok_or(VmError::InvalidAddress(self.pc))?;

        match opcode {
            OP_ICONST => {
                let val = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))?;
                self.operand_stack.push(val);
                self.pc += 5;
            }
            OP_ADD => {
                let b = self.pop_operand()?;
                let a = self.pop_operand()?;
                self.operand_stack.push(a + b);
                self.pc += 1;
            }
            OP_SUB => {
                let b = self.pop_operand()?;
                let a = self.pop_operand()?;
                self.operand_stack.push(a - b);
                self.pc += 1;
            }
            OP_MUL => {
                let b = self.pop_operand()?;
                let a = self.pop_operand()?;
                self.operand_stack.push(a * b);
                self.pc += 1;
            }
            OP_DIV => {
                let b = self.pop_operand()?;
                let a = self.pop_operand()?;
                if b == 0 {
                    return Err(VmError::DivisionByZero);
                }
                self.operand_stack.push(a / b);
                self.pc += 1;
            }
            OP_MOD => {
                let b = self.pop_operand()?;
                let a = self.pop_operand()?;
                if b == 0 {
                    return Err(VmError::DivisionByZero);
                }
                self.operand_stack.push(a % b);
                self.pc += 1;
            }
            OP_LOAD => {
                let var_idx = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))? as usize;
                let val = self.load_local(var_idx)?;
                self.operand_stack.push(val);
                self.pc += 5;
            }
            OP_STORE => {
                let var_idx = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))? as usize;
                let val = self.pop_operand()?;
                self.store_local(var_idx, val)?;
                self.pc += 5;
            }
            OP_JMP => {
                let addr = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))? as usize;
                self.validate_address(addr)?;
                self.pc = addr;
            }
            OP_JZ => {
                let addr = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))? as usize;
                self.validate_address(addr)?;
                let val = self.pop_operand()?;
                if val == 0 {
                    self.pc = addr;
                } else {
                    self.pc += 5;
                }
            }
            OP_JNZ => {
                let addr = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))? as usize;
                self.validate_address(addr)?;
                let val = self.pop_operand()?;
                if val != 0 {
                    self.pc = addr;
                } else {
                    self.pc += 5;
                }
            }
            OP_CALL => {
                let addr = self
                    .bytecode
                    .read_i32(self.pc + 1)
                    .ok_or(VmError::InvalidAddress(self.pc + 1))? as usize;
                self.validate_address(addr)?;
                let return_address = self.pc + 5;
                let frame = StackFrame::new(return_address);
                self.call_stack.push(frame)?;
                self.pc = addr;
            }
            OP_RET => {
                if self.call_stack.is_empty() {
                    self.running = false;
                } else {
                    let frame = self.call_stack.pop().unwrap();
                    self.pc = frame.return_address();
                }
            }
            OP_PRINT => {
                let val = self.pop_operand()?;
                println!("{}", val);
                self.pc += 1;
            }
            OP_HALT => {
                self.running = false;
            }
            _ => {
                return Err(VmError::InvalidOpcode(opcode));
            }
        }

        Ok(())
    }

    fn pop_operand(&mut self) -> Result<i32, VmError> {
        self.operand_stack.pop().ok_or(VmError::StackUnderflow)
    }

    fn load_local(&self, idx: usize) -> Result<i32, VmError> {
        if let Some(frame) = self.call_stack.current() {
            frame.load_local(idx)
        } else {
            Err(VmError::InvalidAddress(idx))
        }
    }

    fn store_local(&mut self, idx: usize, value: i32) -> Result<(), VmError> {
        if let Some(frame) = self.call_stack.current_mut() {
            frame.store_local(idx, value)
        } else {
            Err(VmError::InvalidAddress(idx))
        }
    }

    fn validate_address(&self, addr: usize) -> Result<(), VmError> {
        if addr >= self.bytecode.len() {
            return Err(VmError::InvalidAddress(addr));
        }
        Ok(())
    }

    pub fn operand_stack_len(&self) -> usize {
        self.operand_stack.len()
    }
}
