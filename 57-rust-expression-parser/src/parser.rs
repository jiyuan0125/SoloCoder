use crate::tokenizer::Token;

#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    Number(f64),
    Variable(String),
    Assign(String, Box<Expr>),
    UnaryMinus(Box<Expr>),
    Not(Box<Expr>),
    Add(Box<Expr>, Box<Expr>),
    Sub(Box<Expr>, Box<Expr>),
    Mul(Box<Expr>, Box<Expr>),
    Div(Box<Expr>, Box<Expr>),
    Mod(Box<Expr>, Box<Expr>),
    Eq(Box<Expr>, Box<Expr>),
    Ne(Box<Expr>, Box<Expr>),
    Lt(Box<Expr>, Box<Expr>),
    Le(Box<Expr>, Box<Expr>),
    Gt(Box<Expr>, Box<Expr>),
    Ge(Box<Expr>, Box<Expr>),
    And(Box<Expr>, Box<Expr>),
    Or(Box<Expr>, Box<Expr>),
    IfElse(Box<Expr>, Box<Expr>, Box<Expr>),
    FunctionCall(String, Vec<Expr>),
}

#[derive(Debug, Clone)]
pub enum ParseError {
    UnexpectedToken(Token),
    ExpectedToken(Token),
    UnexpectedEnd,
    InvalidIdentifier,
}

impl std::fmt::Display for ParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ParseError::UnexpectedToken(t) => write!(f, "unexpected token: {:?}", t),
            ParseError::ExpectedToken(t) => write!(f, "expected token: {:?}", t),
            ParseError::UnexpectedEnd => write!(f, "unexpected end of input"),
            ParseError::InvalidIdentifier => write!(f, "invalid identifier"),
        }
    }
}

pub struct Parser {
    tokens: Vec<Token>,
    pos: usize,
}

impl Parser {
    pub fn new(tokens: Vec<Token>) -> Self {
        Parser { tokens, pos: 0 }
    }

    fn peek(&self) -> Option<&Token> {
        self.tokens.get(self.pos)
    }

    fn advance(&mut self) -> Option<Token> {
        let token = self.tokens.get(self.pos).cloned();
        self.pos += 1;
        token
    }

    fn expect(&mut self, expected: Token) -> Result<(), ParseError> {
        match self.advance() {
            Some(token) if token == expected => Ok(()),
            Some(token) => Err(ParseError::UnexpectedToken(token)),
            None => Err(ParseError::UnexpectedEnd),
        }
    }

    pub fn parse(&mut self) -> Result<Expr, ParseError> {
        let result = self.parse_assignment()?;
        if self.peek().is_some() {
            return Err(ParseError::UnexpectedToken(self.peek().unwrap().clone()));
        }
        Ok(result)
    }

    fn parse_assignment(&mut self) -> Result<Expr, ParseError> {
        if let Some(Token::Identifier(name)) = self.peek() {
            let name = name.clone();
            let save_pos = self.pos;
            self.advance();

            if let Some(Token::Assign) = self.peek() {
                self.advance();
                let value = self.parse_or()?;
                return Ok(Expr::Assign(name, Box::new(value)));
            } else {
                self.pos = save_pos;
            }
        }
        self.parse_or()
    }

    fn parse_or(&mut self) -> Result<Expr, ParseError> {
        let mut left = self.parse_and()?;

        while let Some(Token::Or) = self.peek() {
            self.advance();
            let right = self.parse_and()?;
            left = Expr::Or(Box::new(left), Box::new(right));
        }

        Ok(left)
    }

    fn parse_and(&mut self) -> Result<Expr, ParseError> {
        let mut left = self.parse_comparison()?;

        while let Some(Token::And) = self.peek() {
            self.advance();
            let right = self.parse_comparison()?;
            left = Expr::And(Box::new(left), Box::new(right));
        }

        Ok(left)
    }

    fn parse_comparison(&mut self) -> Result<Expr, ParseError> {
        let mut left = self.parse_add_sub()?;

        loop {
            match self.peek() {
                Some(Token::Eq) => {
                    self.advance();
                    let right = self.parse_add_sub()?;
                    left = Expr::Eq(Box::new(left), Box::new(right));
                }
                Some(Token::Ne) => {
                    self.advance();
                    let right = self.parse_add_sub()?;
                    left = Expr::Ne(Box::new(left), Box::new(right));
                }
                Some(Token::Lt) => {
                    self.advance();
                    let right = self.parse_add_sub()?;
                    left = Expr::Lt(Box::new(left), Box::new(right));
                }
                Some(Token::Le) => {
                    self.advance();
                    let right = self.parse_add_sub()?;
                    left = Expr::Le(Box::new(left), Box::new(right));
                }
                Some(Token::Gt) => {
                    self.advance();
                    let right = self.parse_add_sub()?;
                    left = Expr::Gt(Box::new(left), Box::new(right));
                }
                Some(Token::Ge) => {
                    self.advance();
                    let right = self.parse_add_sub()?;
                    left = Expr::Ge(Box::new(left), Box::new(right));
                }
                _ => break,
            }
        }

        Ok(left)
    }

    fn parse_add_sub(&mut self) -> Result<Expr, ParseError> {
        let mut left = self.parse_mul_div_mod()?;

        loop {
            match self.peek() {
                Some(Token::Plus) => {
                    self.advance();
                    let right = self.parse_mul_div_mod()?;
                    left = Expr::Add(Box::new(left), Box::new(right));
                }
                Some(Token::Minus) => {
                    self.advance();
                    let right = self.parse_mul_div_mod()?;
                    left = Expr::Sub(Box::new(left), Box::new(right));
                }
                _ => break,
            }
        }

        Ok(left)
    }

    fn parse_mul_div_mod(&mut self) -> Result<Expr, ParseError> {
        let mut left = self.parse_unary()?;

        loop {
            match self.peek() {
                Some(Token::Star) => {
                    self.advance();
                    let right = self.parse_unary()?;
                    left = Expr::Mul(Box::new(left), Box::new(right));
                }
                Some(Token::Slash) => {
                    self.advance();
                    let right = self.parse_unary()?;
                    left = Expr::Div(Box::new(left), Box::new(right));
                }
                Some(Token::Percent) => {
                    self.advance();
                    let right = self.parse_unary()?;
                    left = Expr::Mod(Box::new(left), Box::new(right));
                }
                _ => break,
            }
        }

        Ok(left)
    }

    fn parse_unary(&mut self) -> Result<Expr, ParseError> {
        match self.peek() {
            Some(Token::Minus) => {
                self.advance();
                let operand = self.parse_unary()?;
                Ok(Expr::UnaryMinus(Box::new(operand)))
            }
            Some(Token::Not) => {
                self.advance();
                let operand = self.parse_unary()?;
                Ok(Expr::Not(Box::new(operand)))
            }
            _ => self.parse_primary(),
        }
    }

    fn parse_primary(&mut self) -> Result<Expr, ParseError> {
        match self.peek() {
            Some(Token::Number(n)) => {
                let n = *n;
                self.advance();
                Ok(Expr::Number(n))
            }
            Some(Token::If) => {
                self.advance();
                self.expect(Token::LParen)?;
                let cond = self.parse_assignment()?;
                self.expect(Token::Comma)?;
                let then_val = self.parse_assignment()?;
                self.expect(Token::Comma)?;
                let else_val = self.parse_assignment()?;
                self.expect(Token::RParen)?;
                Ok(Expr::IfElse(
                    Box::new(cond),
                    Box::new(then_val),
                    Box::new(else_val),
                ))
            }
            Some(Token::Identifier(name)) => {
                let name = name.clone();
                let save_pos = self.pos;
                self.advance();

                if let Some(Token::LParen) = self.peek() {
                    self.advance();
                    let mut args = Vec::new();

                    if self.peek() != Some(&Token::RParen) {
                        args.push(self.parse_assignment()?);
                        while let Some(Token::Comma) = self.peek() {
                            self.advance();
                            args.push(self.parse_assignment()?);
                        }
                    }

                    self.expect(Token::RParen)?;
                    Ok(Expr::FunctionCall(name, args))
                } else {
                    self.pos = save_pos + 1;
                    Ok(Expr::Variable(name))
                }
            }
            Some(Token::LParen) => {
                self.advance();
                let expr = self.parse_assignment()?;
                self.expect(Token::RParen)?;
                Ok(expr)
            }
            Some(token) => Err(ParseError::UnexpectedToken(token.clone())),
            None => Err(ParseError::UnexpectedEnd),
        }
    }
}

pub fn parse(tokens: Vec<Token>) -> Result<Expr, ParseError> {
    let mut parser = Parser::new(tokens);
    parser.parse()
}
