use crate::bytecode::VmError;

pub const MAX_CALL_DEPTH: usize = 64;
pub const LOCAL_VAR_SLOTS: usize = 256;

pub struct StackFrame {
    return_address: usize,
    local_vars: [i32; LOCAL_VAR_SLOTS],
}

impl StackFrame {
    pub fn new(return_address: usize) -> Self {
        Self {
            return_address,
            local_vars: [0; LOCAL_VAR_SLOTS],
        }
    }

    pub fn return_address(&self) -> usize {
        self.return_address
    }

    pub fn load_local(&self, idx: usize) -> Result<i32, VmError> {
        if idx >= LOCAL_VAR_SLOTS {
            return Err(VmError::InvalidAddress(idx));
        }
        Ok(self.local_vars[idx])
    }

    pub fn store_local(&mut self, idx: usize, value: i32) -> Result<(), VmError> {
        if idx >= LOCAL_VAR_SLOTS {
            return Err(VmError::InvalidAddress(idx));
        }
        self.local_vars[idx] = value;
        Ok(())
    }
}

pub struct CallStack {
    frames: Vec<StackFrame>,
}

impl CallStack {
    pub fn new() -> Self {
        Self { frames: Vec::new() }
    }

    pub fn push(&mut self, frame: StackFrame) -> Result<(), VmError> {
        if self.frames.len() >= MAX_CALL_DEPTH {
            return Err(VmError::StackOverflow);
        }
        self.frames.push(frame);
        Ok(())
    }

    pub fn pop(&mut self) -> Option<StackFrame> {
        self.frames.pop()
    }

    pub fn current(&self) -> Option<&StackFrame> {
        self.frames.last()
    }

    pub fn current_mut(&mut self) -> Option<&mut StackFrame> {
        self.frames.last_mut()
    }

    pub fn depth(&self) -> usize {
        self.frames.len()
    }

    pub fn is_empty(&self) -> bool {
        self.frames.is_empty()
    }
}

impl Default for CallStack {
    fn default() -> Self {
        Self::new()
    }
}
