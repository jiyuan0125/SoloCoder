use std::env;
use std::time::Duration;

use axum::{
    http::StatusCode,
    response::IntoResponse,
    routing::{post, Router},
    Json,
};
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tokio::time::timeout;

const MAX_NESTING_DEPTH: usize = 10;
const HEADER_SIZE: usize = 3;
const TIMEOUT_SECS: u64 = 1;

#[derive(Debug, Error)]
enum TlvError {
    #[error("Tag {0} 的长度声明为 {1} 但实际数据为 {2}")]
    LengthMismatch(u8, usize, usize),
    #[error("嵌套深度超限")]
    NestingDepthExceeded,
    #[error("无效的十六进制数据")]
    InvalidHex,
    #[error("数据截断，偏移量 {0}")]
    DataTruncated(usize),
    #[error("无效的 JSON 数据")]
    InvalidJson,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct TlvNode {
    tag: u8,
    #[serde(skip_serializing_if = "Option::is_none")]
    value: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    children: Option<Vec<TlvNode>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    offset: Option<usize>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

impl IntoResponse for TlvError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self {
            TlvError::NestingDepthExceeded => StatusCode::BAD_REQUEST,
            TlvError::LengthMismatch(_, _, _) => StatusCode::BAD_REQUEST,
            TlvError::DataTruncated(_) => StatusCode::BAD_REQUEST,
            TlvError::InvalidHex => StatusCode::BAD_REQUEST,
            TlvError::InvalidJson => StatusCode::BAD_REQUEST,
        };
        (status, Json(ErrorResponse { error: self.to_string() })).into_response()
    }
}

fn try_parse_as_tlv_sequence(data: &[u8]) -> bool {
    let mut offset = 0;
    let len = data.len();
    
    if len == 0 {
        return false;
    }
    
    while offset < len {
        if len - offset < HEADER_SIZE {
            return false;
        }
        
        let declared_length = u16::from_be_bytes([data[offset + 1], data[offset + 2]]) as usize;
        let total_size = HEADER_SIZE + declared_length;
        
        if offset + total_size > len {
            return false;
        }
        
        offset += total_size;
    }
    
    offset == len
}

fn parse_tlv(data: &[u8], offset: usize, depth: usize) -> Result<(TlvNode, usize), TlvError> {
    if depth > MAX_NESTING_DEPTH {
        return Err(TlvError::NestingDepthExceeded);
    }
    
    if data.len() < HEADER_SIZE {
        return Err(TlvError::DataTruncated(offset));
    }
    
    let tag = data[0];
    let declared_length = u16::from_be_bytes([data[1], data[2]]) as usize;
    
    if data.len() < HEADER_SIZE + declared_length {
        return Err(TlvError::LengthMismatch(
            tag,
            declared_length,
            data.len() - HEADER_SIZE,
        ));
    }
    
    let value_bytes = &data[HEADER_SIZE..HEADER_SIZE + declared_length];
    
    let node = if !value_bytes.is_empty() && try_parse_as_tlv_sequence(value_bytes) {
        let mut children = Vec::new();
        let mut sub_offset = 0;
        
        while sub_offset < declared_length {
            let (child, consumed) = parse_tlv(
                &value_bytes[sub_offset..],
                offset + HEADER_SIZE + sub_offset,
                depth + 1,
            )?;
            children.push(child);
            sub_offset += consumed;
        }
        
        TlvNode {
            tag,
            value: None,
            children: Some(children),
            offset: Some(offset),
        }
    } else {
        TlvNode {
            tag,
            value: Some(hex::encode(value_bytes)),
            children: None,
            offset: Some(offset),
        }
    };
    
    Ok((node, HEADER_SIZE + declared_length))
}

fn serialize_tlv(node: &TlvNode) -> Result<Vec<u8>, TlvError> {
    let mut result = Vec::new();
    result.push(node.tag);
    
    let mut value_data = Vec::new();
    
    if let Some(children) = &node.children {
        for child in children {
            value_data.extend(serialize_tlv(child)?);
        }
    } else if let Some(value) = &node.value {
        value_data = hex::decode(value).map_err(|_| TlvError::InvalidHex)?;
    }
    
    let length = value_data.len() as u16;
    result.extend_from_slice(&length.to_be_bytes());
    result.extend(value_data);
    
    Ok(result)
}

async fn parse_handler(body: String) -> Result<impl IntoResponse, TlvError> {
    let data = hex::decode(body.trim()).map_err(|_| TlvError::InvalidHex)?;
    
    let parse_operation = async move {
        let mut nodes = Vec::new();
        let mut offset = 0;
        let data_ref = &data[..];
        
        while offset < data.len() {
            let (node, consumed) = parse_tlv(&data_ref[offset..], offset, 1)?;
            nodes.push(node);
            offset += consumed;
        }
        
        Ok::<Json<serde_json::Value>, TlvError>(Json(serde_json::json!({ "nodes": nodes })))
    };
    
    match timeout(Duration::from_secs(TIMEOUT_SECS), parse_operation).await {
        Ok(result) => match result {
            Ok(json) => Ok(json.into_response()),
            Err(e) => Ok(e.into_response()),
        },
        Err(_) => {
            let response = (
                StatusCode::REQUEST_TIMEOUT,
                Json(ErrorResponse {
                    error: "请求超时".to_string(),
                }),
            );
            Ok(response.into_response())
        }
    }
}

async fn serialize_handler(Json(nodes): Json<serde_json::Value>) -> Result<impl IntoResponse, TlvError> {
    let input_nodes: Vec<TlvNode> = nodes
        .get("nodes")
        .ok_or(TlvError::InvalidJson)
        .and_then(|v| serde_json::from_value(v.clone()).map_err(|_| TlvError::InvalidJson))?;
    
    let serialize_operation = async move {
        let mut result = Vec::new();
        for node in &input_nodes {
            result.extend(serialize_tlv(node)?);
        }
        Ok::<String, TlvError>(hex::encode(result))
    };
    
    match timeout(Duration::from_secs(TIMEOUT_SECS), serialize_operation).await {
        Ok(result) => match result {
            Ok(hex_string) => Ok((StatusCode::OK, hex_string).into_response()),
            Err(e) => Ok(e.into_response()),
        },
        Err(_) => {
            let response = (
                StatusCode::REQUEST_TIMEOUT,
                Json(ErrorResponse {
                    error: "请求超时".to_string(),
                }),
            );
            Ok(response.into_response())
        }
    }
}

#[tokio::main]
async fn main() {
    let port = env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let addr = format!("0.0.0.0:{}", port);
    
    let app = Router::new()
        .route("/parse", post(parse_handler))
        .route("/serialize", post(serialize_handler));
    
    let addr: std::net::SocketAddr = addr.parse().unwrap();
    println!("TLV Parser Service listening on {}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
