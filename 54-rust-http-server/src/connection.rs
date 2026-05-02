use std::collections::HashMap;
use std::io::{self, Read, Write};
use std::net::TcpStream;
use std::sync::mpsc::{self, Sender, Receiver};
use std::sync::{Arc, Mutex};
use std::time::{Instant, Duration, SystemTime};

use crate::http::{Request, RequestParser, Response, StatusCode, ParseResult, Method, MAX_BODY_SIZE, MAX_HEADER_SIZE};
use crate::websocket::{Frame, FrameParser, OpCode, ParseFrameResult, encode_frame, compute_accept_key, MAX_PAYLOAD_SIZE, PING_INTERVAL_SECS, PONG_TIMEOUT_SECS};
use crate::router::Router;

pub type BroadcastSender = Sender<Vec<u8>>;
pub type BroadcastReceiver = Receiver<Vec<u8>>;

pub struct ConnectionManager {
    pub router: Arc<Router>,
    pub ws_connections: Arc<Mutex<HashMap<usize, BroadcastSender>>>,
    next_id: usize,
}

impl ConnectionManager {
    pub fn new(router: Router) -> Self {
        ConnectionManager {
            router: Arc::new(router),
            ws_connections: Arc::new(Mutex::new(HashMap::new())),
            next_id: 1,
        }
    }

    pub fn broadcast(&self, message: Vec<u8>) {
        let connections = self.ws_connections.lock().unwrap();
        for sender in connections.values() {
            let _ = sender.send(message.clone());
        }
    }

    pub fn handle_connection(&mut self, mut stream: TcpStream) {
        let _ = stream.set_read_timeout(Some(Duration::from_secs(1)));
        let _ = stream.set_nodelay(true);

        let mut parser = RequestParser::new();
        let mut buffer = [0u8; 8192];
        let mut keep_alive = true;

        while keep_alive {
            match stream.read(&mut buffer) {
                Ok(0) => break,
                Ok(n) => {
                    parser.consume(&buffer[..n]);
                    
                    loop {
                        match parser.parse() {
                            ParseResult::Complete(req) => {
                                let start = Instant::now();
                                let method = req.method.clone();
                                let path = req.path.clone();
                                
                                if req.is_websocket_upgrade() {
                                    let resp = self.handle_websocket_upgrade(&req);
                                    let status = resp.status;
                                    let _ = stream.write_all(&resp.into_bytes());
                                    let _ = stream.flush();
                                    
                                    let duration = start.elapsed();
                                    log_request(method, &path, status, duration);
                                    
                                    if status == StatusCode::SwitchingProtocols {
                                        self.handle_websocket(stream);
                                        return;
                                    }
                                    
                                    keep_alive = req.connection_keep_alive();
                                    parser.reset();
                                    break;
                                } else {
                                    let resp = self.handle_http_request(&req);
                                    let status = resp.status;
                                    let _ = stream.write_all(&resp.into_bytes());
                                    let _ = stream.flush();
                                    
                                    let duration = start.elapsed();
                                    log_request(method, &path, status, duration);
                                    
                                    keep_alive = req.connection_keep_alive();
                                    parser.reset();
                                }
                            }
                            ParseResult::Partial => break,
                            ParseResult::TooLarge => {
                                let start = Instant::now();
                                let resp = Response::new(StatusCode::RequestHeaderFieldsTooLarge)
                                    .header("Content-Type", "text/plain")
                                    .header("Content-Length", "31")
                                    .body("Request Header Fields Too Large");
                                let status = resp.status;
                                let _ = stream.write_all(&resp.into_bytes());
                                let _ = stream.flush();
                                
                                let duration = start.elapsed();
                                log_request(Method::Get, "", status, duration);
                                
                                keep_alive = false;
                                break;
                            }
                            ParseResult::Error(_) => {
                                let start = Instant::now();
                                let resp = Response::bad_request()
                                    .header("Content-Type", "text/plain")
                                    .header("Content-Length", "11")
                                    .body("Bad Request");
                                let status = resp.status;
                                let _ = stream.write_all(&resp.into_bytes());
                                let _ = stream.flush();
                                
                                let duration = start.elapsed();
                                log_request(Method::Get, "", status, duration);
                                
                                keep_alive = false;
                                break;
                            }
                        }
                    }
                }
                Err(ref e) if e.kind() == io::ErrorKind::WouldBlock || e.kind() == io::ErrorKind::TimedOut => {
                    continue;
                }
                Err(_) => break,
            }
        }
    }

    fn handle_http_request(&self, req: &Request) -> Response {
        if let Some(content_len) = req.content_length() {
            if content_len > MAX_BODY_SIZE {
                return Response::new(StatusCode::PayloadTooLarge)
                    .header("Content-Type", "text/plain")
                    .header("Content-Length", "17")
                    .body("Payload Too Large");
            }
        }

        match self.router.route(req) {
            Some(resp) => resp,
            None => Response::not_found()
                .header("Content-Type", "text/plain")
                .header("Content-Length", "9")
                .body("Not Found"),
        }
    }

    fn handle_websocket_upgrade(&self, req: &Request) -> Response {
        let sec_ws_key = match req.header("sec-websocket-key") {
            Some(key) => key,
            None => {
                return Response::bad_request()
                    .header("Content-Type", "text/plain")
                    .body("Missing Sec-WebSocket-Key");
            }
        };

        let accept_key = compute_accept_key(sec_ws_key);

        Response {
            status: StatusCode::SwitchingProtocols,
            headers: [
                ("Upgrade".to_string(), "websocket".to_string()),
                ("Connection".to_string(), "Upgrade".to_string()),
                ("Sec-WebSocket-Accept".to_string(), accept_key),
            ].iter().cloned().collect(),
            body: Vec::new(),
        }
    }

    fn handle_websocket(&mut self, mut stream: TcpStream) {
        let conn_id = self.next_id;
        self.next_id += 1;

        let (tx, rx): (BroadcastSender, BroadcastReceiver) = mpsc::channel();
        
        {
            let mut connections = self.ws_connections.lock().unwrap();
            connections.insert(conn_id, tx);
        }

        let mut ws_stream = WebSocketStream::new(stream, rx);
        let _ = ws_stream.run();

        {
            let mut connections = self.ws_connections.lock().unwrap();
            connections.remove(&conn_id);
        }
    }
}

struct WebSocketStream {
    stream: TcpStream,
    parser: FrameParser,
    broadcast_rx: BroadcastReceiver,
    last_pong: Instant,
    last_ping: Instant,
}

impl WebSocketStream {
    fn new(stream: TcpStream, broadcast_rx: BroadcastReceiver) -> Self {
        WebSocketStream {
            stream,
            parser: FrameParser::new(),
            broadcast_rx,
            last_pong: Instant::now(),
            last_ping: Instant::now(),
        }
    }

    fn run(&mut self) -> io::Result<()> {
        let mut buffer = [0u8; 8192];

        loop {
            let now = Instant::now();

            if now.duration_since(self.last_ping) > Duration::from_secs(PING_INTERVAL_SECS) {
                let ping_frame = Frame::ping(Vec::new());
                self.send_frame(&ping_frame)?;
                self.last_ping = now;
            }

            if now.duration_since(self.last_pong) > Duration::from_secs(PONG_TIMEOUT_SECS) {
                let close_frame = Frame::close(1000, Some("Pong timeout"));
                let _ = self.send_frame(&close_frame);
                break;
            }

            match self.stream.read(&mut buffer) {
                Ok(0) => break,
                Ok(n) => {
                    self.parser.consume(&buffer[..n]);
                    
                    loop {
                        match self.parser.parse(MAX_PAYLOAD_SIZE) {
                            ParseFrameResult::Complete(frame) => {
                                match self.handle_frame(frame) {
                                    Ok(should_continue) => {
                                        if !should_continue {
                                            return Ok(());
                                        }
                                    }
                                    Err(_) => break,
                                }
                            }
                            ParseFrameResult::Partial => break,
                            ParseFrameResult::TooLarge => {
                                let close_frame = Frame::close(1009, Some("Message too large"));
                                let _ = self.send_frame(&close_frame);
                                break;
                            }
                            ParseFrameResult::Error => {
                                let close_frame = Frame::close(1002, Some("Protocol error"));
                                let _ = self.send_frame(&close_frame);
                                break;
                            }
                        }
                    }
                }
                Err(ref e) if e.kind() == io::ErrorKind::WouldBlock || e.kind() == io::ErrorKind::TimedOut => {}
                Err(_) => break,
            }

            while let Ok(msg) = self.broadcast_rx.try_recv() {
                let frame = Frame::binary(msg);
                let _ = self.send_frame(&frame);
            }

            std::thread::sleep(Duration::from_millis(10));
        }

        Ok(())
    }

    fn handle_frame(&mut self, frame: Frame) -> io::Result<bool> {
        match frame.opcode {
            OpCode::Text | OpCode::Binary => {
                if !frame.fin {
                    let close_frame = Frame::close(1003, Some("Fragmented frames not supported"));
                    let _ = self.send_frame(&close_frame);
                    return Ok(false);
                }
            }
            OpCode::Close => {
                let close_frame = Frame::close(frame.close_code().unwrap_or(1000), None);
                let _ = self.send_frame(&close_frame);
                return Ok(false);
            }
            OpCode::Ping => {
                let pong_frame = Frame::pong(frame.payload);
                self.send_frame(&pong_frame)?;
            }
            OpCode::Pong => {
                self.last_pong = Instant::now();
            }
            OpCode::Continuation => {
                let close_frame = Frame::close(1003, Some("Continuation frames not supported"));
                let _ = self.send_frame(&close_frame);
                return Ok(false);
            }
        }
        Ok(true)
    }

    fn send_frame(&mut self, frame: &Frame) -> io::Result<()> {
        let bytes = encode_frame(frame);
        self.stream.write_all(&bytes)?;
        self.stream.flush()
    }
}

fn log_request(method: Method, path: &str, status: StatusCode, duration: Duration) {
    let datetime = format_system_time(SystemTime::now());
    let duration_ms = duration.as_secs_f64() * 1000.0;
    println!("[{}] {} {} {} {:.2}ms", datetime, method, path, status, duration_ms);
}

fn format_system_time(time: SystemTime) -> String {
    use std::time::UNIX_EPOCH;
    
    let duration = match time.duration_since(UNIX_EPOCH) {
        Ok(d) => d,
        Err(_) => return String::from("1970-01-01 00:00:00.000"),
    };
    
    let secs = duration.as_secs();
    let micros = duration.subsec_micros();
    
    let years = secs / 31536000;
    let leap_years = (years + 1968) / 4 - (years + 1968) / 100 + (years + 1968) / 400 - 477;
    let mut days = (secs % 31536000) / 86400;
    
    if (years + 1970) % 4 == 0 && ((years + 1970) % 100 != 0 || (years + 1970) % 400 == 0) {
        if days > 31 + 28 {
            days -= 1;
        } else if days == 31 + 28 {
        }
    }
    
    let mut month = 1;
    let mut day_in_month = days + 1;
    
    let month_days = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
    for &m_days in &month_days {
        if day_in_month <= m_days {
            break;
        }
        day_in_month -= m_days;
        month += 1;
    }
    
    let hours = (secs % 86400) / 3600;
    let minutes = (secs % 3600) / 60;
    let seconds = secs % 60;
    let millis = micros / 1000;
    
    format!(
        "{:04}-{:02}-{:02} {:02}:{:02}:{:02}.{:03}",
        years + 1970,
        month,
        day_in_month,
        hours,
        minutes,
        seconds,
        millis
    )
}
