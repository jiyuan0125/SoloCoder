use axum::{
    extract::{Path, State},
    http::{HeaderMap, HeaderValue, StatusCode},
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use bytes::Bytes;
use prost::encoding::WireType;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tonic::Code;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ContentType {
    #[serde(rename = "json")]
    Json,
    #[serde(rename = "text")]
    Text,
}

impl Default for ContentType {
    fn default() -> Self {
        ContentType::Json
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FieldMapping {
    pub name: String,
    pub field_type: FieldType,
    pub number: u32,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FieldType {
    String,
    Int32,
    Int64,
    Float,
    Double,
    Bool,
    Bytes,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageSchema {
    pub fields: Vec<FieldMapping>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RouteConfig {
    pub service: String,
    pub method: String,
    pub grpc_endpoint: String,
    #[serde(default)]
    pub content_type: ContentType,
    pub request_schema: MessageSchema,
    pub response_schema: MessageSchema,
    #[serde(default)]
    pub text_separator: Option<String>,
    #[serde(default)]
    pub text_field_order: Option<Vec<String>>,
}

#[derive(Debug, Clone)]
pub struct CompiledRoute {
    pub config: RouteConfig,
}

pub type RouteStore = Arc<RwLock<HashMap<String, CompiledRoute>>>;

fn route_key(service: &str, method: &str) -> String {
    format!("{}/{}", service, method)
}

fn grpc_code_to_http_status(code: Code) -> StatusCode {
    match code {
        Code::Ok => StatusCode::OK,
        Code::InvalidArgument => StatusCode::BAD_REQUEST,
        Code::NotFound => StatusCode::NOT_FOUND,
        Code::Unavailable => StatusCode::SERVICE_UNAVAILABLE,
        Code::DeadlineExceeded => StatusCode::GATEWAY_TIMEOUT,
        Code::Internal => StatusCode::INTERNAL_SERVER_ERROR,
        _ => StatusCode::INTERNAL_SERVER_ERROR,
    }
}

fn encode_varint(mut value: u64, buf: &mut Vec<u8>) {
    loop {
        let mut b = (value & 0x7F) as u8;
        value >>= 7;
        if value != 0 {
            b |= 0x80;
        }
        buf.push(b);
        if value == 0 {
            break;
        }
    }
}

fn decode_varint(bytes: &[u8]) -> (u64, usize) {
    let mut result: u64 = 0;
    let mut shift = 0;
    let mut consumed = 0;
    
    for &b in bytes {
        consumed += 1;
        result |= ((b & 0x7F) as u64) << shift;
        shift += 7;
        if (b & 0x80) == 0 {
            break;
        }
    }
    
    (result, consumed)
}

fn encode_zigzag_32(n: i32) -> u32 {
    ((n << 1) ^ (n >> 31)) as u32
}

fn encode_zigzag_64(n: i64) -> u64 {
    ((n << 1) ^ (n >> 63)) as u64
}

fn decode_zigzag_32(n: u32) -> i32 {
    ((n >> 1) as i32) ^ (-((n & 1) as i32))
}

fn decode_zigzag_64(n: u64) -> i64 {
    ((n >> 1) as i64) ^ (-((n & 1) as i64))
}

fn encode_tag(field_number: u32, wire_type: WireType, buf: &mut Vec<u8>) {
    let tag = (field_number << 3) | (wire_type as u32);
    encode_varint(tag as u64, buf);
}

fn decode_tag(tag: u64) -> (u32, WireType) {
    let field_number = (tag >> 3) as u32;
    let wire_type_value = (tag & 0x07) as u8;
    let wire_type = match wire_type_value {
        0 => WireType::Varint,
        1 => WireType::SixtyFourBit,
        2 => WireType::LengthDelimited,
        5 => WireType::ThirtyTwoBit,
        _ => WireType::LengthDelimited,
    };
    (field_number, wire_type)
}

fn wire_type_for_field(field_type: FieldType) -> WireType {
    match field_type {
        FieldType::String => WireType::LengthDelimited,
        FieldType::Int32 => WireType::Varint,
        FieldType::Int64 => WireType::Varint,
        FieldType::Float => WireType::ThirtyTwoBit,
        FieldType::Double => WireType::SixtyFourBit,
        FieldType::Bool => WireType::Varint,
        FieldType::Bytes => WireType::LengthDelimited,
    }
}

fn encode_proto_value(field: &FieldMapping, value: &serde_json::Value) -> anyhow::Result<Vec<u8>> {
    let mut buf = Vec::new();
    let wire_type = wire_type_for_field(field.field_type);
    
    encode_tag(field.number, wire_type, &mut buf);
    
    match field.field_type {
        FieldType::String => {
            let s = value.as_str().ok_or_else(|| anyhow::anyhow!("Expected string"))?;
            encode_varint(s.len() as u64, &mut buf);
            buf.extend_from_slice(s.as_bytes());
        }
        FieldType::Int32 => {
            let n = value.as_i64().ok_or_else(|| anyhow::anyhow!("Expected integer"))? as i32;
            encode_varint(encode_zigzag_32(n) as u64, &mut buf);
        }
        FieldType::Int64 => {
            let n = value.as_i64().ok_or_else(|| anyhow::anyhow!("Expected integer"))?;
            encode_varint(encode_zigzag_64(n), &mut buf);
        }
        FieldType::Float => {
            let f = value.as_f64().ok_or_else(|| anyhow::anyhow!("Expected float"))? as f32;
            buf.extend_from_slice(&f.to_le_bytes());
        }
        FieldType::Double => {
            let d = value.as_f64().ok_or_else(|| anyhow::anyhow!("Expected double"))?;
            buf.extend_from_slice(&d.to_le_bytes());
        }
        FieldType::Bool => {
            let b = value.as_bool().ok_or_else(|| anyhow::anyhow!("Expected bool"))?;
            encode_varint(b as u64, &mut buf);
        }
        FieldType::Bytes => {
            let s = value.as_str().ok_or_else(|| anyhow::anyhow!("Expected string for bytes"))?;
            let bytes = base64::Engine::decode(&base64::engine::general_purpose::STANDARD, s)?;
            encode_varint(bytes.len() as u64, &mut buf);
            buf.extend_from_slice(&bytes);
        }
    }
    
    Ok(buf)
}

fn json_to_proto(schema: &MessageSchema, json: &serde_json::Value) -> anyhow::Result<Vec<u8>> {
    let obj = json.as_object().ok_or_else(|| anyhow::anyhow!("Expected JSON object"))?;
    let mut buf = Vec::new();
    
    for field in &schema.fields {
        if let Some(value) = obj.get(&field.name) {
            let encoded = encode_proto_value(field, value)?;
            buf.extend(encoded);
        }
    }
    
    Ok(buf)
}

fn decode_proto_value(field: &FieldMapping, bytes: &[u8]) -> anyhow::Result<(serde_json::Value, usize)> {
    let mut pos = 0usize;
    
    let value = match field.field_type {
        FieldType::String => {
            let (len, consumed) = decode_varint(&bytes[pos..]);
            pos += consumed;
            let len = len as usize;
            if len > bytes.len() - pos {
                return Err(anyhow::anyhow!("Invalid length-delimited field"));
            }
            let s = std::str::from_utf8(&bytes[pos..pos + len])?.to_string();
            pos += len;
            serde_json::Value::String(s)
        }
        FieldType::Int32 => {
            let (v, consumed) = decode_varint(&bytes[pos..]);
            pos += consumed;
            let n = decode_zigzag_32(v as u32);
            serde_json::Value::Number(serde_json::Number::from(n))
        }
        FieldType::Int64 => {
            let (v, consumed) = decode_varint(&bytes[pos..]);
            pos += consumed;
            let n = decode_zigzag_64(v);
            serde_json::Value::Number(serde_json::Number::from(n))
        }
        FieldType::Float => {
            if bytes.len() - pos < 4 {
                return Err(anyhow::anyhow!("Insufficient bytes for float"));
            }
            let f = f32::from_le_bytes(bytes[pos..pos + 4].try_into()?);
            pos += 4;
            serde_json::Value::Number(serde_json::Number::from_f64(f as f64)
                .ok_or_else(|| anyhow::anyhow!("Invalid float"))?)
        }
        FieldType::Double => {
            if bytes.len() - pos < 8 {
                return Err(anyhow::anyhow!("Insufficient bytes for double"));
            }
            let d = f64::from_le_bytes(bytes[pos..pos + 8].try_into()?);
            pos += 8;
            serde_json::Value::Number(serde_json::Number::from_f64(d)
                .ok_or_else(|| anyhow::anyhow!("Invalid double"))?)
        }
        FieldType::Bool => {
            let (v, consumed) = decode_varint(&bytes[pos..]);
            pos += consumed;
            serde_json::Value::Bool(v != 0)
        }
        FieldType::Bytes => {
            let (len, consumed) = decode_varint(&bytes[pos..]);
            pos += consumed;
            let len = len as usize;
            if len > bytes.len() - pos {
                return Err(anyhow::anyhow!("Invalid length-delimited field"));
            }
            let encoded = base64::Engine::encode(&base64::engine::general_purpose::STANDARD, &bytes[pos..pos + len]);
            pos += len;
            serde_json::Value::String(encoded)
        }
    };
    
    Ok((value, pos))
}

fn proto_to_json(schema: &MessageSchema, bytes: &[u8]) -> anyhow::Result<serde_json::Value> {
    let mut map = serde_json::Map::new();
    let mut field_map: HashMap<u32, &FieldMapping> = HashMap::new();
    
    for field in &schema.fields {
        field_map.insert(field.number, field);
    }
    
    let mut pos = 0;
    
    while pos < bytes.len() {
        let (tag, consumed) = decode_varint(&bytes[pos..]);
        pos += consumed;
        
        let (field_number, _wire_type) = decode_tag(tag);
        
        if let Some(field) = field_map.get(&field_number) {
            let (value, consumed) = decode_proto_value(field, &bytes[pos..])?;
            map.insert(field.name.clone(), value);
            pos += consumed;
        } else {
            pos = bytes.len();
        }
    }
    
    Ok(serde_json::Value::Object(map))
}

async fn create_route(
    State(routes): State<RouteStore>,
    Json(config): Json<RouteConfig>,
) -> Result<StatusCode, AppError> {
    let compiled = CompiledRoute {
        config: config.clone(),
    };
    
    let key = route_key(&config.service, &config.method);
    routes.write().await.insert(key, compiled);
    
    Ok(StatusCode::CREATED)
}

async fn list_routes(State(routes): State<RouteStore>) -> Json<Vec<RouteConfig>> {
    let routes = routes.read().await;
    let configs: Vec<RouteConfig> = routes.values().map(|r| r.config.clone()).collect();
    Json(configs)
}

async fn handle_request(
    Path((service, method)): Path<(String, String)>,
    headers: HeaderMap,
    State(routes): State<RouteStore>,
    body: String,
) -> Result<Response, AppError> {
    let key = route_key(&service, &method);
    let routes = routes.read().await;
    let compiled = routes.get(&key).ok_or(AppError::RouteNotFound)?;
    let compiled = compiled.clone();
    drop(routes);
    
    let request_msg = parse_request_body(&body, &compiled, &headers)?;
    
    let response_bytes = call_grpc_service(&compiled, request_msg).await?;
    
    let response_msg = proto_to_json(&compiled.config.response_schema, &response_bytes)?;
    
    let json_response = serde_json::to_string_pretty(&response_msg)?;
    
    Ok((
        StatusCode::OK,
        [(axum::http::header::CONTENT_TYPE, HeaderValue::from_static("application/json"))],
        json_response,
    ).into_response())
}

fn parse_request_body(
    body: &str,
    compiled: &CompiledRoute,
    _headers: &HeaderMap,
) -> Result<Vec<u8>, AppError> {
    match compiled.config.content_type {
        ContentType::Json => {
            let json_value: serde_json::Value = serde_json::from_str(body)?;
            let proto_bytes = json_to_proto(&compiled.config.request_schema, &json_value)?;
            Ok(proto_bytes)
        }
        ContentType::Text => parse_text_body(body, compiled),
    }
}

fn parse_text_body(
    body: &str,
    compiled: &CompiledRoute,
) -> Result<Vec<u8>, AppError> {
    let separator = compiled
        .config
        .text_separator
        .as_deref()
        .unwrap_or(",");
    
    let field_order = compiled
        .config
        .text_field_order
        .as_ref()
        .ok_or_else(|| AppError::MissingTextFieldOrder)?;
    
    let parts: Vec<&str> = body.split(separator).collect();
    
    if parts.len() != field_order.len() {
        return Err(AppError::TextFieldCountMismatch {
            expected: field_order.len(),
            got: parts.len(),
        });
    }
    
    let mut json_map = serde_json::Map::new();
    
    for (i, field_name) in field_order.iter().enumerate() {
        json_map.insert(field_name.clone(), serde_json::Value::String(parts[i].to_string()));
    }
    
    let json_value = serde_json::Value::Object(json_map);
    let proto_bytes = json_to_proto(&compiled.config.request_schema, &json_value)?;
    
    Ok(proto_bytes)
}

async fn call_grpc_service(
    compiled: &CompiledRoute,
    request_bytes: Vec<u8>,
) -> Result<Vec<u8>, AppError> {
    use http_body_util::{BodyExt, Full};
    use hyper::Uri;
    use hyper_util::rt::TokioIo;
    use tokio::net::TcpStream;
    
    let endpoint = &compiled.config.grpc_endpoint;
    
    let uri: Uri = endpoint.parse()?;
    let host = uri.host().ok_or_else(|| AppError::Internal("Missing host".to_string()))?;
    let port = uri.port_u16().unwrap_or(80);
    
    let stream = TcpStream::connect((host, port)).await
        .map_err(|_| AppError::GrpcUnavailable)?;
    
    let io = TokioIo::new(stream);
    
    let (mut sender, conn) = hyper::client::conn::http2::handshake(hyper_util::rt::TokioExecutor::new(), io)
        .await
        .map_err(|e| AppError::GrpcConnectError(e.to_string()))?;
    
    tokio::spawn(async move {
        let _ = conn.await;
    });
    
    let mut grpc_frame = Vec::new();
    grpc_frame.push(0u8);
    grpc_frame.extend_from_slice(&(request_bytes.len() as u32).to_be_bytes());
    grpc_frame.extend_from_slice(&request_bytes);
    
    let path = format!("/{}/{}", compiled.config.service, compiled.config.method);
    let authority = format!("{}:{}", host, port);
    
    let req = hyper::Request::builder()
        .method(hyper::Method::POST)
        .uri(format!("http://{}{}", authority, path))
        .header("content-type", "application/grpc")
        .header("te", "trailers")
        .header(":authority", &authority)
        .header(":path", &path)
        .header(":scheme", "http")
        .header(":method", "POST")
        .body(Full::new(Bytes::from(grpc_frame)))
        .map_err(|e| AppError::Internal(e.to_string()))?;
    
    let response = sender.send_request(req).await
        .map_err(|e| AppError::GrpcConnectError(e.to_string()))?;
    
    let status = response.status();
    if !status.is_success() {
        return Err(AppError::GrpcStatus(status.as_u16()));
    }
    
    let body_bytes = response.into_body().collect().await
        .map_err(|e| AppError::GrpcConnectError(e.to_string()))?
        .to_bytes();
    
    if body_bytes.len() < 5 {
        return Err(AppError::Internal("Invalid gRPC response".to_string()));
    }
    
    let _compressed = body_bytes[0] != 0;
    let length = u32::from_be_bytes(body_bytes[1..5].try_into().unwrap()) as usize;
    
    if body_bytes.len() < 5 + length {
        return Err(AppError::Internal("Incomplete gRPC response".to_string()));
    }
    
    let message_bytes = body_bytes[5..5 + length].to_vec();
    
    Ok(message_bytes)
}

#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("Route not found")]
    RouteNotFound,
    #[error("Missing text field order configuration")]
    MissingTextFieldOrder,
    #[error("Text field count mismatch: expected {expected}, got {got}")]
    TextFieldCountMismatch { expected: usize, got: usize },
    #[error("gRPC service unavailable")]
    GrpcUnavailable,
    #[error("gRPC connect error: {0}")]
    GrpcConnectError(String),
    #[error("gRPC HTTP status: {0}")]
    GrpcStatus(u16),
    #[error("Invalid URI: {0}")]
    InvalidUri(#[from] http::uri::InvalidUri),
    #[error("Serialization error: {0}")]
    SerdeError(#[from] serde_json::Error),
    #[error("Internal error: {0}")]
    Internal(String),
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
    #[error("Anyhow error: {0}")]
    AnyhowError(#[from] anyhow::Error),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::RouteNotFound => (StatusCode::NOT_FOUND, self.to_string()),
            AppError::GrpcUnavailable => (StatusCode::BAD_GATEWAY, self.to_string()),
            _ => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
        };
        
        (status, Json(serde_json::json!({ "error": message }))).into_response()
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let routes: RouteStore = Arc::new(RwLock::new(HashMap::new()));
    
    let app = Router::new()
        .route("/routes", post(create_route))
        .route("/routes", get(list_routes))
        .route("/v1/:service/:method", post(handle_request))
        .with_state(routes);
    
    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let addr: std::net::SocketAddr = format!("0.0.0.0:{}", port).parse()?;
    
    println!("Server listening on {}", addr);
    axum::serve(tokio::net::TcpListener::bind(&addr).await?, app).await?;
    
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn test_varint_encoding() {
        let mut buf = Vec::new();
        encode_varint(150, &mut buf);
        assert_eq!(buf, vec![0x96, 0x01]);
    }

    #[test]
    fn test_varint_decoding() {
        let (value, consumed) = decode_varint(&[0x96, 0x01]);
        assert_eq!(value, 150);
        assert_eq!(consumed, 2);
    }

    #[test]
    fn test_zigzag_32() {
        assert_eq!(encode_zigzag_32(0), 0);
        assert_eq!(encode_zigzag_32(-1), 1);
        assert_eq!(encode_zigzag_32(1), 2);
        assert_eq!(encode_zigzag_32(-2), 3);
        assert_eq!(decode_zigzag_32(0), 0);
        assert_eq!(decode_zigzag_32(1), -1);
        assert_eq!(decode_zigzag_32(2), 1);
    }

    #[test]
    fn test_zigzag_64() {
        assert_eq!(encode_zigzag_64(0), 0);
        assert_eq!(encode_zigzag_64(-1), 1);
        assert_eq!(decode_zigzag_64(1), -1);
    }

    #[test]
    fn test_json_to_proto_string() {
        let schema = MessageSchema {
            fields: vec![FieldMapping {
                name: "name".to_string(),
                field_type: FieldType::String,
                number: 1,
            }],
        };
        
        let json = json!({ "name": "test" });
        let proto = json_to_proto(&schema, &json).unwrap();
        
        // Tag: field 1, wire type 2 (length-delimited) = 0x0A
        // Length: 4
        // Data: "test"
        assert_eq!(proto, vec![0x0A, 0x04, b't', b'e', b's', b't']);
    }

    #[test]
    fn test_proto_to_json_string() {
        let schema = MessageSchema {
            fields: vec![FieldMapping {
                name: "name".to_string(),
                field_type: FieldType::String,
                number: 1,
            }],
        };
        
        let proto = vec![0x0A, 0x04, b't', b'e', b's', b't'];
        let json = proto_to_json(&schema, &proto).unwrap();
        
        assert_eq!(json.get("name").unwrap().as_str().unwrap(), "test");
    }

    #[test]
    fn test_json_to_proto_int32() {
        let schema = MessageSchema {
            fields: vec![FieldMapping {
                name: "id".to_string(),
                field_type: FieldType::Int32,
                number: 2,
            }],
        };
        
        let json = json!({ "id": 150 });
        let proto = json_to_proto(&schema, &json).unwrap();
        
        // Tag: field 2, wire type 0 = 0x10
        // Zig-zag encoded 150 = 300 = 0xAC 0x02
        assert_eq!(proto, vec![0x10, 0xAC, 0x02]);
    }

    #[test]
    fn test_grpc_code_mapping() {
        assert_eq!(grpc_code_to_http_status(Code::Ok), StatusCode::OK);
        assert_eq!(grpc_code_to_http_status(Code::InvalidArgument), StatusCode::BAD_REQUEST);
        assert_eq!(grpc_code_to_http_status(Code::NotFound), StatusCode::NOT_FOUND);
        assert_eq!(grpc_code_to_http_status(Code::Unavailable), StatusCode::SERVICE_UNAVAILABLE);
        assert_eq!(grpc_code_to_http_status(Code::DeadlineExceeded), StatusCode::GATEWAY_TIMEOUT);
        assert_eq!(grpc_code_to_http_status(Code::Internal), StatusCode::INTERNAL_SERVER_ERROR);
    }

    #[test]
    fn test_route_key() {
        assert_eq!(route_key("Greeter", "SayHello"), "Greeter/SayHello");
    }
}
