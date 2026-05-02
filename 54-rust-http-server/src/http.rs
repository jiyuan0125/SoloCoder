use std::collections::HashMap;
use std::fmt;

pub const MAX_HEADER_SIZE: usize = 8 * 1024;
pub const MAX_BODY_SIZE: usize = 1 * 1024 * 1024;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Method {
    Get,
    Post,
    Delete,
}

impl Method {
    pub fn as_str(&self) -> &'static str {
        match self {
            Method::Get => "GET",
            Method::Post => "POST",
            Method::Delete => "DELETE",
        }
    }
}

impl fmt::Display for Method {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone)]
pub struct Request {
    pub method: Method,
    pub path: String,
    pub raw_path: String,
    pub query: HashMap<String, String>,
    pub headers: HashMap<String, String>,
    pub body: Vec<u8>,
}

impl Request {
    pub fn query(&self) -> &HashMap<String, String> {
        &self.query
    }

    pub fn header(&self, name: &str) -> Option<&str> {
        self.headers.get(&name.to_lowercase()).map(|s| s.as_str())
    }

    pub fn content_length(&self) -> Option<usize> {
        self.header("content-length")
            .and_then(|v| v.parse::<usize>().ok())
    }

    pub fn is_websocket_upgrade(&self) -> bool {
        let upgrade = self.header("upgrade");
        let connection = self.header("connection");
        
        upgrade.map(|u| u.to_lowercase().contains("websocket")).unwrap_or(false)
            && connection.map(|c| c.to_lowercase().contains("upgrade")).unwrap_or(false)
            && self.header("sec-websocket-key").is_some()
    }

    pub fn keep_alive(&self) -> bool {
        match self.header("connection") {
            Some("keep-alive") => true,
            Some("close") => false,
            None => true,
            Some(v) => v.to_lowercase().contains("keep-alive"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Response {
    pub status_code: u16,
    pub status_text: &'static str,
    pub headers: HashMap<String, String>,
    pub body: Vec<u8>,
}

impl Response {
    pub fn new(status_code: u16) -> Self {
        let status_text = match status_code {
            100 => "Continue",
            101 => "Switching Protocols",
            200 => "OK",
            201 => "Created",
            204 => "No Content",
            400 => "Bad Request",
            403 => "Forbidden",
            404 => "Not Found",
            405 => "Method Not Allowed",
            413 => "Payload Too Large",
            431 => "Request Header Fields Too Large",
            500 => "Internal Server Error",
            _ => "Unknown",
        };

        Response {
            status_code,
            status_text,
            headers: HashMap::new(),
            body: Vec::new(),
        }
    }

    pub fn with_body(status_code: u16, body: Vec<u8>) -> Self {
        let mut resp = Response::new(status_code);
        resp.headers.insert("Content-Length".to_string(), body.len().to_string());
        resp.body = body;
        resp
    }

    pub fn with_text(status_code: u16, text: &str) -> Self {
        let mut resp = Response::with_body(status_code, text.as_bytes().to_vec());
        resp.headers.insert("Content-Type".to_string(), "text/plain; charset=utf-8".to_string());
        resp
    }

    pub fn with_json(status_code: u16, json: &str) -> Self {
        let mut resp = Response::with_body(status_code, json.as_bytes().to_vec());
        resp.headers.insert("Content-Type".to_string(), "application/json; charset=utf-8".to_string());
        resp
    }

    pub fn header(&mut self, name: &str, value: &str) -> &mut Self {
        self.headers.insert(name.to_string(), value.to_string());
        self
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut result = Vec::new();
        
        result.extend_from_slice(format!("HTTP/1.1 {} {}\r\n", self.status_code, self.status_text).as_bytes());
        
        for (name, value) in &self.headers {
            result.extend_from_slice(format!("{}: {}\r\n", name, value).as_bytes());
        }
        
        result.extend_from_slice(b"\r\n");
        result.extend_from_slice(&self.body);
        
        result
    }
}

pub enum ParseError {
    Incomplete,
    HeaderTooLarge,
    InvalidRequest,
    InvalidMethod,
}

pub fn parse_request_line_and_headers(buffer: &[u8]) -> Result<(Request, usize), ParseError> {
    let mut header_end = None;
    for i in 0..buffer.len().saturating_sub(3) {
        if buffer[i] == b'\r' && buffer[i + 1] == b'\n' 
            && buffer[i + 2] == b'\r' && buffer[i + 3] == b'\n' {
            header_end = Some(i + 4);
            break;
        }
    }
    
    let header_end = match header_end {
        Some(end) => end,
        None => {
            if buffer.len() > MAX_HEADER_SIZE {
                return Err(ParseError::HeaderTooLarge);
            }
            return Err(ParseError::Incomplete);
        }
    };
    
    if header_end > MAX_HEADER_SIZE {
        return Err(ParseError::HeaderTooLarge);
    }
    
    let header_bytes = &buffer[..header_end];
    let header_str = match String::from_utf8_lossy(header_bytes) {
        s => s,
    };
    
    let mut lines = header_str.split("\r\n");
    let request_line = lines.next().ok_or(ParseError::InvalidRequest)?;
    
    let mut parts = request_line.split_whitespace();
    let method_str = parts.next().ok_or(ParseError::InvalidRequest)?;
    let raw_path = parts.next().ok_or(ParseError::InvalidRequest)?;
    let _version = parts.next().ok_or(ParseError::InvalidRequest)?;
    
    let method = match method_str {
        "GET" => Method::Get,
        "POST" => Method::Post,
        "DELETE" => Method::Delete,
        _ => return Err(ParseError::InvalidMethod),
    };
    
    let (path, query) = parse_path_and_query(raw_path);
    
    let mut headers = HashMap::new();
    for line in lines {
        if line.is_empty() {
            continue;
        }
        if let Some(colon_pos) = line.find(':') {
            let name = line[..colon_pos].trim().to_lowercase();
            let value = line[colon_pos + 1..].trim().to_string();
            headers.insert(name, value);
        }
    }
    
    Ok((Request {
        method,
        path,
        raw_path: raw_path.to_string(),
        query,
        headers,
        body: Vec::new(),
    }, header_end))
}

fn parse_path_and_query(raw_path: &str) -> (String, HashMap<String, String>) {
    let mut query = HashMap::new();
    
    if let Some(query_start) = raw_path.find('?') {
        let path = raw_path[..query_start].to_string();
        let query_str = &raw_path[query_start + 1..];
        
        for pair in query_str.split('&') {
            if let Some(eq_pos) = pair.find('=') {
                let key = percent_decode(&pair[..eq_pos]);
                let value = percent_decode(&pair[eq_pos + 1..]);
                query.insert(key, value);
            } else if !pair.is_empty() {
                query.insert(percent_decode(pair), String::new());
            }
        }
        
        (path, query)
    } else {
        (raw_path.to_string(), query)
    }
}

fn percent_decode(s: &str) -> String {
    let mut result = String::new();
    let mut chars = s.chars();
    
    while let Some(c) = chars.next() {
        if c == '%' {
            let hex: String = chars.by_ref().take(2).collect();
            if hex.len() == 2 {
                if let Ok(byte) = u8::from_str_radix(&hex, 16) {
                    result.push(byte as char);
                    continue;
                }
            }
            result.push('%');
            result.push_str(&hex);
        } else if c == '+' {
            result.push(' ');
        } else {
            result.push(c);
        }
    }
    
    result
}
