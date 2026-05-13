use std::net::SocketAddr;
use std::process::Stdio;
use std::time::Duration;

use tokio::net::TcpStream;
use tokio::process::Command;
use trust_dns_resolver::TokioAsyncResolver;

use crate::models::{HealthStatus, ProbeParams, ProbeResult};
use uuid::Uuid;

pub async fn execute_probe(probe_id: Uuid, params: &ProbeParams) -> ProbeResult {
    let start = std::time::Instant::now();
    let (status, message) = match params {
        ProbeParams::Http {
            url,
            expected_status,
            expected_body,
            timeout_secs,
        } => execute_http(url, *expected_status, expected_body.as_deref(), *timeout_secs).await,
        ProbeParams::Tcp {
            host,
            port,
            timeout_secs,
        } => execute_tcp(host, *port, *timeout_secs).await,
        ProbeParams::Dns {
            domain,
            timeout_secs,
        } => execute_dns(domain, *timeout_secs).await,
        ProbeParams::Script {
            command,
            timeout_secs,
        } => execute_script(command, *timeout_secs).await,
    };

    let duration_ms = start.elapsed().as_millis() as u64;
    let timestamp = chrono::Utc::now().timestamp();

    ProbeResult {
        probe_id,
        status,
        timestamp,
        message,
        duration_ms,
    }
}

async fn execute_http(
    url: &str,
    expected_status: Option<u16>,
    expected_body: Option<&str>,
    timeout_secs: Option<u64>,
) -> (HealthStatus, Option<String>) {
    let timeout = Duration::from_secs(timeout_secs.unwrap_or(10));
    let client = match reqwest::Client::builder()
        .timeout(timeout)
        .build()
    {
        Ok(c) => c,
        Err(e) => return (HealthStatus::Unhealthy, Some(format!("Client build error: {}", e))),
    };

    match client.get(url).send().await {
        Ok(resp) => {
            let status_code = resp.status().as_u16();
            let expected = expected_status.unwrap_or(200);

            if status_code != expected {
                return (
                    HealthStatus::Unhealthy,
                    Some(format!(
                        "HTTP status mismatch: expected {}, got {}",
                        expected, status_code
                    )),
                );
            }

            if let Some(expected_body) = expected_body {
                match resp.text().await {
                    Ok(body) => {
                        if !body.contains(expected_body) {
                            return (
                                HealthStatus::Unhealthy,
                                Some("Body does not contain expected content".to_string()),
                            );
                        }
                    }
                    Err(e) => return (HealthStatus::Unhealthy, Some(format!("Body read error: {}", e))),
                }
            }

            (HealthStatus::Healthy, None)
        }
        Err(e) => (HealthStatus::Unhealthy, Some(format!("HTTP request failed: {}", e))),
    }
}

async fn execute_tcp(
    host: &str,
    port: u16,
    timeout_secs: Option<u64>,
) -> (HealthStatus, Option<String>) {
    let timeout = Duration::from_secs(timeout_secs.unwrap_or(5));
    let addr_str = format!("{}:{}", host, port);

    match addr_str.parse::<SocketAddr>() {
        Ok(addr) => match tokio::time::timeout(timeout, TcpStream::connect(addr)).await {
            Ok(Ok(_)) => (HealthStatus::Healthy, None),
            Ok(Err(e)) => (HealthStatus::Unhealthy, Some(format!("Connection failed: {}", e))),
            Err(_) => (HealthStatus::Unhealthy, Some("Connection timeout".to_string())),
        },
        Err(_) => match tokio::time::timeout(timeout, TcpStream::connect((host, port))).await {
            Ok(Ok(_)) => (HealthStatus::Healthy, None),
            Ok(Err(e)) => (HealthStatus::Unhealthy, Some(format!("Connection failed: {}", e))),
            Err(_) => (HealthStatus::Unhealthy, Some("Connection timeout".to_string())),
        },
    }
}

async fn execute_dns(domain: &str, timeout_secs: Option<u64>) -> (HealthStatus, Option<String>) {
    let timeout = Duration::from_secs(timeout_secs.unwrap_or(5));

    let resolver = match TokioAsyncResolver::tokio_from_system_conf() {
        Ok(r) => r,
        Err(e) => return (HealthStatus::Unhealthy, Some(format!("Resolver error: {}", e))),
    };

    match tokio::time::timeout(timeout, resolver.lookup_ip(domain)).await {
        Ok(Ok(response)) => {
            if response.iter().next().is_some() {
                (HealthStatus::Healthy, None)
            } else {
                (
                    HealthStatus::Unhealthy,
                    Some("No DNS records found".to_string()),
                )
            }
        }
        Ok(Err(e)) => (HealthStatus::Unhealthy, Some(format!("DNS lookup failed: {}", e))),
        Err(_) => (HealthStatus::Unhealthy, Some("DNS lookup timeout".to_string())),
    }
}

async fn execute_script(command: &str, timeout_secs: Option<u64>) -> (HealthStatus, Option<String>) {
    let timeout = Duration::from_secs(timeout_secs.unwrap_or(30));

    let child = if cfg!(target_os = "windows") {
        Command::new("cmd")
            .arg("/C")
            .arg(command)
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
    } else {
        Command::new("sh")
            .arg("-c")
            .arg(command)
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
    };

    let mut child = match child {
        Ok(c) => c,
        Err(e) => return (HealthStatus::Unhealthy, Some(format!("Failed to spawn: {}", e))),
    };

    match tokio::time::timeout(timeout, child.wait()).await {
        Ok(Ok(status)) => {
            if status.success() {
                (HealthStatus::Healthy, None)
            } else {
                (
                    HealthStatus::Unhealthy,
                    Some(format!("Exit code: {}", status.code().unwrap_or(-1))),
                )
            }
        }
        Ok(Err(e)) => (HealthStatus::Unhealthy, Some(format!("Wait error: {}", e))),
        Err(_) => {
            let _ = child.kill().await;
            (HealthStatus::Unhealthy, Some("Script timeout".to_string()))
        }
    }
}
