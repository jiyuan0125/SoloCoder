use std::collections::HashMap;
use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum StatusCode {
    SwitchingProtocols = 101,
    Ok = 200,
    NoContent = 204,
    BadRequest = 400,
    NotFound = 404,
    MethodNotAllowed = 405,
    RequestHeaderFieldsTooLarge = 431,
    PayloadTooLarge = 413,
    InternalServerError = 500,
}

impl StatusCode {
    pub fn as_u16(&self) -> u16 {
        *self as u16
    }

    pub fn reason_phrase(&self) -> &'static str {
        match self {
            StatusCode::SwitchingProtocols => "Switching Protocols",
            StatusCode::Ok => "OK",
            StatusCode::NoContent => "No Content",
            StatusCode::BadRequest => "Bad Request",
            StatusCode::NotFound => "Not Found",
            StatusCode::MethodNotAllowed => "Method Not Allowed",
            StatusCode::RequestHeaderFieldsTooLarge => "Request Header Fields Too Large",
            StatusCode::PayloadTooLarge => "Payload Too Large",
            StatusCode::InternalServerError => "Internal Server Error",
        }
    }
}

impl fmt::Display for StatusCode {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.as_u16())
    }
}

pub struct Response {
    pub status: StatusCode,
    pub headers: HashMap<String, String>,
    pub body: Vec<u8>,
}

impl Response {
    pub fn new(status: StatusCode) -> Self {
        Response {
            status,
            headers: HashMap::new(),
            body: Vec::new(),
        }
    }

    pub fn ok() -> Self {
        Response::new(StatusCode::Ok)
    }

    pub fn not_found() -> Self {
        Response::new(StatusCode::NotFound)
    }

    pub fn bad_request() -> Self {
        Response::new(StatusCode::BadRequest)
    }

    pub fn method_not_allowed() -> Self {
        Response::new(StatusCode::MethodNotAllowed)
    }

    pub fn internal_server_error() -> Self {
        Response::new(StatusCode::InternalServerError)
    }

    pub fn header(mut self, name: impl Into<String>, value: impl Into<String>) -> Self {
        self.headers.insert(name.into(), value.into());
        self
    }

    pub fn body(mut self, body: impl Into<Vec<u8>>) -> Self {
        self.body = body.into();
        self
    }

    pub fn text(body: impl Into<String>) -> Self {
        let body_str = body.into();
        Response::ok()
            .header("Content-Type", "text/plain; charset=utf-8")
            .header("Content-Length", body_str.len().to_string())
            .body(body_str)
    }

    pub fn json(body: impl Into<String>) -> Self {
        let body_str = body.into();
        Response::ok()
            .header("Content-Type", "application/json")
            .header("Content-Length", body_str.len().to_string())
            .body(body_str)
    }

    pub fn into_bytes(self) -> Vec<u8> {
        let mut bytes = Vec::new();
        
        let status_line = format!(
            "HTTP/1.1 {} {}\r\n",
            self.status.as_u16(),
            self.status.reason_phrase()
        );
        bytes.extend_from_slice(status_line.as_bytes());

        for (name, value) in &self.headers {
            let header_line = format!("{}: {}\r\n", name, value);
            bytes.extend_from_slice(header_line.as_bytes());
        }

        bytes.extend_from_slice(b"\r\n");
        bytes.extend_from_slice(&self.body);

        bytes
    }
}
