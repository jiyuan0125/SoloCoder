mod http;
mod websocket;
mod router;
mod connection;

use std::net::TcpListener;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;
use std::io;

use crate::http::{Request, Response, StatusCode};
use crate::router::Router;
use crate::connection::ConnectionManager;

static SHUTDOWN: AtomicBool = AtomicBool::new(false);

extern "C" fn handle_sigterm(_signum: i32) {
    SHUTDOWN.store(true, Ordering::SeqCst);
}

fn register_signal_handler() {
    unsafe {
        let sigaction = libc::sigaction {
            sa_sigaction: handle_sigterm as *const () as usize,
            sa_mask: std::mem::zeroed(),
            sa_flags: 0,
            #[cfg(target_os = "linux")]
            sa_restorer: None,
        };
        
        libc::sigaction(libc::SIGTERM, &sigaction, std::ptr::null_mut());
        libc::sigaction(libc::SIGINT, &sigaction, std::ptr::null_mut());
    }
}

fn main() -> io::Result<()> {
    register_signal_handler();

    let mut router = Router::new();
    
    router.get("/", |_req: &Request| -> Response {
        Response::text("Welcome to Rust HTTP Server!")
    });
    
    router.get("/api/users", |req: &Request| -> Response {
        let page = req.query().get("page").cloned().unwrap_or_else(|| "1".to_string());
        let size = req.query().get("size").cloned().unwrap_or_else(|| "10".to_string());
        
        let json = format!(
            r#"{{"page": {}, "size": {}, "users": ["user1", "user2", "user3"]}}"#,
            page, size
        );
        Response::json(json)
    });
    
    router.post("/api/users", |req: &Request| -> Response {
        let body_str = String::from_utf8_lossy(&req.body);
        let json = format!(r#"{{"status": "created", "received": {}}}"#, body_str);
        Response::new(StatusCode::Ok)
            .header("Content-Type", "application/json")
            .header("Content-Length", json.len().to_string())
            .body(json)
    });
    
    router.delete("/api/users/", |req: &Request| -> Response {
        let json = format!(r#"{{"status": "deleted", "path": "{}"}}"#, req.path);
        Response::json(json)
    });
    
    router.get("/ws", |_req: &Request| -> Response {
        Response::text("WebSocket endpoint - connect with WebSocket client")
    });

    let listener = TcpListener::bind("127.0.0.1:8080")?;
    listener.set_nonblocking(true)?;
    
    println!("Server listening on 127.0.0.1:8080");
    println!("Press Ctrl+C to shutdown");

    let manager = Arc::new(Mutex::new(ConnectionManager::new(router)));
    let mut threads = Vec::new();

    loop {
        if SHUTDOWN.load(Ordering::SeqCst) {
            println!("\nShutting down server...");
            break;
        }

        match listener.accept() {
            Ok((stream, addr)) => {
                println!("New connection from: {}", addr);
                let manager_clone = Arc::clone(&manager);
                
                let handle = thread::spawn(move || {
                    let mut manager = manager_clone.lock().unwrap();
                    manager.handle_connection(stream);
                });
                threads.push(handle);
            }
            Err(ref e) if e.kind() == io::ErrorKind::WouldBlock => {
                thread::sleep(Duration::from_millis(10));
            }
            Err(e) => {
                eprintln!("Accept error: {}", e);
                break;
            }
        }
    }

    println!("Waiting for active connections to finish (max 5 seconds)...");
    
    let start = std::time::Instant::now();
    while !threads.is_empty() && start.elapsed() < Duration::from_secs(5) {
        threads.retain(|handle| !handle.is_finished());
        thread::sleep(Duration::from_millis(100));
    }

    if !threads.is_empty() {
        println!("Force shutdown after 5 second timeout");
    } else {
        println!("All connections finished, exiting gracefully");
    }

    Ok(())
}

#[allow(dead_code)]
fn broadcast_example() {
    // 这个函数展示如何从其他地方调用 broadcast
    // 实际使用时，需要通过 Arc<Mutex<ConnectionManager>> 访问
    // manager.broadcast(b"Hello all clients!".to_vec());
}
