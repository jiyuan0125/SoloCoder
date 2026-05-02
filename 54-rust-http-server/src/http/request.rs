use std::collections::HashMap;
use std::io::{self, Read};

use super::method::Method;

pub const MAX_HEADER_SIZE: usize = 8 * 1024;
pub const MAX_BODY_SIZE: usize = 1 * 1024 * 1024;

#[derive(Debug)]
pub struct ParseError;

#[derive(Debug)]
pub enum ParseResult {
    Complete(Request),
    Partial,
    Error(ParseError),
    TooLarge,
}

#[derive(Debug, Clone)]
pub struct Request {
    pub method: Method,
    pub path: String,
    pub query_params: HashMap<String, String>,
    pub headers: HashMap<String, String>,
    pub body: Vec<u8>,
}

impl Request {
    pub fn new() -> Self {
        Request {
            method: Method::Get,
            path: String::new(),
            query_params: HashMap::new(),
            headers: HashMap::new(),
            body: Vec::new(),
        }
    }

    pub fn query(&self) -> &HashMap<String, String> {
        &self.query_params
    }

    pub fn header(&self, name: &str) -> Option<&str> {
        self.headers.get(&name.to_lowercase()).map(|s| s.as_str())
    }

    pub fn content_length(&self) -> Option<usize> {
        self.header("content-length")
            .and_then(|v| v.parse::<usize>().ok())
    }

    pub fn connection_keep_alive(&self) -> bool {
        match self.header("connection") {
            Some(v) => v.eq_ignore_ascii_case("keep-alive"),
            None => false,
        }
    }

    pub fn is_websocket_upgrade(&self) -> bool {
        self.header("upgrade")
            .map(|v| v.eq_ignore_ascii_case("websocket"))
            .unwrap_or(false)
            && self.header("sec-websocket-key").is_some()
    }
}

pub struct RequestParser {
    buffer: Vec<u8>,
    headers_complete: bool,
    body_remaining: usize,
}

impl RequestParser {
    pub fn new() -> Self {
        RequestParser {
            buffer: Vec::new(),
            headers_complete: false,
            body_remaining: 0,
        }
    }

    pub fn consume(&mut self, data: &[u8]) {
        self.buffer.extend_from_slice(data);
    }

    pub fn parse(&mut self) -> ParseResult {
        if !self.headers_complete {
            let headers_end = self.find_headers_end();
            if headers_end.is_none() {
                if self.buffer.len() > MAX_HEADER_SIZE {
                    return ParseResult::TooLarge;
                }
                return ParseResult::Partial;
            }
            let headers_end = headers_end.unwrap();
            
            if headers_end > MAX_HEADER_SIZE {
                return ParseResult::TooLarge;
            }

            let headers_data = &self.buffer[..headers_end];
            let parse_result = self.parse_headers(headers_data);
            
            match parse_result {
                Ok((method, path, query_params, headers)) => {
                    let content_length = headers.get("content-length")
                        .and_then(|v| v.parse::<usize>().ok())
                        .unwrap_or(0);

                    if content_length > MAX_BODY_SIZE {
                        return ParseResult::TooLarge;
                    }

                    self.body_remaining = content_length;
                    self.headers_complete = true;

                    let request = Request {
                        method,
                        path,
                        query_params,
                        headers,
                        body: Vec::new(),
                    };

                    if self.body_remaining == 0 {
                        self.buffer.drain(..headers_end + 4);
                        return ParseResult::Complete(request);
                    } else {
                        let body_start = headers_end + 4;
                        let available = self.buffer.len().saturating_sub(body_start);
                        let to_read = std::cmp::min(available, self.body_remaining);
                        
                        if to_read > 0 {
                            let mut req = request;
                            req.body.extend_from_slice(&self.buffer[body_start..body_start + to_read]);
                            self.buffer.drain(..body_start + to_read);
                            self.body_remaining -= to_read;
                            
                            if self.body_remaining == 0 {
                                return ParseResult::Complete(req);
                            }
                        }
                        return ParseResult::Partial;
                    }
                }
                Err(_) => ParseResult::Error(ParseError),
            }
        } else {
            if self.body_remaining > 0 && !self.buffer.is_empty() {
                let to_read = std::cmp::min(self.buffer.len(), self.body_remaining);
                let body_data: Vec<u8> = self.buffer.drain(..to_read).collect();
                self.body_remaining -= to_read;
                
                if self.body_remaining == 0 {
                    let mut req = Request::new();
                    req.body = body_data;
                    ParseResult::Complete(req)
                } else {
                    ParseResult::Partial
                }
            } else {
                ParseResult::Partial
            }
        }
    }

    fn find_headers_end(&self) -> Option<usize> {
        for i in 0..self.buffer.len().saturating_sub(3) {
            if self.buffer[i] == b'\r' 
                && self.buffer[i+1] == b'\n'
                && self.buffer[i+2] == b'\r'
                && self.buffer[i+3] == b'\n' {
                return Some(i);
            }
        }
        None
    }

    fn parse_headers(&self, data: &[u8]) -> Result<(Method, String, HashMap<String, String>, HashMap<String, String>), ParseError> {
        let text = String::from_utf8_lossy(data);
        let mut lines = text.lines();

        let request_line = lines.next().ok_or(ParseError)?;
        let mut parts = request_line.split_whitespace();

        let method_str = parts.next().ok_or(ParseError)?;
        let method: Method = method_str.parse().map_err(|_| ParseError)?;

        let full_path = parts.next().ok_or(ParseError)?;
        let (path, query_params) = self.parse_path_and_query(full_path);

        let mut headers = HashMap::new();
        for line in lines {
            if let Some(colon_pos) = line.find(':') {
                let name = line[..colon_pos].trim().to_lowercase();
                let value = line[colon_pos + 1..].trim().to_string();
                headers.insert(name, value);
            }
        }

        Ok((method, path, query_params, headers))
    }

    fn parse_path_and_query(&self, full_path: &str) -> (String, HashMap<String, String>) {
        let mut query_params = HashMap::new();
        
        if let Some(question_pos) = full_path.find('?') {
            let path = full_path[..question_pos].to_string();
            let query_str = &full_path[question_pos + 1..];
            
            for pair in query_str.split('&') {
                if let Some(eq_pos) = pair.find('=') {
                    let key = pair[..eq_pos].to_string();
                    let value = pair[eq_pos + 1..].to_string();
                    query_params.insert(key, value);
                }
            }
            
            (path, query_params)
        } else {
            (full_path.to_string(), query_params)
        }
    }

    pub fn reset(&mut self) {
        self.buffer.clear();
        self.headers_complete = false;
        self.body_remaining = 0;
    }
}
