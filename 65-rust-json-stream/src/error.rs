use crate::event::Position;
use std::fmt;

#[derive(Debug)]
pub enum ErrorKind {
    Io(std::io::Error),
    Utf8(std::str::Utf8Error),
    ParseInt(std::num::ParseIntError),
    ParseFloat(std::num::ParseFloatError),
    UnexpectedChar(char),
    UnexpectedEnd,
    InvalidEscape(u32),
    InvalidUnicodeEscape,
    InvalidUtf8Sequence,
    MaxDepthExceeded,
    ExpectedColon,
    ExpectedValue,
    ExpectedCommaOrBracket,
    TrailingComma,
}

#[derive(Debug)]
pub struct Error {
    pub kind: ErrorKind,
    pub position: Option<Position>,
}

impl Error {
    pub fn new(kind: ErrorKind) -> Self {
        Error { kind, position: None }
    }

    pub fn with_position(mut self, position: Position) -> Self {
        self.position = Some(position);
        self
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.kind {
            ErrorKind::Io(e) => write!(f, "IO error: {}", e),
            ErrorKind::Utf8(e) => write!(f, "UTF-8 error: {}", e),
            ErrorKind::ParseInt(e) => write!(f, "Parse int error: {}", e),
            ErrorKind::ParseFloat(e) => write!(f, "Parse float error: {}", e),
            ErrorKind::UnexpectedChar(c) => write!(f, "Unexpected character '{}'", c),
            ErrorKind::UnexpectedEnd => write!(f, "Unexpected end of input"),
            ErrorKind::InvalidEscape(n) => write!(f, "Invalid escape sequence: \\u{:04X}", n),
            ErrorKind::InvalidUnicodeEscape => write!(f, "Invalid unicode escape sequence"),
            ErrorKind::InvalidUtf8Sequence => write!(f, "Invalid UTF-8 sequence"),
            ErrorKind::MaxDepthExceeded => write!(f, "Maximum nesting depth exceeded (max 128)"),
            ErrorKind::ExpectedColon => write!(f, "Expected colon after key"),
            ErrorKind::ExpectedValue => write!(f, "Expected value"),
            ErrorKind::ExpectedCommaOrBracket => write!(f, "Expected comma or closing bracket"),
            ErrorKind::TrailingComma => write!(f, "Trailing comma"),
        }?;

        if let Some(pos) = &self.position {
            write!(f, " at {}", pos)?;
        }

        Ok(())
    }
}

impl std::error::Error for Error {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match &self.kind {
            ErrorKind::Io(e) => Some(e),
            ErrorKind::Utf8(e) => Some(e),
            ErrorKind::ParseInt(e) => Some(e),
            ErrorKind::ParseFloat(e) => Some(e),
            _ => None,
        }
    }
}

impl From<std::io::Error> for Error {
    fn from(e: std::io::Error) -> Self {
        Error {
            kind: ErrorKind::Io(e),
            position: None,
        }
    }
}

impl From<std::str::Utf8Error> for Error {
    fn from(e: std::str::Utf8Error) -> Self {
        Error {
            kind: ErrorKind::Utf8(e),
            position: None,
        }
    }
}

pub type Result<T> = std::result::Result<T, Error>;
