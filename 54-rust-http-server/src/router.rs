use crate::http::{Method, Request, Response};

pub trait Handler: Send + Sync + 'static {
    fn handle(&self, req: &Request) -> Response;
}

impl<F> Handler for F
where
    F: Fn(&Request) -> Response + Send + Sync + 'static,
{
    fn handle(&self, req: &Request) -> Response {
        self(req)
    }
}

struct RouteEntry {
    method: Method,
    prefix: String,
    handler: Box<dyn Handler>,
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
    }

    pub fn route(&self, req: &Request) -> Option<Response> {
        for entry in &self.routes {
            if entry.method == req.method && req.path.starts_with(&entry.prefix) {
                return Some(entry.handler.handle(req));
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
