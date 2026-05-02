use std::collections::HashMap;

use crate::{
    middleware::{MiddlewareChain, MiddlewareFn},
    radix_tree::RadixTree,
    Handler, Method, Request, Response, RouterError,
};

pub struct Router {
    trees: HashMap<Method, RadixTree>,
    middlewares: MiddlewareChain,
    global_middlewares: MiddlewareChain,
}

impl Router {
    pub fn new() -> Self {
        Router {
            trees: HashMap::new(),
            middlewares: MiddlewareChain::new(),
            global_middlewares: MiddlewareChain::new(),
        }
    }

    pub fn get(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::GET, path, handler)
    }

    pub fn post(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::POST, path, handler)
    }

    pub fn put(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::PUT, path, handler)
    }

    pub fn delete(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::DELETE, path, handler)
    }

    pub fn patch(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::PATCH, path, handler)
    }

    pub fn head(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::HEAD, path, handler)
    }

    pub fn options(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::OPTIONS, path, handler)
    }

    pub fn connect(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::CONNECT, path, handler)
    }

    pub fn trace(&mut self, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        self.register(Method::TRACE, path, handler)
    }

    fn register(&mut self, method: Method, path: &str, handler: Handler) -> Result<&mut Self, RouterError> {
        let handler_name = std::any::type_name::<Handler>();

        let tree = self.trees.entry(method.clone()).or_insert_with(RadixTree::new);

        tree.insert(path, method, handler, handler_name)?;

        Ok(self)
    }

    pub fn middleware(&mut self, middleware: MiddlewareFn) -> &mut Self {
        self.middlewares.add(middleware);
        self
    }

    pub fn global_middleware(&mut self, middleware: MiddlewareFn) -> &mut Self {
        self.global_middlewares.add(middleware);
        self
    }

    pub fn handle(&self, req: &mut Request) -> Response {
        let tree = match self.trees.get(&req.method) {
            Some(t) => t,
            None => return Response::not_found(),
        };

        let match_result = tree.lookup(&req.path, &req.method);

        match match_result {
            Some((handler_info, params)) => {
                let handler = handler_info.handler;

                if !self.global_middlewares.is_empty() {
                    let mut combined = MiddlewareChain::new();
                    for m in self.global_middlewares.middlewares() {
                        combined.add(*m);
                    }
                    for m in self.middlewares.middlewares() {
                        combined.add(*m);
                    }
                    combined.execute(req, handler, params)
                } else if !self.middlewares.is_empty() {
                    self.middlewares.execute(req, handler, params)
                } else {
                    handler(req, &params)
                }
            }
            None => Response::not_found(),
        }
    }

    pub fn routes(&self) -> Vec<RouteInfo> {
        let mut routes = Vec::new();

        for (method, tree) in &self.trees {
            for entry in tree.list_routes() {
                routes.push(RouteInfo {
                    method: method.clone(),
                    path: entry.path,
                    handler_name: entry.handler_name,
                });
            }
        }

        routes.sort_by(|a, b| {
            a.method.as_str().cmp(b.method.as_str())
                .then_with(|| a.path.cmp(&b.path))
        });

        routes
    }
}

impl Default for Router {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug, Clone)]
pub struct RouteInfo {
    pub method: Method,
    pub path: String,
    pub handler_name: &'static str,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{Method, Params, Request, Response};

    fn root_handler(_req: &Request, _params: &Params) -> Response {
        Response::ok("Welcome")
    }

    fn users_handler(_req: &Request, _params: &Params) -> Response {
        Response::ok("Users list")
    }

    fn user_handler(_req: &Request, params: &Params) -> Response {
        let id = params.get_str("id").unwrap_or("");
        Response::ok(&format!("User: {}", id))
    }

    fn post_handler(_req: &Request, params: &Params) -> Response {
        let user_id = params.get_str("id").unwrap_or("");
        let post_id = params.get_str("post_id").unwrap_or("");
        Response::ok(&format!("Post {}/{}", user_id, post_id))
    }

    fn files_handler(_req: &Request, params: &Params) -> Response {
        let file = params.get_str("*").unwrap_or("");
        Response::ok(&format!("File: {}", file))
    }

    fn static_handler(_req: &Request, params: &Params) -> Response {
        let path = params.get_str("**").unwrap_or("");
        Response::ok(&format!("Static: {}", path))
    }

    #[test]
    fn test_router_basic_routing() {
        let mut router = Router::new();
        router.get("/", root_handler).unwrap();
        router.get("/users", users_handler).unwrap();

        let mut req = Request::new(Method::GET, "/");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "Welcome");

        let mut req = Request::new(Method::GET, "/users");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "Users list");
    }

    #[test]
    fn test_router_path_params() {
        let mut router = Router::new();
        router.get("/users/:id", user_handler).unwrap();
        router.get("/users/:id/posts/:post_id", post_handler).unwrap();

        let mut req = Request::new(Method::GET, "/users/42");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "User: 42");

        let mut req = Request::new(Method::GET, "/users/42/posts/10");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "Post 42/10");
    }

    #[test]
    fn test_router_wildcard() {
        let mut router = Router::new();
        router.get("/files/*", files_handler).unwrap();

        let mut req = Request::new(Method::GET, "/files/test.txt");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "File: test.txt");

        let mut req = Request::new(Method::GET, "/files/a/b.txt");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 404);
    }

    #[test]
    fn test_router_double_wildcard() {
        let mut router = Router::new();
        router.get("/static/**", static_handler).unwrap();

        let mut req = Request::new(Method::GET, "/static/css/style.css");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "Static: css/style.css");

        let mut req = Request::new(Method::GET, "/static/js/app.js");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 200);
        assert_eq!(resp.body, "Static: js/app.js");
    }

    #[test]
    fn test_router_different_methods() {
        let mut router = Router::new();
        router.get("/users", users_handler).unwrap();
        router.post("/users", root_handler).unwrap();

        let mut req = Request::new(Method::GET, "/users");
        let resp = router.handle(&mut req);
        assert_eq!(resp.body, "Users list");

        let mut req = Request::new(Method::POST, "/users");
        let resp = router.handle(&mut req);
        assert_eq!(resp.body, "Welcome");
    }

    #[test]
    fn test_router_not_found() {
        let router = Router::new();

        let mut req = Request::new(Method::GET, "/nonexistent");
        let resp = router.handle(&mut req);
        assert_eq!(resp.status, 404);
    }

    #[test]
    fn test_router_routes_list() {
        let mut router = Router::new();
        router.get("/", root_handler).unwrap();
        router.get("/users", users_handler).unwrap();
        router.get("/users/:id", user_handler).unwrap();

        let routes = router.routes();
        assert_eq!(routes.len(), 3);

        let paths: Vec<_> = routes.iter().map(|r| r.path.as_str()).collect();
        assert!(paths.contains(&"/"));
        assert!(paths.contains(&"/users"));
        assert!(paths.contains(&"/users/:id"));
    }

    #[test]
    fn test_router_invalid_path() {
        let mut router = Router::new();

        let result = router.get("//invalid", root_handler);
        assert!(result.is_err());

        let result = router.get("/users/:", root_handler);
        assert!(result.is_err());

        let result = router.get("/files/*/extra", root_handler);
        assert!(result.is_err());
    }

    #[test]
    fn test_router_duplicate_route() {
        let mut router = Router::new();
        router.get("/users", users_handler).unwrap();

        let result = router.get("/users", root_handler);
        assert!(result.is_err());
    }
}
