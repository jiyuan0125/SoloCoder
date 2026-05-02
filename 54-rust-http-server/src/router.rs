use crate::context::Context;
use crate::http::{Method, Request, Response};

pub trait Handler: Send + Sync + 'static {
    fn handle(&self, req: &Request, ctx: &Context) -> Response;
}

impl<F> Handler for F
where
    F: Fn(&Request, &Context) -> Response + Send + Sync + 'static,
{
    fn handle(&self, req: &Request, ctx: &Context) -> Response {
        self(req, ctx)
    }
}

struct RouteEntry {
    method: Method,
    prefix: String,
    handler: Box<dyn Handler>,
}

impl RouteEntry {
    fn matches(&self, method: Method, path: &str) -> bool {
        if self.method != method {
            return false;
        }
        
        if self.prefix == "/" {
            return path == "/";
        }
        
        path == &self.prefix || path.starts_with(&format!("{}/", self.prefix))
    }
}

pub struct Router {
    routes: Vec<RouteEntry>,
}

impl Router {
    pub fn new() -> Self {
        Router { routes: Vec::new() }
    }

    pub fn get<H: Handler>(&mut self, prefix: impl Into<String>, handler: H) {
        self.add_route(Method::Get, prefix, handler);
    }

    pub fn post<H: Handler>(&mut self, prefix: impl Into<String>, handler: H) {
        self.add_route(Method::Post, prefix, handler);
    }

    pub fn delete<H: Handler>(&mut self, prefix: impl Into<String>, handler: H) {
        self.add_route(Method::Delete, prefix, handler);
    }

    fn add_route<H: Handler>(&mut self, method: Method, prefix: impl Into<String>, handler: H) {
        self.routes.push(RouteEntry {
            method,
            prefix: prefix.into(),
            handler: Box::new(handler),
        });
        
        self.routes.sort_by(|a, b| b.prefix.len().cmp(&a.prefix.len()));
    }

    pub fn route(&self, req: &Request, ctx: &Context) -> Option<Response> {
        for entry in &self.routes {
            if entry.matches(req.method, &req.path) {
                return Some(entry.handler.handle(req, ctx));
            }
        }
        None
    }
}

impl Default for Router {
    fn default() -> Self {
        Router::new()
    }
}
