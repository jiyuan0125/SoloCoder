use prost_types::{ListValue, Struct, Value};
use serde_json::{Map, Number, Value as JsonValue};
use std::collections::BTreeMap;
use thiserror::Error;
use tonic::Code;

#[derive(Debug, Error)]
pub enum GatewayError {
    #[error("Backend unavailable: {0}")]
    BackendUnavailable(String),
    #[error("gRPC status: {0} - {1}")]
    GrpcStatus(String, String),
    #[error("JSON encoding error: {0}")]
    JsonEncoding(String),
    #[error("Protobuf encoding error: {0}")]
    ProtobufEncoding(String),
    #[error("Other error: {0}")]
    Other(String),
}

pub struct GatewayRequest {
    pub backend_address: String,
    pub service: String,
    pub method: String,
    pub json_body: JsonValue,
    pub proto_definition: String,
}

fn json_to_proto_value(json: &JsonValue) -> Value {
    match json {
        JsonValue::Null => Value {
            kind: Some(prost_types::value::Kind::NullValue(0)),
        },
        JsonValue::Bool(b) => Value {
            kind: Some(prost_types::value::Kind::BoolValue(*b)),
        },
        JsonValue::Number(n) => {
            if let Some(i) = n.as_i64() {
                Value {
                    kind: Some(prost_types::value::Kind::NumberValue(i as f64)),
                }
            } else if let Some(f) = n.as_f64() {
                Value {
                    kind: Some(prost_types::value::Kind::NumberValue(f)),
                }
            } else {
                Value {
                    kind: Some(prost_types::value::Kind::StringValue(n.to_string())),
                }
            }
        }
        JsonValue::String(s) => Value {
            kind: Some(prost_types::value::Kind::StringValue(s.clone())),
        },
        JsonValue::Array(arr) => {
            let values: Vec<Value> = arr.iter().map(json_to_proto_value).collect();
            Value {
                kind: Some(prost_types::value::Kind::ListValue(ListValue { values })),
            }
        }
        JsonValue::Object(obj) => {
            let mut fields = BTreeMap::new();
            for (k, v) in obj {
                fields.insert(k.clone(), json_to_proto_value(v));
            }
            Value {
                kind: Some(prost_types::value::Kind::StructValue(Struct { fields })),
            }
        }
    }
}

fn proto_to_json_value(value: &Value) -> JsonValue {
    match &value.kind {
        Some(prost_types::value::Kind::NullValue(_)) => JsonValue::Null,
        Some(prost_types::value::Kind::BoolValue(b)) => JsonValue::Bool(*b),
        Some(prost_types::value::Kind::NumberValue(n)) => {
            if let Some(int) = Number::from_f64(*n) {
                JsonValue::Number(int)
            } else {
                JsonValue::String(n.to_string())
            }
        }
        Some(prost_types::value::Kind::StringValue(s)) => JsonValue::String(s.clone()),
        Some(prost_types::value::Kind::ListValue(list)) => {
            let arr: Vec<JsonValue> = list.values.iter().map(proto_to_json_value).collect();
            JsonValue::Array(arr)
        }
        Some(prost_types::value::Kind::StructValue(s)) => {
            let mut obj = Map::new();
            for (k, v) in &s.fields {
                obj.insert(k.clone(), proto_to_json_value(v));
            }
            JsonValue::Object(obj)
        }
        None => JsonValue::Null,
    }
}

fn encode_struct(s: &Struct) -> Vec<u8> {
    use prost::Message;
    let mut buf = Vec::new();
    s.encode(&mut buf).unwrap();
    buf
}

fn decode_struct(data: &[u8]) -> Result<Struct, String> {
    use prost::Message;
    Struct::decode(data).map_err(|e| e.to_string())
}

pub async fn invoke_grpc_service(req: GatewayRequest) -> Result<JsonValue, GatewayError> {
    let input_struct = match req.json_body {
        JsonValue::Object(obj) => {
            let mut fields = BTreeMap::new();
            for (k, v) in obj {
                fields.insert(k, json_to_proto_value(&v));
            }
            Struct { fields }
        }
        _ => {
            let mut fields = BTreeMap::new();
            fields.insert("value".to_string(), json_to_proto_value(&req.json_body));
            Struct { fields }
        }
    };

    let endpoint = format!("http://{}", req.backend_address);
    let full_method = format!("/{}/{}", req.service, req.method);
    let encoded = encode_struct(&input_struct);

    let response = invoke_raw_grpc(&endpoint, &full_method, &encoded).await;

    match response {
        Ok(response_bytes) => {
            let output_struct = decode_struct(&response_bytes)
                .map_err(|e| GatewayError::ProtobufEncoding(e))?;
            Ok(proto_to_json_value(&Value {
                kind: Some(prost_types::value::Kind::StructValue(output_struct)),
            }))
        }
        Err(e) => match e {
            RawGrpcError::ConnectionError(msg) => Err(GatewayError::BackendUnavailable(msg)),
            RawGrpcError::GrpcStatus(code, msg) => {
                let code_str = match code {
                    Code::Ok => "OK",
                    Code::InvalidArgument => "INVALID_ARGUMENT",
                    Code::NotFound => "NOT_FOUND",
                    Code::Unavailable => "UNAVAILABLE",
                    Code::DeadlineExceeded => "DEADLINE_EXCEEDED",
                    _ => "INTERNAL",
                };
                Err(GatewayError::GrpcStatus(code_str.to_string(), msg))
            }
            RawGrpcError::Other(msg) => Err(GatewayError::Other(msg)),
        },
    }
}

enum RawGrpcError {
    ConnectionError(String),
    GrpcStatus(Code, String),
    Other(String),
}

async fn invoke_raw_grpc(endpoint: &str, full_method: &str, payload: &[u8]) -> Result<Vec<u8>, RawGrpcError> {
    use http::header::{CONTENT_TYPE, TE};
    use http::Request;
    use hyper::client::HttpConnector;
    use hyper::{Body, Client};

    let mut connector = HttpConnector::new();
    connector.enforce_http(false);

    let client = Client::builder()
        .http2_only(true)
        .build::<_, Body>(connector);

    let url = format!("{}{}", endpoint, full_method);

    let request = Request::builder()
        .method("POST")
        .uri(url)
        .header(CONTENT_TYPE, "application/grpc")
        .header(TE, "trailers")
        .body(Body::from(encode_grpc_payload(payload)))
        .map_err(|e| RawGrpcError::Other(e.to_string()))?;

    let resp = client.request(request)
        .await
        .map_err(|e| RawGrpcError::ConnectionError(e.to_string()))?;

    let status = resp.status();
    let (_, body) = resp.into_parts();

    let body_bytes = hyper::body::to_bytes(body)
        .await
        .map_err(|e| RawGrpcError::Other(e.to_string()))?;

    if !status.is_success() {
        let status_code = match status.as_u16() {
            400 => Code::InvalidArgument,
            404 => Code::NotFound,
            503 => Code::Unavailable,
            504 => Code::DeadlineExceeded,
            _ => Code::Internal,
        };
        return Err(RawGrpcError::GrpcStatus(
            status_code,
            format!("HTTP status: {}", status),
        ));
    }

    if body_bytes.len() < 5 {
        return Err(RawGrpcError::Other("Invalid gRPC response".to_string()));
    }

    let compressed = body_bytes[0];
    let length = u32::from_be_bytes([
        body_bytes[1],
        body_bytes[2],
        body_bytes[3],
        body_bytes[4],
    ]) as usize;

    if body_bytes.len() < 5 + length {
        return Err(RawGrpcError::Other("Incomplete gRPC response".to_string()));
    }

    if compressed != 0 {
        return Err(RawGrpcError::Other("Compression not supported".to_string()));
    }

    let message = &body_bytes[5..5 + length];
    Ok(message.to_vec())
}

fn encode_grpc_payload(message: &[u8]) -> Vec<u8> {
    let mut result = Vec::with_capacity(5 + message.len());
    result.push(0);
    result.extend_from_slice(&(message.len() as u32).to_be_bytes());
    result.extend_from_slice(message);
    result
}
