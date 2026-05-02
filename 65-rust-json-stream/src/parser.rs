use crate::error::{Error, ErrorKind, Result};
use crate::event::{Event, Number, Position};
use crate::lexer::{Lexer, Token};
use std::io::Read;

#[derive(Debug, Clone, Copy)]
pub struct Config {
    pub max_depth: usize,
    pub allow_trailing_comma: bool,
}

impl Default for Config {
    fn default() -> Self {
        Config {
            max_depth: 128,
            allow_trailing_comma: true,
        }
    }
}

pub struct Parser<R: Read, F: FnMut(Event)> {
    lexer: Lexer<R>,
    current_token: Option<(Token, Position)>,
    callback: F,
    config: Config,
    depth: usize,
}

impl<R: Read, F: FnMut(Event)> Parser<R, F> {
    pub fn new(reader: R, callback: F) -> Self {
        Parser::with_config(reader, callback, Config::default())
    }

    pub fn with_config(reader: R, callback: F, config: Config) -> Self {
        let mut parser = Parser {
            lexer: Lexer::new(reader),
            current_token: None,
            callback,
            config,
            depth: 0,
        };
        let _ = parser.advance_token();
        parser
    }

    fn advance_token(&mut self) -> Result<()> {
        self.current_token = self.lexer.next_token()?;
        Ok(())
    }

    fn current(&self) -> Option<(&Token, &Position)> {
        self.current_token.as_ref().map(|(t, p)| (t, p))
    }

    fn position(&self) -> Position {
        self.current()
            .map(|(_, p)| *p)
            .unwrap_or_else(|| Position::new(0, 0))
    }

    fn check_depth(&mut self) -> Result<()> {
        if self.depth >= self.config.max_depth {
            return Err(Error::new(ErrorKind::MaxDepthExceeded)
                .with_position(self.position()));
        }
        self.depth += 1;
        Ok(())
    }

    fn pop_depth(&mut self) {
        if self.depth > 0 {
            self.depth -= 1;
        }
    }

    pub fn parse(&mut self) -> Result<()> {
        self.parse_value()?;
        if self.current().is_some() {
            return Err(Error::new(ErrorKind::UnexpectedChar(
                match self.current() {
                    Some((Token::ObjectStart, _)) => '{',
                    Some((Token::ObjectEnd, _)) => '}',
                    Some((Token::ArrayStart, _)) => '[',
                    Some((Token::ArrayEnd, _)) => ']',
                    Some((Token::Colon, _)) => ':',
                    Some((Token::Comma, _)) => ',',
                    _ => ' ',
                }
            )).with_position(self.position()));
        }
        Ok(())
    }

    fn parse_value(&mut self) -> Result<()> {
        match self.current() {
            Some((Token::ObjectStart, pos)) => {
                let pos = *pos;
                self.check_depth()?;
                self.advance_token()?;
                (self.callback)(Event::ObjectStart(pos));
                self.parse_object()?;
                self.pop_depth();
            }
            Some((Token::ArrayStart, pos)) => {
                let pos = *pos;
                self.check_depth()?;
                self.advance_token()?;
                (self.callback)(Event::ArrayStart(pos));
                self.parse_array()?;
                self.pop_depth();
            }
            Some((Token::String(s), pos)) => {
                let pos = *pos;
                let s = s.clone();
                self.advance_token()?;
                (self.callback)(Event::String(&s, pos));
            }
            Some((Token::Integer(i), pos)) => {
                let pos = *pos;
                let num = Number::Integer(*i);
                self.advance_token()?;
                (self.callback)(Event::Number(num, pos));
            }
            Some((Token::Float(f), pos)) => {
                let pos = *pos;
                let num = Number::Float(*f);
                self.advance_token()?;
                (self.callback)(Event::Number(num, pos));
            }
            Some((Token::Bool(b), pos)) => {
                let pos = *pos;
                let val = *b;
                self.advance_token()?;
                (self.callback)(Event::Bool(val, pos));
            }
            Some((Token::Null, pos)) => {
                let pos = *pos;
                self.advance_token()?;
                (self.callback)(Event::Null(pos));
            }
            Some((_, pos)) => {
                return Err(Error::new(ErrorKind::ExpectedValue)
                    .with_position(*pos));
            }
            None => {
                return Err(Error::new(ErrorKind::UnexpectedEnd));
            }
        }
        Ok(())
    }

    fn parse_object(&mut self) -> Result<()> {
        let mut expecting_key = true;
        
        loop {
            match self.current() {
                Some((Token::ObjectEnd, pos)) => {
                    let pos = *pos;
                    self.advance_token()?;
                    (self.callback)(Event::ObjectEnd(pos));
                    return Ok(());
                }
                Some((Token::Comma, _)) => {
                    let comma_pos = self.position();
                    
                    if expecting_key {
                        return Err(Error::new(ErrorKind::ExpectedValue)
                            .with_position(comma_pos));
                    }
                    
                    self.advance_token()?;
                    
                    match self.current() {
                        Some((Token::ObjectEnd, _)) => {
                            if self.config.allow_trailing_comma {
                                continue;
                            } else {
                                return Err(Error::new(ErrorKind::TrailingComma)
                                    .with_position(comma_pos));
                            }
                        }
                        Some((Token::Comma, _)) => {
                            return Err(Error::new(ErrorKind::ExpectedValue)
                                .with_position(self.position()));
                        }
                        Some((Token::String(_), _)) => {
                            expecting_key = true;
                            continue;
                        }
                        Some((_, pos)) => {
                            return Err(Error::new(ErrorKind::ExpectedValue)
                                .with_position(*pos));
                        }
                        None => {
                            return Err(Error::new(ErrorKind::UnexpectedEnd));
                        }
                    }
                }
                Some((Token::String(key), pos)) => {
                    if !expecting_key {
                        return Err(Error::new(ErrorKind::ExpectedCommaOrBracket)
                            .with_position(*pos));
                    }
                    
                    let key_pos = *pos;
                    let key = key.clone();
                    self.advance_token()?;
                    (self.callback)(Event::Key(&key, key_pos));
                    
                    match self.current() {
                        Some((Token::Colon, _)) => {
                            self.advance_token()?;
                        }
                        Some((_, pos)) => {
                            return Err(Error::new(ErrorKind::ExpectedColon)
                                .with_position(*pos));
                        }
                        None => {
                            return Err(Error::new(ErrorKind::UnexpectedEnd));
                        }
                    }
                    
                    self.parse_value()?;
                    expecting_key = false;
                }
                Some((_, pos)) => {
                    return Err(Error::new(ErrorKind::ExpectedValue)
                        .with_position(*pos));
                }
                None => {
                    return Err(Error::new(ErrorKind::UnexpectedEnd));
                }
            }
        }
    }

    fn parse_array(&mut self) -> Result<()> {
        let mut expecting_value = true;
        
        loop {
            match self.current() {
                Some((Token::ArrayEnd, pos)) => {
                    let pos = *pos;
                    self.advance_token()?;
                    (self.callback)(Event::ArrayEnd(pos));
                    return Ok(());
                }
                Some((Token::Comma, _)) => {
                    let comma_pos = self.position();
                    
                    if expecting_value {
                        return Err(Error::new(ErrorKind::ExpectedValue)
                            .with_position(comma_pos));
                    }
                    
                    self.advance_token()?;
                    
                    match self.current() {
                        Some((Token::ArrayEnd, _)) => {
                            if self.config.allow_trailing_comma {
                                continue;
                            } else {
                                return Err(Error::new(ErrorKind::TrailingComma)
                                    .with_position(comma_pos));
                            }
                        }
                        Some((Token::Comma, _)) => {
                            return Err(Error::new(ErrorKind::ExpectedValue)
                                .with_position(self.position()));
                        }
                        Some(_) => {
                            expecting_value = true;
                            continue;
                        }
                        None => {
                            return Err(Error::new(ErrorKind::UnexpectedEnd));
                        }
                    }
                }
                Some(_) => {
                    if !expecting_value {
                        return Err(Error::new(ErrorKind::ExpectedCommaOrBracket)
                            .with_position(self.position()));
                    }
                    
                    self.parse_value()?;
                    expecting_value = false;
                }
                None => {
                    return Err(Error::new(ErrorKind::UnexpectedEnd));
                }
            }
        }
    }
}

pub fn parse<R: Read, F: FnMut(Event)>(reader: R, callback: F) -> Result<()> {
    let mut parser = Parser::new(reader, callback);
    parser.parse()
}

pub fn parse_with_config<R: Read, F: FnMut(Event)>(
    reader: R,
    callback: F,
    config: Config,
) -> Result<()> {
    let mut parser = Parser::with_config(reader, callback, config);
    parser.parse()
}
