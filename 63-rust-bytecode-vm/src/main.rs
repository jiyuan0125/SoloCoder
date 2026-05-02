use std::env;
use std::process;

use bytecode_vm::{Bytecode, VirtualMachine, VmError};

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        eprintln!("Usage: {} <bytecode_file>", args[0]);
        process::exit(1);
    }

    let bytecode_path = &args[1];

    let bytecode = match Bytecode::from_file(bytecode_path) {
        Ok(bc) => bc,
        Err(e) => {
            eprintln!("Error loading bytecode: {}", e);
            process::exit(1);
        }
    };

    let mut vm = VirtualMachine::new(bytecode);

    let result = vm.run();

    match result {
        Ok(()) => {
            let stack_len = vm.operand_stack_len();
            println!("Execution completed. Remaining stack elements: {}", stack_len);
            if stack_len != 0 {
                eprintln!("Warning: Stack is not empty after execution");
            }
        }
        Err(e) => {
            eprintln!("Runtime error: {}", e);
            match e {
                VmError::DivisionByZero => process::exit(2),
                VmError::InvalidAddress(_) => process::exit(3),
                VmError::StackOverflow => process::exit(4),
                VmError::StackUnderflow => process::exit(5),
                VmError::InvalidOpcode(_) => process::exit(6),
                VmError::Halted => process::exit(0),
            }
        }
    }
}
