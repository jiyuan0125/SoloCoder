use crate::connection::{self, ConnectionContext, ConnectionResult, WebSocketClients};
use crate::router::Router;
use crate::websocket::Frame;
use std::io::{self, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::{Arc, Mutex, atomic::{AtomicUsize, Ordering}};
use std::thread;
use std::time::Duration;

pub const GRACEFUL_SHUTDOWN_TIMEOUT: u64 = 5;

static NEXT_CLIENT_ID: AtomicUsize = AtomicUsize::new(1);

pub struct Server {
    router: Arc<Router>,
    ws_clients: WebSocketClients,
    shutdown_signal: Arc<Mutex<bool>>,
    active_connections: Arc<AtomicUsize>,
    listener: Option<TcpListener>,
}

impl Server {
    pub fn new() -> Self {
        Server {
            router: Arc::new(Router::new()),
            ws_clients: Arc::new(Mutex::new(std::collections::HashMap::new())),
            shutdown_signal: Arc::new(Mutex::new(false)),
            active_connections: Arc::new(AtomicUsize::new(0)),
            listener: None,
        }
    }

    pub fn router(&mut self) -> &mut Router {
        Arc::get_mut(&mut self.router).expect("Router is already shared")
    }

    pub fn broadcast(&self, message: &str) {
        let frame = Frame::text(message.to_string());
        let frame_bytes = frame.encode();
        
        let clients = self.ws_clients.lock().unwrap();
        for (_, client_stream) in clients.iter() {
            if let Ok(mut stream) = client_stream.lock() {
                let _ = stream.write_all(&frame_bytes);
                let _ = stream.flush();
            }
        }
    }

    pub fn broadcast_binary(&self, data: &[u8]) {
        let frame = Frame::binary(data.to_vec());
        let frame_bytes = frame.encode();
        
        let clients = self.ws_clients.lock().unwrap();
        for (_, client_stream) in clients.iter() {
            if let Ok(mut stream) = client_stream.lock() {
                let _ = stream.write_all(&frame_bytes);
                let _ = stream.flush();
            }
        }
    }

    pub fn bind(&mut self, addr: &str) -> io::Result<()> {
        let listener = TcpListener::bind(addr)?;
        listener.set_nonblocking(true)?;
        self.listener = Some(listener);
        Ok(())
    }

    pub fn run(&mut self) -> io::Result<()> {
        let listener = self.listener.take().expect("Server not bound");
        
        println!("Server starting...");
        
        let shutdown_signal = self.shutdown_signal.clone();
        let active_connections = self.active_connections.clone();
        
        ctrlc::set_handler(move || {
            println!("\nShutting down gracefully...");
            *shutdown_signal.lock().unwrap() = true;
        }).expect("Error setting Ctrl-C handler");
        
        for stream_result in listener.incoming() {
            if *self.shutdown_signal.lock().unwrap() {
                break;
            }
            
            match stream_result {
                Ok(stream) => {
                    self.handle_connection(stream);
                }
                Err(ref e) if e.kind() == io::ErrorKind::WouldBlock => {
                    thread::sleep(Duration::from_millis(10));
                    continue;
                }
                Err(e) => {
                    eprintln!("Accept error: {}", e);
                    continue;
                }
            }
        }
        
        println!("Waiting for active connections to finish...");
        let start = std::time::Instant::now();
        
        while active_connections.load(Ordering::SeqCst) > 0 {
            if start.elapsed().as_secs() >= GRACEFUL_SHUTDOWN_TIMEOUT {
                println!("Graceful shutdown timeout, forcing exit...");
                break;
            }
            thread::sleep(Duration::from_millis(100));
        }
        
        println!("Server stopped.");
        Ok(())
    }

    fn handle_connection(&self, stream: TcpStream) {
        let router = self.router.clone();
        let ws_clients = self.ws_clients.clone();
        let shutdown_signal = self.shutdown_signal.clone();
        let active_connections = self.active_connections.clone();
        let client_id = NEXT_CLIENT_ID.fetch_add(1, Ordering::SeqCst);
        
        active_connections.fetch_add(1, Ordering::SeqCst);
        
        thread::spawn(move || {
            let result = Self::process_connection(
                stream,
                router,
                ws_clients,
                shutdown_signal,
                client_id,
            );
            
            active_connections.fetch_sub(1, Ordering::SeqCst);
            
            if let Err(e) = result {
                eprintln!("Connection error: {}", e);
            }
        });
    }

    fn process_connection(
        stream: TcpStream,
        router: Arc<Router>,
        ws_clients: WebSocketClients,
        shutdown_signal: Arc<Mutex<bool>>,
        client_id: usize,
    ) -> io::Result<()> {
        let mut ctx = ConnectionContext {
            stream: stream.try_clone()?,
            router,
            ws_clients,
            client_id,
            shutdown_signal,
        };
        
        match connection::handle_http_connection(&mut ctx) {
            Ok(ConnectionResult::Close) => {}
            Ok(ConnectionResult::UpgradedToWebSocket) => {
                connection::handle_websocket_connection(&mut ctx)?;
            }
            Err(e) => {
                if e.kind() != io::ErrorKind::TimedOut && e.kind() != io::ErrorKind::WouldBlock {
                    return Err(e);
                }
            }
        }
        
        Ok(())
    }
}

impl Default for Server {
    fn default() -> Self {
        Server::new()
    }
}
