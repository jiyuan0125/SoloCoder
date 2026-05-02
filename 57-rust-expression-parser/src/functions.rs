#[derive(Debug, Clone)]
pub enum FunctionError {
    WrongArgumentCount { expected: usize, got: usize },
    InvalidArgument(String),
}

impl std::fmt::Display for FunctionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            FunctionError::WrongArgumentCount { expected, got } => {
                write!(f, "expected {} arguments, got {}", expected, got)
            }
            FunctionError::InvalidArgument(msg) => write!(f, "{}", msg),
        }
    }
}

pub type FunctionResult = Result<f64, FunctionError>;

pub fn call_function(name: &str, args: Vec<f64>) -> FunctionResult {
    match name {
        "sqrt" => {
            if args.len() != 1 {
                return Err(FunctionError::WrongArgumentCount { expected: 1, got: args.len() });
            }
            let x = args[0];
            if x < 0.0 {
                return Err(FunctionError::InvalidArgument("sqrt of negative number".to_string()));
            }
            Ok(x.sqrt())
        }
        "abs" => {
            if args.len() != 1 {
                return Err(FunctionError::WrongArgumentCount { expected: 1, got: args.len() });
            }
            Ok(args[0].abs())
        }
        "max" => {
            if args.len() != 2 {
                return Err(FunctionError::WrongArgumentCount { expected: 2, got: args.len() });
            }
            Ok(args[0].max(args[1]))
        }
        "min" => {
            if args.len() != 2 {
                return Err(FunctionError::WrongArgumentCount { expected: 2, got: args.len() });
            }
            Ok(args[0].min(args[1]))
        }
        _ => Err(FunctionError::InvalidArgument(format!("unknown function: {}", name))),
    }
}
