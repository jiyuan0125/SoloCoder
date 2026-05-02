use crate::parser::Expr;
use crate::functions::call_function;
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub enum EvalError {
    DivisionByZero,
    UndefinedVariable(String),
    FunctionError(String),
    Other(String),
}

impl std::fmt::Display for EvalError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EvalError::DivisionByZero => write!(f, "division by zero"),
            EvalError::UndefinedVariable(name) => write!(f, "undefined variable: {}", name),
            EvalError::FunctionError(msg) => write!(f, "{}", msg),
            EvalError::Other(msg) => write!(f, "{}", msg),
        }
    }
}

#[derive(Debug, Clone)]
pub enum EvalResult {
    Value(f64),
    Assignment,
}

pub struct Evaluator {
    variables: HashMap<String, f64>,
}

impl Evaluator {
    pub fn new() -> Self {
        Evaluator {
            variables: HashMap::new(),
        }
    }

    pub fn eval(&mut self, expr: &Expr) -> Result<EvalResult, EvalError> {
        self.eval_expr(expr)
    }

    fn eval_expr(&mut self, expr: &Expr) -> Result<EvalResult, EvalError> {
        match expr {
            Expr::Number(n) => Ok(EvalResult::Value(*n)),
            Expr::Variable(name) => {
                self.variables
                    .get(name)
                    .copied()
                    .map(EvalResult::Value)
                    .ok_or_else(|| EvalError::UndefinedVariable(name.clone()))
            }
            Expr::Assign(name, value_expr) => {
                let value = self.eval_value(value_expr)?;
                self.variables.insert(name.clone(), value);
                Ok(EvalResult::Assignment)
            }
            Expr::UnaryMinus(operand) => {
                let value = self.eval_value(operand)?;
                Ok(EvalResult::Value(-value))
            }
            Expr::Not(operand) => {
                let value = self.eval_value(operand)?;
                let result = if value == 0.0 { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Add(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                Ok(EvalResult::Value(va + vb))
            }
            Expr::Sub(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                Ok(EvalResult::Value(va - vb))
            }
            Expr::Mul(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                Ok(EvalResult::Value(va * vb))
            }
            Expr::Div(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                if vb == 0.0 {
                    return Err(EvalError::DivisionByZero);
                }
                Ok(EvalResult::Value(va / vb))
            }
            Expr::Mod(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                if vb == 0.0 {
                    return Err(EvalError::DivisionByZero);
                }
                Ok(EvalResult::Value(va % vb))
            }
            Expr::Eq(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                let result = if (va - vb).abs() < f64::EPSILON { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Ne(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                let result = if (va - vb).abs() >= f64::EPSILON { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Lt(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                let result = if va < vb { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Le(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                let result = if va <= vb { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Gt(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                let result = if va > vb { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Ge(a, b) => {
                let va = self.eval_value(a)?;
                let vb = self.eval_value(b)?;
                let result = if va >= vb { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::And(a, b) => {
                let va = self.eval_value(a)?;
                if va == 0.0 {
                    return Ok(EvalResult::Value(0.0));
                }
                let vb = self.eval_value(b)?;
                let result = if vb != 0.0 { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::Or(a, b) => {
                let va = self.eval_value(a)?;
                if va != 0.0 {
                    return Ok(EvalResult::Value(1.0));
                }
                let vb = self.eval_value(b)?;
                let result = if vb != 0.0 { 1.0 } else { 0.0 };
                Ok(EvalResult::Value(result))
            }
            Expr::IfElse(cond, then_val, else_val) => {
                let cond_val = self.eval_value(cond)?;
                let then_result = self.eval_value(then_val)?;
                let else_result = self.eval_value(else_val)?;
                let result = if cond_val != 0.0 { then_result } else { else_result };
                Ok(EvalResult::Value(result))
            }
            Expr::FunctionCall(name, args) => {
                let mut arg_values = Vec::new();
                for arg in args {
                    arg_values.push(self.eval_value(arg)?);
                }
                match call_function(name, arg_values) {
                    Ok(v) => Ok(EvalResult::Value(v)),
                    Err(e) => Err(EvalError::FunctionError(e.to_string())),
                }
            }
        }
    }

    fn eval_value(&mut self, expr: &Expr) -> Result<f64, EvalError> {
        match self.eval_expr(expr)? {
            EvalResult::Value(v) => Ok(v),
            EvalResult::Assignment => Err(EvalError::Other("unexpected assignment in expression".to_string())),
        }
    }
}
