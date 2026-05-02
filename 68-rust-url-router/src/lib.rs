pub mod middleware;
pub mod params;
pub mod radix_tree;
pub mod route;
pub mod router;

pub use middleware::{Middleware, Next};
pub use params::Params;
pub use radix_tree::RadixTree;
pub use route::{PathParser, Segment};
pub use router::Router;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum Method {
    GET,
    POST,
    PUT,
    DELETE,
    PATCH,
    HEAD,
    OPTIONS,
    CONNECT,
    TRACE,
}

impl Method {
    pub fn as_str(&self) -> &'static str {
        match self {
            Method::GET => "GET",
            Method::POST => "POST",
            Method::PUT => "PUT",
            Method::DELETE => "DELETE",
            Method::PATCH => "PATCH",
            Method::HEAD => "HEAD",
            Method::OPTIONS => "OPTIONS",
            Method::CONNECT => "CONNECT",
            Method::TRACE => "TRACE",
        }
    }
}

#[derive(Debug, Clone)]
pub struct Request {
    pub method: Method,
    pub path: String,
    pub headers: Vec<(String, String)>,
}

impl Request {
    pub fn new(method: Method, path: &str) -> Self {
        Request {
            method,
            path: path.to_string(),
            headers: Vec::new(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Response {
    pub status: u16,
    pub body: String,
    pub headers: Vec<(String, String)>,
}

impl Response {
    pub fn new(status: u16, body: &str) -> Self {
        Response {
            status,
            body: body.to_string(),
            headers: Vec::new(),
        }
    }

    pub fn ok(body: &str) -> Self {
        Response::new(200, body)
    }

    pub fn not_found() -> Self {
        Response::new(404, "Not Found")
    }
}

pub type Handler = fn(&Request, &Params) -> Response;

#[derive(Debug, Clone, thiserror::Error)]
pub enum RouterError {
    #[error("Invalid path format: {0}")]
    InvalidPathFormat(String),
    #[error("Route already exists for method {method:?} and path {path}")]
    RouteAlreadyExists { method: Method, path: String },
    #[error("Wildcard `*` can only appear at the end of the path")]
    WildcardNotAtEnd,
    #[error("Wildcard `**` can only appear at the end of the path")]
    DoubleWildcardNotAtEnd,
    #[error("Cannot have both `*` and `**` wildcards in the same path")]
    ConflictingWildcards,
}
