#[derive(Debug, Clone, PartialEq)]
pub enum Token {
    Number(f64),
    Identifier(String),
    If,
    Assign,
    Plus,
    Minus,
    Star,
    Slash,
    Percent,
    Eq,
    Ne,
    Lt,
    Le,
    Gt,
    Ge,
    And,
    Or,
    Not,
    LParen,
    RParen,
    Comma,
}

#[derive(Debug, Clone)]
pub enum TokenizeError {
    UnexpectedChar(char),
    InvalidNumber(String),
}

impl std::fmt::Display for TokenizeError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TokenizeError::UnexpectedChar(c) => write!(f, "unexpected character: '{}'", c),
            TokenizeError::InvalidNumber(s) => write!(f, "invalid number: '{}'", s),
        }
    }
}

pub struct Tokenizer {
    input: Vec<char>,
    pos: usize,
}

impl Tokenizer {
    pub fn new(input: &str) -> Self {
        Tokenizer {
            input: input.chars().collect(),
            pos: 0,
        }
    }

    fn peek(&self) -> Option<char> {
        self.input.get(self.pos).copied()
    }

    fn advance(&mut self) -> Option<char> {
        let c = self.input.get(self.pos).copied();
        self.pos += 1;
        c
    }

    fn skip_whitespace(&mut self) {
        while let Some(c) = self.peek() {
            if c.is_whitespace() {
                self.advance();
            } else {
                break;
            }
        }
    }

    fn read_number(&mut self, first: char) -> Result<Token, TokenizeError> {
        let mut num_str = String::new();
        num_str.push(first);

        let mut has_dot = false;

        while let Some(c) = self.peek() {
            if c.is_ascii_digit() {
                num_str.push(c);
                self.advance();
            } else if c == '.' && !has_dot {
                num_str.push(c);
                has_dot = true;
                self.advance();
            } else {
                break;
            }
        }

        num_str.parse::<f64>()
            .map(Token::Number)
            .map_err(|_| TokenizeError::InvalidNumber(num_str))
    }

    fn read_identifier(&mut self, first: char) -> Token {
        let mut ident = String::new();
        ident.push(first);

        while let Some(c) = self.peek() {
            if c.is_ascii_alphanumeric() || c == '_' {
                ident.push(c);
                self.advance();
            } else {
                break;
            }
        }

        match ident.as_str() {
            "if" => Token::If,
            "sqrt" | "abs" | "max" | "min" => Token::Identifier(ident),
            _ => Token::Identifier(ident),
        }
    }

    pub fn tokenize(&mut self) -> Result<Vec<Token>, TokenizeError> {
        let mut tokens = Vec::new();

        loop {
            self.skip_whitespace();

            let Some(c) = self.advance() else {
                break;
            };

            let token = match c {
                '=' => {
                    if self.peek() == Some('=') {
                        self.advance();
                        Token::Eq
                    } else {
                        Token::Assign
                    }
                }
                '!' => {
                    if self.peek() == Some('=') {
                        self.advance();
                        Token::Ne
                    } else {
                        Token::Not
                    }
                }
                '<' => {
                    if self.peek() == Some('=') {
                        self.advance();
                        Token::Le
                    } else {
                        Token::Lt
                    }
                }
                '>' => {
                    if self.peek() == Some('=') {
                        self.advance();
                        Token::Ge
                    } else {
                        Token::Gt
                    }
                }
                '&' => {
                    if self.peek() == Some('&') {
                        self.advance();
                        Token::And
                    } else {
                        return Err(TokenizeError::UnexpectedChar(c));
                    }
                }
                '|' => {
                    if self.peek() == Some('|') {
                        self.advance();
                        Token::Or
                    } else {
                        return Err(TokenizeError::UnexpectedChar(c));
                    }
                }
                '+' => Token::Plus,
                '-' => Token::Minus,
                '*' => Token::Star,
                '/' => Token::Slash,
                '%' => Token::Percent,
                '(' => Token::LParen,
                ')' => Token::RParen,
                ',' => Token::Comma,
                _ => {
                    if c.is_ascii_digit() || c == '.' {
                        self.read_number(c)?
                    } else if c.is_ascii_alphabetic() || c == '_' {
                        self.read_identifier(c)
                    } else {
                        return Err(TokenizeError::UnexpectedChar(c));
                    }
                }
            };

            tokens.push(token);
        }

        Ok(tokens)
    }
}

pub fn tokenize(input: &str) -> Result<Vec<Token>, TokenizeError> {
    let mut tokenizer = Tokenizer::new(input);
    tokenizer.tokenize()
}
