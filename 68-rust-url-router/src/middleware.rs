use crate::{Handler, Params, Request, Response};

pub type MiddlewareFn = fn(&mut Request, &mut Next) -> Response;

pub struct Next {
    handler: Handler,
    params: Params,
    remaining: Vec<MiddlewareFn>,
}

impl Next {
    pub fn run(&mut self, req: &mut Request) -> Response {
        if let Some(middleware) = self.remaining.pop() {
            middleware(req, self)
        } else {
            (self.handler)(req, &self.params)
        }
    }
}

#[derive(Debug, Clone)]
pub struct MiddlewareChain {
    middlewares: Vec<MiddlewareFn>,
}

impl MiddlewareChain {
    pub fn new() -> Self {
        MiddlewareChain {
            middlewares: Vec::new(),
        }
    }

    pub fn add(&mut self, middleware: MiddlewareFn) {
        self.middlewares.push(middleware);
    }

    pub fn middlewares(&self) -> &[MiddlewareFn] {
        &self.middlewares
    }

    pub fn execute(&self, req: &mut Request, handler: Handler, params: Params) -> Response {
        let mut reversed = self.middlewares.clone();
        reversed.reverse();

        let mut next = Next {
            handler,
            params,
            remaining: reversed,
        };

        next.run(req)
    }

    pub fn len(&self) -> usize {
        self.middlewares.len()
    }

    pub fn is_empty(&self) -> bool {
        self.middlewares.is_empty()
    }
}

impl Default for MiddlewareChain {
    fn default() -> Self {
        Self::new()
    }
}

pub struct Middleware;

impl Middleware {
    pub fn new(f: MiddlewareFn) -> MiddlewareFn {
        f
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{Method, Params, Request, Response};

    fn test_handler(_req: &Request, _params: &Params) -> Response {
        Response::ok("handler called")
    }

    fn logging_middleware(req: &mut Request, next: &mut Next) -> Response {
        let method = req.method.as_str().to_string();
        let path = req.path.clone();
        let mut resp = next.run(req);
        resp.body = format!("[LOG] {} {} - {}", method, path, resp.body);
        resp
    }

    fn auth_middleware(req: &mut Request, next: &mut Next) -> Response {
        let has_auth = req.headers.iter().any(|(k, _)| k == "Authorization");
        if has_auth {
            next.run(req)
        } else {
            Response::new(401, "Unauthorized")
        }
    }

    #[test]
    fn test_middleware_chain_executes_handler() {
        let mut chain = MiddlewareChain::new();
        let mut req = Request::new(Method::GET, "/test");

        let resp = chain.execute(&mut req, test_handler, Params::new());
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "handler called");
    }

    #[test]
    fn test_middleware_before_and_after() {
        let mut chain = MiddlewareChain::new();
        chain.add(logging_middleware);

        let mut req = Request::new(Method::GET, "/test");

        let resp = chain.execute(&mut req, test_handler, Params::new());
        assert_eq!(resp.status, 200);
        assert!(resp.body.starts_with("[LOG]"));
        assert!(resp.body.contains("handler called"));
    }

    #[test]
    fn test_middleware_skips_handler() {
        let mut chain = MiddlewareChain::new();
        chain.add(auth_middleware);

        let mut req = Request::new(Method::GET, "/protected");

        let resp = chain.execute(&mut req, test_handler, Params::new());
        assert_eq!(resp.status, 401);
        assert_eq!(resp.body, "Unauthorized");
    }

    #[test]
    fn test_middleware_passes_auth() {
        let mut chain = MiddlewareChain::new();
        chain.add(auth_middleware);

        let mut req = Request::new(Method::GET, "/protected");
        req.headers.push(("Authorization".to_string(), "Bearer token".to_string()));

        let resp = chain.execute(&mut req, test_handler, Params::new());
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "handler called");
    }

    #[test]
    fn test_multiple_middlewares() {
        let mut chain = MiddlewareChain::new();
        chain.add(logging_middleware);
        chain.add(auth_middleware);

        let mut req = Request::new(Method::GET, "/protected");
        req.headers.push(("Authorization".to_string(), "Bearer token".to_string()));

        let resp = chain.execute(&mut req, test_handler, Params::new());
        assert_eq!(resp.status, 200);
        assert!(resp.body.starts_with("[LOG]"));
        assert!(resp.body.contains("handler called"));
    }
}
