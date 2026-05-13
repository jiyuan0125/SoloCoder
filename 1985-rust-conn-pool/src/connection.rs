use std::collections::VecDeque;
use std::time::{Duration, Instant};
use std::sync::Arc;
use tokio::sync::{oneshot, Notify, Semaphore, OwnedSemaphorePermit};
use tokio::net::TcpStream;
use parking_lot::Mutex;
use uuid::Uuid;

use crate::types::{PoolConfig, PoolState, PoolStats, ConnectionHealth};
use crate::rate_limiter::RateLimiter;

#[derive(Debug)]
pub struct ConnectionInfo {
    pub id: String,
    pub stream: Option<TcpStream>,
    pub created_at: Instant,
    pub last_used_at: Instant,
    pub health: ConnectionHealth,
    pub heartbeat_failures: u32,
}

impl ConnectionInfo {
    pub fn new(stream: TcpStream) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            stream: Some(stream),
            created_at: Instant::now(),
            last_used_at: Instant::now(),
            health: ConnectionHealth::Unknown,
            heartbeat_failures: 0,
        }
    }
}

#[derive(Debug)]
struct Waiter {
    tx: oneshot::Sender<Arc<Mutex<ConnectionInfo>>>,
    requested_at: Instant,
}

#[derive(Debug, Clone)]
pub struct ConnectionPool {
    name: String,
    config: PoolConfig,
    inner: Arc<PoolInner>,
}

#[derive(Debug)]
struct PoolInner {
    active: Mutex<Vec<Arc<Mutex<ConnectionInfo>>>>,
    idle: Mutex<VecDeque<Arc<Mutex<ConnectionInfo>>>>,
    waiters: Mutex<VecDeque<Waiter>>,
    state: Mutex<PoolState>,
    semaphore: Arc<Semaphore>,
    rate_limiter: RateLimiter,
    notify: Notify,
    stats: Mutex<PoolStatsInner>,
}

#[derive(Debug, Default)]
struct PoolStatsInner {
    total_created: u64,
    total_destroyed: u64,
}

#[derive(Debug, Clone)]
pub enum PoolError {
    PoolShuttingDown,
    PoolClosed,
    PoolNotFound,
    AcquireTimeout,
    ConnectionError(String),
    RateLimited,
}

impl std::fmt::Display for PoolError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            PoolError::PoolShuttingDown => write!(f, "Pool is shutting down"),
            PoolError::PoolClosed => write!(f, "Pool is closed"),
            PoolError::PoolNotFound => write!(f, "Pool not found"),
            PoolError::AcquireTimeout => write!(f, "Acquire timeout"),
            PoolError::ConnectionError(s) => write!(f, "Connection error: {}", s),
            PoolError::RateLimited => write!(f, "Rate limited"),
        }
    }
}

impl std::error::Error for PoolError {}

impl ConnectionPool {
    pub fn new(name: String, config: PoolConfig) -> Self {
        let semaphore = Arc::new(Semaphore::new(config.max_connections as usize));
        let rate_limiter = RateLimiter::new(config.max_create_per_second);
        
        let pool = Self {
            name: name.clone(),
            config: config.clone(),
            inner: Arc::new(PoolInner {
                active: Mutex::new(Vec::new()),
                idle: Mutex::new(VecDeque::new()),
                waiters: Mutex::new(VecDeque::new()),
                state: Mutex::new(PoolState::Active),
                semaphore,
                rate_limiter,
                notify: Notify::new(),
                stats: Mutex::new(PoolStatsInner::default()),
            }),
        };
        
        pool
    }
    
    pub fn name(&self) -> &str {
        &self.name
    }
    
    pub fn config(&self) -> &PoolConfig {
        &self.config
    }
    
    pub fn get_state(&self) -> PoolState {
        *self.inner.state.lock()
    }
    
    pub fn get_stats(&self) -> PoolStats {
        let active_count = self.inner.active.lock().len() as u32;
        let idle_count = self.inner.idle.lock().len() as u32;
        let waiting_count = self.inner.waiters.lock().len() as u32;
        let stats = self.inner.stats.lock();
        let state = *self.inner.state.lock();
        
        PoolStats {
            name: self.name.clone(),
            active_connections: active_count,
            idle_connections: idle_count,
            waiting_requests: waiting_count,
            total_created: stats.total_created,
            total_destroyed: stats.total_destroyed,
            state,
        }
    }
    
    pub async fn acquire(&self) -> Result<AcquiredConnection, PoolError> {
        if *self.inner.state.lock() != PoolState::Active {
            return Err(PoolError::PoolShuttingDown);
        }
        
        let acquire_timeout = Duration::from_millis(self.config.acquire_timeout_ms);
        let start = Instant::now();
        
        loop {
            if start.elapsed() >= acquire_timeout {
                return Err(PoolError::AcquireTimeout);
            }
            
            if let Some(conn) = self.try_acquire_idle() {
                return Ok(AcquiredConnection::new(conn, self.inner.clone()));
            }
            
            if self.can_create_new() {
                if self.inner.rate_limiter.try_acquire() {
                    let permit = self.inner.semaphore.clone().acquire_owned().await
                        .map_err(|e| PoolError::ConnectionError(e.to_string()))?;
                    
                    match self.create_connection().await {
                        Ok(conn) => {
                            let conn_arc = Arc::new(Mutex::new(conn));
                            self.inner.active.lock().push(conn_arc.clone());
                            self.inner.stats.lock().total_created += 1;
                            return Ok(AcquiredConnection::with_permit(conn_arc, self.inner.clone(), permit));
                        }
                        Err(e) => {
                            drop(permit);
                            return Err(e);
                        }
                    }
                }
            }
            
            let (tx, rx) = oneshot::channel();
            let waiter = Waiter {
                tx,
                requested_at: Instant::now(),
            };
            self.inner.waiters.lock().push_back(waiter);
            
            let remaining = acquire_timeout.saturating_sub(start.elapsed());
            let wait_fut = async {
                match tokio::time::timeout(remaining, rx).await {
                    Ok(Ok(conn)) => Some(conn),
                    _ => None,
                }
            };
            
            if let Some(conn) = wait_fut.await {
                return Ok(AcquiredConnection::new(conn, self.inner.clone()));
            }
        }
    }
    
    fn try_acquire_idle(&self) -> Option<Arc<Mutex<ConnectionInfo>>> {
        let mut idle = self.inner.idle.lock();
        while let Some(conn) = idle.pop_front() {
            let info = conn.lock();
            if info.health == ConnectionHealth::Unhealthy {
                drop(info);
                self.destroy_connection(&conn);
                continue;
            }
            
            let elapsed = info.created_at.elapsed();
            let max_lifetime = Duration::from_millis(self.config.max_lifetime_ms);
            if elapsed >= max_lifetime {
                drop(info);
                self.destroy_connection(&conn);
                continue;
            }
            
            drop(info);
            self.inner.active.lock().push(conn.clone());
            return Some(conn);
        }
        None
    }
    
    fn can_create_new(&self) -> bool {
        let active_count = self.inner.active.lock().len() as u32;
        let idle_count = self.inner.idle.lock().len() as u32;
        let total = active_count + idle_count;
        total < self.config.max_connections
    }
    
    async fn create_connection(&self) -> Result<ConnectionInfo, PoolError> {
        let addr = format!("{}:{}", self.config.address, self.config.port);
        let timeout = Duration::from_millis(self.config.connection_timeout_ms);
        
        let stream = tokio::time::timeout(
            timeout,
            TcpStream::connect(&addr),
        ).await
            .map_err(|_| PoolError::ConnectionError("Connection timeout".to_string()))?
            .map_err(|e| PoolError::ConnectionError(e.to_string()))?;
        
        Ok(ConnectionInfo::new(stream))
    }
    
    fn release_connection(&self, conn: &Arc<Mutex<ConnectionInfo>>) {
        {
            let mut active = self.inner.active.lock();
            if let Some(pos) = active.iter().position(|c| Arc::ptr_eq(c, conn)) {
                active.remove(pos);
            }
        }
        
        let should_destroy = {
            let info = conn.lock();
            info.health == ConnectionHealth::Unhealthy || 
            info.created_at.elapsed() >= Duration::from_millis(self.config.max_lifetime_ms)
        };
        
        if should_destroy {
            self.destroy_connection(conn);
        } else {
            {
                let mut info = conn.lock();
                info.last_used_at = Instant::now();
            }
            
            let mut waiter_served = false;
            {
                let mut waiters = self.inner.waiters.lock();
                while let Some(waiter) = waiters.pop_front() {
                    if waiter.tx.send(conn.clone()).is_ok() {
                        waiter_served = true;
                        break;
                    }
                }
            }
            
            if !waiter_served {
                self.inner.idle.lock().push_back(conn.clone());
            }
        }
        
        self.inner.notify.notify_one();
    }
    
    fn destroy_connection(&self, conn: &Arc<Mutex<ConnectionInfo>>) {
        let mut info = conn.lock();
        info.stream.take();
        drop(info);
        self.inner.stats.lock().total_destroyed += 1;
    }
    
    pub fn initiate_shutdown(&self) {
        *self.inner.state.lock() = PoolState::ShuttingDown;
        self.clear_idle_connections();
        self.reject_all_waiters();
    }
    
    pub fn check_and_close(&self) -> bool {
        let active_count = self.inner.active.lock().len();
        if active_count == 0 && *self.inner.state.lock() == PoolState::ShuttingDown {
            *self.inner.state.lock() = PoolState::Closed;
            true
        } else {
            false
        }
    }
    
    fn clear_idle_connections(&self) {
        let mut idle = self.inner.idle.lock();
        while let Some(conn) = idle.pop_front() {
            self.destroy_connection(&conn);
        }
    }
    
    fn reject_all_waiters(&self) {
        let mut waiters = self.inner.waiters.lock();
        while let Some(waiter) = waiters.pop_front() {
            drop(waiter.tx);
        }
    }
    
    pub fn maintain_idle(&self) {
        if *self.inner.state.lock() != PoolState::Active {
            return;
        }
        
        let idle_count = self.inner.idle.lock().len() as u32;
        
        if idle_count < self.config.min_idle {
            let needed = self.config.min_idle - idle_count;
            for _ in 0..needed {
                if !self.can_create_new() {
                    break;
                }
                if !self.inner.rate_limiter.try_acquire() {
                    break;
                }
                
                let inner_clone = self.inner.clone();
                let config_clone = self.config.clone();
                tokio::spawn(async move {
                    let addr = format!("{}:{}", config_clone.address, config_clone.port);
                    let timeout = Duration::from_millis(config_clone.connection_timeout_ms);
                    
                    if let Ok(Ok(stream)) = tokio::time::timeout(
                        timeout,
                        TcpStream::connect(&addr),
                    ).await {
                        let conn = ConnectionInfo::new(stream);
                        let conn_arc = Arc::new(Mutex::new(conn));
                        inner_clone.idle.lock().push_back(conn_arc);
                        inner_clone.stats.lock().total_created += 1;
                    }
                });
            }
        }
    }
    
    pub fn reap_idle(&self) {
        let now = Instant::now();
        let idle_timeout = Duration::from_millis(self.config.idle_timeout_ms);
        
        let mut idle = self.inner.idle.lock();
        let mut to_remove = Vec::new();
        
        for (idx, conn) in idle.iter().enumerate() {
            let info = conn.lock();
            if now.duration_since(info.last_used_at) >= idle_timeout {
                to_remove.push(idx);
            }
        }
        
        for idx in to_remove.iter().rev() {
            if let Some(conn) = idle.remove(*idx) {
                drop(idle);
                self.destroy_connection(&conn);
                idle = self.inner.idle.lock();
            }
        }
    }
    
    pub async fn perform_health_check(&self) {
        let idle_conns: Vec<Arc<Mutex<ConnectionInfo>>> = self.inner.idle.lock().iter().cloned().collect();
        
        for conn in idle_conns {
            let heartbeat_data = [0u8; 4];
            
            let stream_option = {
                let info = conn.lock();
                info.stream.is_some()
            };
            
            if !stream_option {
                conn.lock().health = ConnectionHealth::Unhealthy;
                continue;
            }
            
            let writable_result = {
                let info = conn.lock();
                if let Some(ref stream) = info.stream {
                    Some(stream.try_write(&heartbeat_data))
                } else {
                    None
                }
            };
            
            match writable_result {
                Some(Ok(_)) => {
                    let mut info = conn.lock();
                    info.heartbeat_failures = 0;
                    info.health = ConnectionHealth::Healthy;
                }
                Some(Err(e)) if e.kind() == std::io::ErrorKind::WouldBlock => {
                    let mut info = conn.lock();
                    info.heartbeat_failures = 0;
                    info.health = ConnectionHealth::Healthy;
                }
                Some(Err(_)) => {
                    let mut info = conn.lock();
                    info.heartbeat_failures += 1;
                    if info.heartbeat_failures >= self.config.heartbeat_failed_threshold {
                        info.health = ConnectionHealth::Unhealthy;
                    }
                }
                None => {
                    conn.lock().health = ConnectionHealth::Unhealthy;
                }
            }
        }
    }
}

pub struct AcquiredConnection {
    conn: Option<Arc<Mutex<ConnectionInfo>>>,
    pool_inner: Arc<PoolInner>,
    _permit: Option<OwnedSemaphorePermit>,
}

impl AcquiredConnection {
    fn new(conn: Arc<Mutex<ConnectionInfo>>, pool_inner: Arc<PoolInner>) -> Self {
        Self {
            conn: Some(conn),
            pool_inner,
            _permit: None,
        }
    }
    
    fn with_permit(
        conn: Arc<Mutex<ConnectionInfo>>, 
        pool_inner: Arc<PoolInner>,
        permit: OwnedSemaphorePermit,
    ) -> Self {
        Self {
            conn: Some(conn),
            pool_inner,
            _permit: Some(permit),
        }
    }
    
    pub fn id(&self) -> String {
        self.conn.as_ref().map(|c| c.lock().id.clone()).unwrap_or_default()
    }
}

impl Drop for AcquiredConnection {
    fn drop(&mut self) {
        if let Some(conn) = self.conn.take() {
            let active = {
                let mut active = self.pool_inner.active.lock();
                if let Some(pos) = active.iter().position(|c| Arc::ptr_eq(c, &conn)) {
                    active.remove(pos);
                    true
                } else {
                    false
                }
            };
            
            if active {
                let should_destroy = {
                    let info = conn.lock();
                    info.health == ConnectionHealth::Unhealthy || 
                    info.created_at.elapsed() >= Duration::from_millis(3600000)
                };
                
                if should_destroy {
                    let mut info = conn.lock();
                    info.stream.take();
                    drop(info);
                    self.pool_inner.stats.lock().total_destroyed += 1;
                } else {
                    {
                        let mut info = conn.lock();
                        info.last_used_at = Instant::now();
                    }
                    
                    let mut waiter_served = false;
                    {
                        let mut waiters = self.pool_inner.waiters.lock();
                        while let Some(waiter) = waiters.pop_front() {
                            if waiter.tx.send(conn.clone()).is_ok() {
                                waiter_served = true;
                                break;
                            }
                        }
                    }
                    
                    if !waiter_served {
                        self.pool_inner.idle.lock().push_back(conn.clone());
                    }
                }
                
                self.pool_inner.notify.notify_one();
            }
        }
    }
}
