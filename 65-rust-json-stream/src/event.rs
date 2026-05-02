#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Position {
    pub line: u32,
    pub column: u32,
}

impl Position {
    pub fn new(line: u32, column: u32) -> Self {
        Position { line, column }
    }
}

impl std::fmt::Display for Position {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}:{}", self.line, self.column)
    }
}

#[derive(Debug, Clone, PartialEq)]
pub enum Number {
    Integer(i64),
    Float(f64),
}

impl Number {
    pub fn to_f64(&self) -> f64 {
        match self {
            Number::Integer(i) => *i as f64,
            Number::Float(f) => *f,
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub enum Event<'a> {
    ObjectStart(Position),
    ObjectEnd(Position),
    ArrayStart(Position),
    ArrayEnd(Position),
    Key(&'a str, Position),
    String(&'a str, Position),
    Number(Number, Position),
    Bool(bool, Position),
    Null(Position),
}
