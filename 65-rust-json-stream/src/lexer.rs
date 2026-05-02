use crate::error::{Error, ErrorKind, Result};
use crate::event::Position;
use std::io::{BufReader, Read};

#[derive(Debug, Clone, PartialEq)]
pub enum Token {
    ObjectStart,
    ObjectEnd,
    ArrayStart,
    ArrayEnd,
    Colon,
    Comma,
    String(String),
    Integer(i64),
    Float(f64),
    Bool(bool),
    Null,
}

pub struct Lexer<R: Read> {
    reader: BufReader<R>,
    current_byte: Option<u8>,
    line: u32,
    column: u32,
    token_start_position: Position,
    buf: Vec<u8>,
}

impl<R: Read> Lexer<R> {
    pub fn new(reader: R) -> Self {
        let mut lexer = Lexer {
            reader: BufReader::new(reader),
            current_byte: None,
            line: 1,
            column: 0,
            token_start_position: Position::new(1, 1),
            buf: Vec::new(),
        };
        let _ = lexer.advance();
        lexer
    }

    fn advance(&mut self) -> Option<u8> {
        let mut buf = [0; 1];
        match self.reader.read(&mut buf) {
            Ok(0) => {
                self.current_byte = None;
                None
            }
            Ok(n) if n == 1 => {
                let byte = buf[0];
                if let Some(current) = self.current_byte {
                    if current == b'\n' {
                        self.line += 1;
                        self.column = 1;
                    } else {
                        self.column += 1;
                    }
                } else {
                    self.column = 1;
                }
                self.current_byte = Some(byte);
                Some(byte)
            }
            Ok(_) => {
                self.current_byte = None;
                None
            }
            Err(_) => {
                self.current_byte = None;
                None
            }
        }
    }

    fn current(&mut self) -> Option<u8> {
        self.current_byte
    }

    fn position(&self) -> Position {
        Position::new(self.line, self.column)
    }

    fn skip_whitespace(&mut self) {
        while let Some(b) = self.current() {
            if b.is_ascii_whitespace() {
                self.advance();
            } else {
                break;
            }
        }
    }

    pub fn next_token(&mut self) -> Result<Option<(Token, Position)>> {
        self.skip_whitespace();
        
        self.token_start_position = self.position();
        
        match self.current() {
            Some(b'{') => {
                self.advance();
                Ok(Some((Token::ObjectStart, self.token_start_position)))
            }
            Some(b'}') => {
                self.advance();
                Ok(Some((Token::ObjectEnd, self.token_start_position)))
            }
            Some(b'[') => {
                self.advance();
                Ok(Some((Token::ArrayStart, self.token_start_position)))
            }
            Some(b']') => {
                self.advance();
                Ok(Some((Token::ArrayEnd, self.token_start_position)))
            }
            Some(b':') => {
                self.advance();
                Ok(Some((Token::Colon, self.token_start_position)))
            }
            Some(b',') => {
                self.advance();
                Ok(Some((Token::Comma, self.token_start_position)))
            }
            Some(b'"') => self.read_string(),
            Some(b'-') => self.read_number(),
            Some(b) if b.is_ascii_digit() => self.read_number(),
            Some(b't') => self.read_identifier(b"true", Token::Bool(true)),
            Some(b'f') => self.read_identifier(b"false", Token::Bool(false)),
            Some(b'n') => self.read_identifier(b"null", Token::Null),
            Some(b) => Err(Error::new(ErrorKind::UnexpectedChar(b as char))
                .with_position(self.token_start_position)),
            None => Ok(None),
        }
    }

    fn read_identifier(&mut self, expected: &[u8], token: Token) -> Result<Option<(Token, Position)>> {
        let start_pos = self.token_start_position;
        
        for &expected_byte in expected {
            match self.current() {
                Some(b) if b == expected_byte => {
                    self.advance();
                }
                Some(b) => {
                    return Err(Error::new(ErrorKind::UnexpectedChar(b as char))
                        .with_position(self.position()));
                }
                None => {
                    return Err(Error::new(ErrorKind::UnexpectedEnd)
                        .with_position(start_pos));
                }
            }
        }
        
        Ok(Some((token, start_pos)))
    }

    fn read_string(&mut self) -> Result<Option<(Token, Position)>> {
        let start_pos = self.token_start_position;
        self.buf.clear();
        
        self.advance();
        
        loop {
            match self.current() {
                Some(b'"') => {
                    self.advance();
                    let s = String::from_utf8(self.buf.clone())
                        .map_err(|_| Error::new(ErrorKind::InvalidUtf8Sequence)
                            .with_position(start_pos))?;
                    return Ok(Some((Token::String(s), start_pos)));
                }
                Some(b'\\') => {
                    self.advance();
                    self.read_escape_sequence()?;
                }
                Some(b) if b < 0x20 => {
                    return Err(Error::new(ErrorKind::UnexpectedChar(b as char))
                        .with_position(self.position()));
                }
                Some(b) => {
                    self.buf.push(b);
                    self.advance();
                }
                None => {
                    return Err(Error::new(ErrorKind::UnexpectedEnd)
                        .with_position(start_pos));
                }
            }
        }
    }

    fn read_escape_sequence(&mut self) -> Result<()> {
        let pos = self.position();
        match self.current() {
            Some(b'"') => {
                self.buf.push(b'"');
                self.advance();
            }
            Some(b'\\') => {
                self.buf.push(b'\\');
                self.advance();
            }
            Some(b'/') => {
                self.buf.push(b'/');
                self.advance();
            }
            Some(b'b') => {
                self.buf.push(b'\x08');
                self.advance();
            }
            Some(b'f') => {
                self.buf.push(b'\x0c');
                self.advance();
            }
            Some(b'n') => {
                self.buf.push(b'\n');
                self.advance();
            }
            Some(b'r') => {
                self.buf.push(b'\r');
                self.advance();
            }
            Some(b't') => {
                self.buf.push(b'\t');
                self.advance();
            }
            Some(b'u') => {
                self.advance();
                let codepoint = self.read_unicode_escape()?;
                self.encode_utf8_codepoint(codepoint)?;
            }
            Some(b) => {
                return Err(Error::new(ErrorKind::UnexpectedChar(b as char))
                    .with_position(pos));
            }
            None => {
                return Err(Error::new(ErrorKind::UnexpectedEnd)
                    .with_position(pos));
            }
        }
        Ok(())
    }

    fn read_unicode_escape(&mut self) -> Result<u32> {
        let pos = self.position();
        let mut codepoint = 0u32;
        
        for _ in 0..4 {
            match self.current() {
                Some(b) => {
                    let digit = match b {
                        b'0'..=b'9' => (b - b'0') as u32,
                        b'a'..=b'f' => (b - b'a' + 10) as u32,
                        b'A'..=b'F' => (b - b'A' + 10) as u32,
                        _ => {
                            return Err(Error::new(ErrorKind::InvalidUnicodeEscape)
                                .with_position(self.position()));
                        }
                    };
                    codepoint = (codepoint << 4) | digit;
                    self.advance();
                }
                None => {
                    return Err(Error::new(ErrorKind::UnexpectedEnd)
                        .with_position(pos));
                }
            }
        }
        
        Ok(codepoint)
    }

    fn encode_utf8_codepoint(&mut self, codepoint: u32) -> Result<()> {
        if codepoint <= 0x7F {
            self.buf.push(codepoint as u8);
        } else if codepoint <= 0x7FF {
            self.buf.push(0xC0 | ((codepoint >> 6) as u8));
            self.buf.push(0x80 | ((codepoint & 0x3F) as u8));
        } else if codepoint <= 0xFFFF {
            if (0xD800..=0xDFFF).contains(&codepoint) {
                return Err(Error::new(ErrorKind::InvalidEscape(codepoint))
                    .with_position(self.position()));
            }
            self.buf.push(0xE0 | ((codepoint >> 12) as u8));
            self.buf.push(0x80 | (((codepoint >> 6) & 0x3F) as u8));
            self.buf.push(0x80 | ((codepoint & 0x3F) as u8));
        } else if codepoint <= 0x10FFFF {
            self.buf.push(0xF0 | ((codepoint >> 18) as u8));
            self.buf.push(0x80 | (((codepoint >> 12) & 0x3F) as u8));
            self.buf.push(0x80 | (((codepoint >> 6) & 0x3F) as u8));
            self.buf.push(0x80 | ((codepoint & 0x3F) as u8));
        } else {
            return Err(Error::new(ErrorKind::InvalidEscape(codepoint))
                .with_position(self.position()));
        }
        Ok(())
    }

    fn read_number(&mut self) -> Result<Option<(Token, Position)>> {
        let start_pos = self.token_start_position;
        self.buf.clear();
        
        let mut is_float = false;
        
        if self.current() == Some(b'-') {
            self.buf.push(b'-');
            self.advance();
        }
        
        match self.current() {
            Some(b'0') => {
                self.buf.push(b'0');
                self.advance();
            }
            Some(b) if (b'1'..=b'9').contains(&b) => {
                while let Some(b) = self.current() {
                    if b.is_ascii_digit() {
                        self.buf.push(b);
                        self.advance();
                    } else {
                        break;
                    }
                }
            }
            _ => {
                return Err(Error::new(ErrorKind::ExpectedValue)
                    .with_position(start_pos));
            }
        }
        
        if self.current() == Some(b'.') {
            is_float = true;
            self.buf.push(b'.');
            self.advance();
            
            if !self.current().map(|b| b.is_ascii_digit()).unwrap_or(false) {
                return Err(Error::new(ErrorKind::ExpectedValue)
                    .with_position(self.position()));
            }
            
            while let Some(b) = self.current() {
                if b.is_ascii_digit() {
                    self.buf.push(b);
                    self.advance();
                } else {
                    break;
                }
            }
        }
        
        if self.current() == Some(b'e') || self.current() == Some(b'E') {
            is_float = true;
            self.buf.push(b'e');
            self.advance();
            
            if self.current() == Some(b'+') || self.current() == Some(b'-') {
                let sign = self.current().unwrap();
                self.buf.push(sign);
                self.advance();
            }
            
            if !self.current().map(|b| b.is_ascii_digit()).unwrap_or(false) {
                return Err(Error::new(ErrorKind::ExpectedValue)
                    .with_position(self.position()));
            }
            
            while let Some(b) = self.current() {
                if b.is_ascii_digit() {
                    self.buf.push(b);
                    self.advance();
                } else {
                    break;
                }
            }
        }
        
        let s = String::from_utf8_lossy(&self.buf);
        
        let token = if is_float {
            let f: f64 = s.parse()
                .map_err(|e: std::num::ParseFloatError| {
                    Error::new(ErrorKind::ParseFloat(e)).with_position(start_pos)
                })?;
            Token::Float(f)
        } else {
            let i: i64 = s.parse()
                .map_err(|e: std::num::ParseIntError| {
                    Error::new(ErrorKind::ParseInt(e)).with_position(start_pos)
                })?;
            Token::Integer(i)
        };
        
        Ok(Some((token, start_pos)))
    }
}
