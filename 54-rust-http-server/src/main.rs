mod http;
mod websocket;
mod router;
mod connection;
mod server;

pub use http::{Request, Response, Method};
pub use websocket::Frame;
pub use router::Router;
pub use server::Server;

fn handler<F>(f: F) -> F
where
    F: Fn(&Request) -> Response + Send + Sync + 'static,
{
    f
}

fn main() {
    let mut server = Server::new();
    
    let router = server.router();
    
    router.get("/", handler(|_req| {
        Response::with_text(200, "Welcome to the HTTP Server!")
    }));
    
    router.get("/api/users", handler(|req| {
        let query = req.query();
        let page = query.get("page").unwrap_or(&"1".to_string()).clone();
        let size = query.get("size").unwrap_or(&"10".to_string()).clone();
        
        let json = format!(
            r#"{{"users": ["user1", "user2"], "page": {}, "size": {}}}"#,
            page, size
        );
        Response::with_json(200, &json)
    }));
    
    router.post("/api/users", handler(|req| {
        let body = String::from_utf8_lossy(&req.body);
        let json = format!(r#"{{"created": true, "data": {}}}"#, body);
        Response::with_json(201, &json)
    }));
    
    router.delete("/api/users/", handler(|req| {
        let id = req.path.trim_start_matches("/api/users/");
        let json = format!(r#"{{"deleted": true, "id": "{}"}}"#, id);
        Response::with_json(200, &json)
    }));
    
    router.get("/ws", handler(|_req| {
        Response::with_text(426, "WebSocket Upgrade Required")
    }));
    
    server.bind("127.0.0.1:8080").expect("Failed to bind");
    println!("Server listening on http://127.0.0.1:8080");
    
    server.run().expect("Server error");
}
