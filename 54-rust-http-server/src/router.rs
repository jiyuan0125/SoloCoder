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
    path_prefix: String,
    handler: Box<dyn Handler>,
}

pub struct Router {
    routes: Vec<RouteEntry>,
}

impl Router {
    pub fn new() -> Self {
        Router { routes: Vec::new() }
    }

    pub fn get<F>(&mut self, path_prefix: &str, handler: F)
    where
        F: Handler,
    {
        self.add_route(Method::Get, path_prefix, handler);
    }

    pub fn post<F>(&mut self, path_prefix: &str, handler: F)
    where
        F: Handler,
    {
        self.add_route(Method::Post, path_prefix, handler);
    }

    pub fn delete<F>(&mut self, path_prefix: &str, handler: F)
    where
        F: Handler,
    {
        self.add_route(Method::Delete, path_prefix, handler);
    }

    fn add_route<F>(&mut self, method: Method, path_prefix: &str, handler: F)
    where
        F: Handler,
    {
        self.routes.push(RouteEntry {
            method,
            path_prefix: path_prefix.to_string(),
            handler: Box::new(handler),
        });
        self.routes.sort_by(|a, b| {
            b.path_prefix.len().cmp(&a.path_prefix.len())
        });
    }

    pub fn match_route(&self, method: Method, path: &str) -> Option<&dyn Handler> {
        for route in &self.routes {
            if route.method != method {
                continue;
            }
            
            if path_matches_prefix(path, &route.path_prefix) {
                return Some(&*route.handler);
            }
        }
        None
    }

    pub fn handle(&self, req: &Request) -> Response {
        if let Some(handler) = self.match_route(req.method, &req.path) {
            handler.handle(req)
        } else {
            Response::with_text(404, "Not Found")
        }
    }
}

impl Default for Router {
    fn default() -> Self {
        Router::new()
    }
}

fn path_matches_prefix(path: &str, prefix: &str) -> bool {
    if path == prefix {
        return true;
    }
    
    if path.starts_with(prefix) {
        let next_char = path.chars().nth(prefix.len());
        return next_char == Some('/');
    }
    
    false
}
