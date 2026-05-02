use crate::http::{self, Method, Response, ParseError, MAX_BODY_SIZE};
use crate::websocket::{self, Frame, OpCode, FrameParseError, PING_INTERVAL_SECS, PONG_TIMEOUT_SECS};
use crate::router::Router;
use std::collections::HashMap;
use std::io::{self, Read, Write};
use std::net::TcpStream;
use std::sync::{Arc, Mutex};
use std::time::{Instant, Duration};

pub type WebSocketClients = Arc<Mutex<HashMap<usize, Arc<Mutex<TcpStream>>>>>;

pub struct ConnectionContext {
    pub stream: TcpStream,
    pub router: Arc<Router>,
    pub ws_clients: WebSocketClients,
    pub client_id: usize,
    pub shutdown_signal: Arc<Mutex<bool>>,
}

pub enum ConnectionResult {
    Close,
    UpgradedToWebSocket,
}

pub fn handle_http_connection(ctx: &mut ConnectionContext) -> io::Result<ConnectionResult> {
    let mut buffer = Vec::with_capacity(8192);
    let mut read_buf = [0u8; 4096];
    
    ctx.stream.set_read_timeout(Some(Duration::from_secs(30)))?;
    
    loop {
        if *ctx.shutdown_signal.lock().unwrap() {
            return Ok(ConnectionResult::Close);
        }
        
        match ctx.stream.read(&mut read_buf) {
            Ok(0) => return Ok(ConnectionResult::Close),
            Ok(n) => {
                buffer.extend_from_slice(&read_buf[..n]);
            }
            Err(ref e) if e.kind() == io::ErrorKind::WouldBlock || e.kind() == io::ErrorKind::TimedOut => {
                if buffer.is_empty() {
                    continue;
                }
            }
            Err(e) => return Err(e),
        }
        
        let start_time = Instant::now();
        
        let (mut request, header_end) = match http::parse_request_line_and_headers(&buffer) {
            Ok((req, end)) => (req, end),
            Err(ParseError::Incomplete) => continue,
            Err(ParseError::HeaderTooLarge) => {
                let response = Response::with_text(431, "Request Header Fields Too Large");
                ctx.stream.write_all(&response.to_bytes())?;
                println!("[{}] UNKNOWN / 431 {}ms", 
                    chrono::Local::now().format("%Y-%m-%d %H:%M:%S"),
                    start_time.elapsed().as_millis());
                return Ok(ConnectionResult::Close);
            }
            Err(_) => {
                let response = Response::with_text(400, "Bad Request");
                ctx.stream.write_all(&response.to_bytes())?;
                return Ok(ConnectionResult::Close);
            }
        };
        
        if request.method == Method::Post {
            let content_length = request.content_length().unwrap_or(0);
            
            if content_length > MAX_BODY_SIZE {
                let response = Response::with_text(413, "Payload Too Large");
                ctx.stream.write_all(&response.to_bytes())?;
                log_request(&request.method, &request.raw_path, 413, start_time.elapsed());
                return Ok(ConnectionResult::Close);
            }
            
            let mut body_read = buffer.len().saturating_sub(header_end);
            
            while body_read < content_length {
                match ctx.stream.read(&mut read_buf) {
                    Ok(0) => return Ok(ConnectionResult::Close),
                    Ok(n) => {
                        buffer.extend_from_slice(&read_buf[..n]);
                        body_read += n;
                    }
                    Err(e) => return Err(e),
                }
            }
            
            request.body = buffer[header_end..header_end + content_length].to_vec();
            buffer.drain(..header_end + content_length);
        } else {
            buffer.drain(..header_end);
        }
        
        if request.is_websocket_upgrade() {
            let sec_key = request.header("sec-websocket-key").unwrap();
            let handshake = websocket::create_handshake_response(sec_key);
            ctx.stream.write_all(&handshake)?;
            ctx.stream.flush()?;
            log_request(&request.method, &request.raw_path, 101, start_time.elapsed());
            return Ok(ConnectionResult::UpgradedToWebSocket);
        }
        
        let response = ctx.router.handle(&request);
        let status_code = response.status_code;
        ctx.stream.write_all(&response.to_bytes())?;
        ctx.stream.flush()?;
        
        log_request(&request.method, &request.raw_path, status_code, start_time.elapsed());
        
        if !request.keep_alive() {
            return Ok(ConnectionResult::Close);
        }
    }
}

pub fn handle_websocket_connection(ctx: &mut ConnectionContext) -> io::Result<()> {
    let stream_clone = ctx.stream.try_clone()?;
    let client_id = ctx.client_id;
    
    {
        let mut clients = ctx.ws_clients.lock().unwrap();
        clients.insert(client_id, Arc::new(Mutex::new(stream_clone)));
    }
    
    let result = handle_websocket_frames(ctx);
    
    {
        let mut clients = ctx.ws_clients.lock().unwrap();
        clients.remove(&client_id);
    }
    
    result
}

fn handle_websocket_frames(ctx: &mut ConnectionContext) -> io::Result<()> {
    let mut buffer = Vec::with_capacity(8192);
    let mut read_buf = [0u8; 4096];
    
    let mut last_pong = Instant::now();
    let mut last_ping = Instant::now();
    
    ctx.stream.set_read_timeout(Some(Duration::from_secs(1)))?;
    
    loop {
        if *ctx.shutdown_signal.lock().unwrap() {
            let close_frame = Frame::close(1001, "Server shutting down");
            let _ = ctx.stream.write_all(&close_frame.encode());
            return Ok(());
        }
        
        let now = Instant::now();
        
        if now.duration_since(last_ping).as_secs() >= PING_INTERVAL_SECS {
            let ping_frame = Frame::ping(b"ping".to_vec());
            ctx.stream.write_all(&ping_frame.encode())?;
            ctx.stream.flush()?;
            last_ping = now;
        }
        
        if now.duration_since(last_pong).as_secs() >= PONG_TIMEOUT_SECS {
            return Ok(());
        }
        
        match ctx.stream.read(&mut read_buf) {
            Ok(0) => return Ok(()),
            Ok(n) => {
                buffer.extend_from_slice(&read_buf[..n]);
            }
            Err(ref e) if e.kind() == io::ErrorKind::WouldBlock || e.kind() == io::ErrorKind::TimedOut => {
                continue;
            }
            Err(e) => return Err(e),
        }
        
        loop {
            if buffer.is_empty() {
                break;
            }
            
            match websocket::parse_frame(&buffer, true) {
                Ok((frame, consumed)) => {
                    buffer.drain(..consumed);
                    
                    match frame.opcode {
                        OpCode::Text | OpCode::Binary => {
                            if !frame.fin {
                                let close_frame = Frame::close(1003, "Continuation frames not supported");
                                let _ = ctx.stream.write_all(&close_frame.encode());
                                return Ok(());
                            }
                        }
                        OpCode::Close => {
                            let close_frame = Frame::close(frame.close_code().unwrap_or(1000), "");
                            let _ = ctx.stream.write_all(&close_frame.encode());
                            return Ok(());
                        }
                        OpCode::Ping => {
                            let pong_frame = Frame::pong(frame.payload);
                            ctx.stream.write_all(&pong_frame.encode())?;
                            ctx.stream.flush()?;
                        }
                        OpCode::Pong => {
                            last_pong = Instant::now();
                        }
                        OpCode::Continuation => {
                            let close_frame = Frame::close(1003, "Continuation frames not supported");
                            let _ = ctx.stream.write_all(&close_frame.encode());
                            return Ok(());
                        }
                    }
                }
                Err(FrameParseError::Incomplete) => {
                    break;
                }
                Err(FrameParseError::PayloadTooLarge) => {
                    let close_frame = Frame::close(1009, "Message too large");
                    let _ = ctx.stream.write_all(&close_frame.encode());
                    return Ok(());
                }
                Err(_) => {
                    let close_frame = Frame::close(1002, "Protocol error");
                    let _ = ctx.stream.write_all(&close_frame.encode());
                    return Ok(());
                }
            }
        }
    }
}

fn log_request(method: &Method, path: &str, status: u16, duration: Duration) {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs();
    
    let datetime = chrono::DateTime::from_timestamp(now as i64, 0)
        .unwrap()
        .format("%Y-%m-%d %H:%M:%S");
    
    let ms = duration.as_millis();
    println!("[{}] {} {} {} {}ms", datetime, method, path, status, ms);
}
