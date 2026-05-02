use expr_calculator::tokenizer::tokenize;
use expr_calculator::parser::parse;
use expr_calculator::evaluator::{Evaluator, EvalResult, EvalError};
use std::io::{self, BufRead};

fn format_value(value: f64) -> String {
    if value.fract() == 0.0 {
        format!("{}", value as i64)
    } else {
        format!("{}", value)
    }
}

fn main() {
    let stdin = io::stdin();
    let mut evaluator = Evaluator::new();

    println!("Expression Calculator (type 'exit' to quit)");
    println!("--------------------------------------------");

    for line in stdin.lock().lines() {
        match line {
            Ok(input) => {
                let trimmed = input.trim();

                if trimmed.is_empty() {
                    continue;
                }

                if trimmed == "exit" || trimmed == "quit" {
                    println!("Bye!");
                    break;
                }

                let result = (|| -> Result<Option<String>, String> {
                    let tokens = tokenize(trimmed)
                        .map_err(|e| format!("{}", e))?;

                    if tokens.is_empty() {
                        return Ok(None);
                    }

                    let expr = parse(tokens)
                        .map_err(|e| format!("{}", e))?;

                    match evaluator.eval(&expr) {
                        Ok(EvalResult::Value(v)) => Ok(Some(format!("= {}", format_value(v)))),
                        Ok(EvalResult::Assignment) => Ok(None),
                        Err(EvalError::DivisionByZero) => Err("division by zero".to_string()),
                        Err(e) => Err(format!("{}", e)),
                    }
                })();

                match result {
                    Ok(Some(output)) => println!("{}", output),
                    Ok(None) => {}
                    Err(msg) => {
                        if msg == "division by zero" {
                            println!("Error: division by zero");
                        } else {
                            println!("Error: {}", msg);
                        }
                    }
                }
            }
            Err(e) => {
                println!("Error: failed to read input: {}", e);
                break;
            }
        }
    }
}
