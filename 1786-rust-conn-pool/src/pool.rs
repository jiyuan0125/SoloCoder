use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::{Duration, Instant};

use dashmap::DashMap;
use parking_lot::Mutex;
use tokio::net::TcpStream;
use tokio::sync::{broadcast, mpsc, Semaphore};
use tokio::time::interval;
use tracing::{info, warn, error};
use uuid::Uuid;

#[derive(Clone, Debug)]
pub struct PoolConfig {
    pub max_connections: usize,
    pub idle_timeout: Duration,
    pub health_check_interval: Duration,
    pub graceful_shutdown_timeout: Duration,
    pub max_health_failures: u32,
}

impl Default for PoolConfig {
    fn default() -> Self {
        Self {
            max_connections: 10,
            idle_timeout: Duration::from_secs(60),
            health_check_interval: Duration::from_secs(30),
            graceful_shutdown_timeout: Duration::from_secs(10),
            max_health_failures: 2,
        }
    }
}

#[derive(Clone, Debug, Default)]
pub struct PoolStats {
    pub total_active: usize,
    pub total_idle: usize,
    pub total_borrowed: usize,
    pub total_borrow_count: u64,
    pub total_create_failures: u64,
    pub per_backend: HashMap<String, BackendStats>,
}

#[derive(Clone, Debug, Default)]
pub struct BackendStats {
    pub active: usize,
    pub idle: usize,
    pub borrowed: usize,
    pub borrow_count: u64,
    pub create_failures: u64,
}

struct Connection {
    id: Uuid,
    target: SocketAddr,
    stream: Option<TcpStream>,
    last_used: Instant,
    health_failures: u32,
    borrowed_at: Option<Instant>,
}

impl Connection {
    fn new(target: SocketAddr, stream: TcpStream) -> Self {
        Self {
            id: Uuid::new_v4(),
            target,
            stream: Some(stream),
            last_used: Instant::now(),
            health_failures: 0,
            borrowed_at: None,
        }
    }

    fn is_idle_expired(&self, timeout: Duration) -> bool {
        self.borrowed_at.is_none() && self.last_used.elapsed() > timeout
    }

    fn record_health_failure(&mut self) {
        self.health_failures += 1;
    }

    fn reset_health(&mut self) {
        self.health_failures = 0;
    }

    fn should_remove(&self, max_failures: u32) -> bool {
        self.health_failures >= max_failures
    }
}

struct BackendPool {
    target: SocketAddr,
    idle: Vec<Connection>,
    borrowed: HashMap<Uuid, Connection>,
    semaphore: Arc<Semaphore>,
    stats: BackendStatsInternal,
}

#[derive(Default)]
struct BackendStatsInternal {
    borrow_count: u64,
    create_failures: u64,
}

impl BackendPool {
    fn new(target: SocketAddr, max_connections: usize) -> Self {
        Self {
            target,
            idle: Vec::new(),
            borrowed: HashMap::new(),
            semaphore: Arc::new(Semaphore::new(max_connections)),
            stats: BackendStatsInternal::default(),
        }
    }

    fn update_max_connections(&mut self, new_max: usize) {
        let current_max = self.semaphore.available_permits();
        if new_max > current_max {
            self.semaphore.add_permits(new_max - current_max);
        } else if new_max < current_max {
            let diff = current_max - new_max;
            for _ in 0..diff {
                let _ = self.semaphore.forget_permits(1);
            }
        }
    }

    fn get_idle(&mut self) -> Option<Connection> {
        self.idle.pop()
    }

    fn add_idle(&mut self, mut conn: Connection) {
        conn.last_used = Instant::now();
        conn.borrowed_at = None;
        self.idle.push(conn);
    }

    fn add_borrowed(&mut self, mut conn: Connection) {
        conn.borrowed_at = Some(Instant::now());
        self.stats.borrow_count += 1;
        self.borrowed.insert(conn.id, conn);
    }

    fn remove_borrowed(&mut self, id: &Uuid) -> Option<Connection> {
        self.borrowed.remove(id)
    }

    fn record_create_failure(&mut self) {
        self.stats.create_failures += 1;
    }

    fn cleanup_expired(&mut self, timeout: Duration) -> usize {
        let before = self.idle.len();
        self.idle.retain(|c| !c.is_idle_expired(timeout));
        before - self.idle.len()
    }

    fn snapshot_stats(&self) -> BackendStats {
        BackendStats {
            active: self.idle.len() + self.borrowed.len(),
            idle: self.idle.len(),
            borrowed: self.borrowed.len(),
            borrow_count: self.stats.borrow_count,
            create_failures: self.stats.create_failures,
        }
    }
}

pub struct PoolInner {
    config: Mutex<PoolConfig>,
    backends: DashMap<SocketAddr, Arc<Mutex<BackendPool>>>,
    shutdown_tx: Mutex<Option<broadcast::Sender<()>>>,
}

#[derive(Clone)]
pub struct ConnectionPool {
    inner: Arc<PoolInner>,
}

pub struct PooledConnection {
    pool: ConnectionPool,
    target: SocketAddr,
    conn_id: Option<Uuid>,
}

impl Drop for PooledConnection {
    fn drop(&mut self) {
        if let Some(id) = self.conn_id.take() {
            let pool = self.pool.clone();
            let target = self.target;
            tokio::spawn(async move {
                pool.return_connection(target, id).await;
            });
        }
    }
}

impl ConnectionPool {
    pub fn new(config: PoolConfig) -> Self {
        let (shutdown_tx, _) = broadcast::channel(1);
        Self {
            inner: Arc::new(PoolInner {
                config: Mutex::new(config),
                backends: DashMap::new(),
                shutdown_tx: Mutex::new(Some(shutdown_tx)),
            }),
        }
    }

    fn get_or_create_backend(&self, target: SocketAddr) -> Arc<Mutex<BackendPool>> {
        self.inner
            .backends
            .entry(target)
            .or_insert_with(|| {
                let config = self.inner.config.lock();
                Arc::new(Mutex::new(BackendPool::new(target, config.max_connections)))
            })
            .clone()
    }

    pub async fn acquire(&self, target: SocketAddr) -> Result<PooledConnection, String> {
        let backend = self.get_or_create_backend(target);
        let config = self.inner.config.lock().clone();

        let permit = backend
            .lock()
            .semaphore
            .clone()
            .acquire_owned()
            .map_err(|e| format!("semaphore closed: {}", e))?;

        let idle_conn = {
            let mut be = backend.lock();
            be.get_idle()
        };

        let conn = if let Some(mut c) = idle_conn {
            info!("reusing idle connection to {} (id={})", target, c.id);
            c.reset_health();
            c
        } else {
            info!("creating new connection to {}", target);
            match TcpStream::connect(target).await {
                Ok(stream) => Connection::new(target, stream),
                Err(e) => {
                    {
                        let mut be = backend.lock();
                        be.record_create_failure();
                    }
                    drop(permit);
                    error!("failed to create connection to {}: {}", target, e);
                    return Err(format!("connection failed: {}", e));
                }
            }
        };

        let conn_id = conn.id;
        {
            let mut be = backend.lock();
            be.add_borrowed(conn);
        }

        drop(permit);

        Ok(PooledConnection {
            pool: self.clone(),
            target,
            conn_id: Some(conn_id),
        })
    }

    async fn return_connection(&self, target: SocketAddr, conn_id: Uuid) {
        let backend = self.get_or_create_backend(target);

        let conn = {
            let mut be = backend.lock();
            be.remove_borrowed(&conn_id)
        };

        if let Some(mut conn) = conn {
            {
                let mut be = backend.lock();
                be.add_idle(conn);
            }
            info!("returned connection to {} (id={})", target, conn_id);
        } else {
            warn!("connection not found in borrowed map: {} (id={})", target, conn_id);
        }
    }

    pub fn update_config(&self, new_config: PoolConfig) {
        let mut config = self.inner.config.lock();
        *config = new_config.clone();

        for entry in self.inner.backends.iter() {
            let backend = entry.value();
            let mut be = backend.lock();
            be.update_max_connections(new_config.max_connections);
        }

        info!("config updated: max_connections={}, idle_timeout={:?}",
              new_config.max_connections, new_config.idle_timeout);
    }

    pub fn get_config(&self) -> PoolConfig {
        self.inner.config.lock().clone()
    }

    pub fn get_stats(&self) -> PoolStats {
        let mut total_active = 0;
        let mut total_idle = 0;
        let mut total_borrowed = 0;
        let mut total_borrow_count = 0;
        let mut total_create_failures = 0;
        let mut per_backend = HashMap::new();

        for entry in self.inner.backends.iter() {
            let target = entry.key().to_string();
            let backend = entry.value();
            let be = backend.lock();
            let stats = be.snapshot_stats();

            total_active += stats.active;
            total_idle += stats.idle;
            total_borrowed += stats.borrowed;
            total_borrow_count += stats.borrow_count;
            total_create_failures += stats.create_failures;
            per_backend.insert(target, stats);
        }

        PoolStats {
            total_active,
            total_idle,
            total_borrowed,
            total_borrow_count,
            total_create_failures,
            per_backend,
        }
    }

    pub async fn start_background_tasks(&self) {
        let pool = self.clone();
        let mut shutdown_rx = {
            let tx = pool.inner.shutdown_tx.lock();
            tx.as_ref().unwrap().subscribe()
        };

        loop {
            let config = pool.get_config();
            tokio::select! {
                _ = tokio::time::sleep(config.idle_timeout / 2) => {
                    pool.cleanup_idle_connections();
                }
                _ = tokio::time::sleep(config.health_check_interval) => {
                    pool.run_health_checks().await;
                }
                _ = shutdown_rx.recv() => {
                    info!("background tasks shutting down");
                    break;
                }
            }
        }
    }

    fn cleanup_idle_connections(&self) {
        let config = self.inner.config.lock().clone();
        let mut total_cleaned = 0;

        for entry in self.inner.backends.iter() {
            let backend = entry.value();
            let mut be = backend.lock();
            let cleaned = be.cleanup_expired(config.idle_timeout);
            if cleaned > 0 {
                info!("cleaned {} idle connections for {}", cleaned, entry.key());
            }
            total_cleaned += cleaned;
        }

        if total_cleaned > 0 {
            info!("total idle connections cleaned: {}", total_cleaned);
        }
    }

    async fn run_health_checks(&self) {
        let config = self.inner.config.lock().clone();
        let targets: Vec<SocketAddr> = self.inner.backends.iter().map(|e| *e.key()).collect();

        for target in targets {
            if let Some(backend) = self.inner.backends.get(&target) {
                let backend = backend.value().clone();
                let to_check: Vec<Connection> = {
                    let be = backend.lock();
                    be.idle.clone()
                };

                for mut conn in to_check {
                    if conn.stream.is_some() {
                        let stream = conn.stream.as_mut().unwrap();
                        match stream.ready(tokio::io::Interest::WRITABLE).await {
                            Ok(_) => {
                                conn.reset_health();
                            }
                            Err(_) => {
                                conn.record_health_failure();
                                warn!("health check failed for {} (id={}, failures={})",
                                      target, conn.id, conn.health_failures);
                            }
                        }

                        if conn.should_remove(config.max_health_failures) {
                            warn!("removing unhealthy connection {} (id={}) after {} failures",
                                  target, conn.id, config.max_health_failures);
                            let mut be = backend.lock();
                            be.idle.retain(|c| c.id != conn.id);
                        } else {
                            let mut be = backend.lock();
                            if let Some(pos) = be.idle.iter().position(|c| c.id == conn.id) {
                                be.idle[pos] = conn;
                            }
                        }
                    }
                }
            }
        }
    }

    pub async fn shutdown(&self) {
        info!("initiating pool shutdown");

        if let Some(tx) = self.inner.shutdown_tx.lock().take() {
            let _ = tx.send(());
        }

        let config = self.inner.config.lock().clone();
        let timeout = config.graceful_shutdown_timeout;
        let start = Instant::now();

        loop {
            let all_borrowed: Vec<(SocketAddr, Vec<(Uuid, Instant)>)> = self
                .inner
                .backends
                .iter()
                .map(|entry| {
                    let target = *entry.key();
                    let be = entry.value().lock();
                    let borrowed = be
                        .borrowed
                        .values()
                        .map(|c| (c.id, c.borrowed_at.unwrap_or(start)))
                        .collect();
                    (target, borrowed)
                })
                .collect();

            let total_borrowed: usize = all_borrowed.iter().map(|(_, v)| v.len()).sum();

            if total_borrowed == 0 {
                info!("all connections returned, shutdown complete");
                break;
            }

            if start.elapsed() > timeout {
                warn!("graceful shutdown timeout ({}s exceeded), forcing close of {} connections",
                      timeout.as_secs(), total_borrowed);

                for (target, conns) in all_borrowed {
                    for (id, borrowed_at) in conns {
                        let duration = borrowed_at.elapsed();
                        warn!(
                            "connection not returned: target={}, id={}, borrowed_duration={:?}",
                            target, id, duration
                        );
                    }
                }
                break;
            }

            info!("waiting for {} borrowed connections to return...", total_borrowed);
            tokio::time::sleep(Duration::from_millis(100)).await;
        }

        info!("pool shutdown finished");
    }
}
