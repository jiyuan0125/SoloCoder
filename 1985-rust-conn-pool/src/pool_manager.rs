use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use dashmap::DashMap;
use tracing::{info, warn};

use crate::types::{PoolConfig, PoolStats};
use crate::connection::{ConnectionPool, PoolError, AcquiredConnection};

#[derive(Debug, Clone)]
pub struct PoolManager {
    pools: Arc<DashMap<String, ConnectionPool>>,
    shutdown_pools: Arc<RwLock<HashMap<String, ConnectionPool>>>,
}

impl PoolManager {
    pub fn new() -> Self {
        Self {
            pools: Arc::new(DashMap::new()),
            shutdown_pools: Arc::new(RwLock::new(HashMap::new())),
        }
    }
    
    pub fn register_pool(&self, name: String, config: PoolConfig) -> Result<(), PoolError> {
        if self.pools.contains_key(&name) {
            return Err(PoolError::ConnectionError(format!("Pool '{}' already exists", name)));
        }
        
        let pool = ConnectionPool::new(name.clone(), config);
        self.pools.insert(name.clone(), pool);
        
        info!("Pool registered: {}", name);
        Ok(())
    }
    
    pub async fn unregister_pool(&self, name: &str) -> Result<(), PoolError> {
        if let Some((_, pool)) = self.pools.remove(name) {
            pool.initiate_shutdown();
            self.shutdown_pools.write().await.insert(name.to_string(), pool);
            info!("Pool initiated shutdown: {}", name);
            Ok(())
        } else {
            Err(PoolError::PoolNotFound)
        }
    }
    
    pub fn get_pool(&self, name: &str) -> Option<ConnectionPool> {
        self.pools.get(name).map(|entry| entry.value().clone())
    }
    
    pub async fn acquire(&self, pool_name: &str) -> Result<AcquiredConnection, PoolError> {
        let pool = self.get_pool(pool_name)
            .ok_or(PoolError::PoolNotFound)?;
        
        pool.acquire().await
    }
    
    pub fn get_all_stats(&self) -> Vec<PoolStats> {
        let mut stats = Vec::new();
        
        for entry in self.pools.iter() {
            stats.push(entry.value().get_stats());
        }
        
        stats
    }
    
    pub fn get_pool_stats(&self, name: &str) -> Option<PoolStats> {
        self.pools.get(name).map(|entry| entry.value().get_stats())
    }
    
    pub async fn start_background_tasks(&self) {
        let manager_clone = self.clone();
        tokio::spawn(async move {
            let mut health_interval = tokio::time::interval(Duration::from_secs(30));
            loop {
                health_interval.tick().await;
                manager_clone.run_health_checks().await;
            }
        });
        
        let manager_clone = self.clone();
        tokio::spawn(async move {
            let mut maintenance_interval = tokio::time::interval(Duration::from_secs(60));
            loop {
                maintenance_interval.tick().await;
                manager_clone.run_maintenance().await;
            }
        });
        
        let manager_clone = self.clone();
        tokio::spawn(async move {
            let mut cleanup_interval = tokio::time::interval(Duration::from_secs(10));
            loop {
                cleanup_interval.tick().await;
                manager_clone.cleanup_shutdown_pools().await;
            }
        });
    }
    
    async fn run_health_checks(&self) {
        let pools: Vec<ConnectionPool> = self.pools.iter()
            .map(|entry| entry.value().clone())
            .collect();
        
        for pool in pools {
            tokio::spawn(async move {
                if let Err(e) = tokio::time::timeout(
                    Duration::from_secs(10),
                    pool.perform_health_check(),
                ).await {
                    warn!("Health check timeout for pool {}: {}", pool.name(), e);
                }
            });
        }
    }
    
    async fn run_maintenance(&self) {
        let pools: Vec<ConnectionPool> = self.pools.iter()
            .map(|entry| entry.value().clone())
            .collect();
        
        for pool in pools {
            pool.reap_idle();
            pool.maintain_idle();
        }
    }
    
    async fn cleanup_shutdown_pools(&self) {
        let mut shutdown = self.shutdown_pools.write().await;
        let mut to_remove = Vec::new();
        
        for (name, pool) in shutdown.iter() {
            if pool.check_and_close() {
                to_remove.push(name.clone());
                info!("Pool closed successfully: {}", name);
            }
        }
        
        for name in to_remove {
            shutdown.remove(&name);
        }
    }
}

impl Default for PoolManager {
    fn default() -> Self {
        Self::new()
    }
}
