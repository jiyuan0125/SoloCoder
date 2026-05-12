package com.example.pool.core;

import com.example.pool.factory.ConnectionFactory;
import com.example.pool.model.PoolConfig;
import com.example.pool.model.PoolStatus;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;

import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.ReentrantLock;

@Slf4j
public class ConnectionPool {
    
    @Getter
    private final PoolConfig config;
    
    private final ConnectionFactory connectionFactory;
    private final Map<String, String> connectionParams;
    
    private final Semaphore semaphore;
    private final BlockingQueue<PooledConnection> idleConnections;
    private final AtomicInteger activeCount = new AtomicInteger(0);
    private final AtomicInteger waitingCount = new AtomicInteger(0);
    
    private volatile PoolStatus status = PoolStatus.INITIALIZING;
    private final ReentrantLock statusLock = new ReentrantLock();
    
    public ConnectionPool(PoolConfig config, ConnectionFactory connectionFactory) {
        this.config = config;
        this.connectionFactory = connectionFactory;
        this.connectionParams = config.getConnectionParams();
        this.semaphore = new Semaphore(config.getMaxConnections());
        this.idleConnections = new LinkedBlockingQueue<>(config.getMaxConnections());
    }
    
    public void initialize() throws Exception {
        statusLock.lock();
        try {
            if (status != PoolStatus.INITIALIZING) {
                throw new IllegalStateException("Pool is not in INITIALIZING state");
            }
            
            int minIdle = config.getMinIdle();
            for (int i = 0; i < minIdle; i++) {
                PooledConnection conn = createPooledConnection();
                idleConnections.offer(conn);
            }
            
            status = PoolStatus.RUNNING;
            log.info("Pool {} initialized with {} idle connections", config.getName(), minIdle);
        } finally {
            statusLock.unlock();
        }
    }
    
    public Object acquire() throws Exception {
        if (status == PoolStatus.CLOSED || status == PoolStatus.CLOSING) {
            throw new IllegalStateException("Pool is closed or closing");
        }
        
        waitingCount.incrementAndGet();
        try {
            boolean acquired = semaphore.tryAcquire(
                config.getAcquireTimeoutSeconds(), 
                TimeUnit.SECONDS
            );
            
            if (!acquired) {
                throw new Exception("Acquire connection timeout");
            }
            
            try {
                if (status == PoolStatus.PAUSED) {
                    semaphore.release();
                    throw new IllegalStateException("Pool is paused");
                }
                
                PooledConnection pooledConn = idleConnections.poll();
                if (pooledConn == null) {
                    pooledConn = createPooledConnection();
                } else if (!connectionFactory.validateConnection(pooledConn.getConnection())) {
                    destroyConnection(pooledConn);
                    pooledConn = createPooledConnection();
                }
                
                activeCount.incrementAndGet();
                return pooledConn.getConnection();
            } catch (Exception e) {
                semaphore.release();
                throw e;
            }
        } finally {
            waitingCount.decrementAndGet();
        }
    }
    
    public void release(Object connection) throws Exception {
        if (connection == null) {
            return;
        }
        
        PooledConnection pooledConn = null;
        boolean isValid = false;
        
        try {
            isValid = connectionFactory.validateConnection(connection);
        } catch (Exception e) {
            log.warn("Error validating connection for pool {}: {}", config.getName(), e.getMessage());
        }
        
        if (isValid) {
            pooledConn = new PooledConnection(connection);
            pooledConn.markReturned();
            if (!idleConnections.offer(pooledConn)) {
                pooledConn = null;
            }
        }
        
        if (pooledConn == null) {
            try {
                connectionFactory.closeConnection(connection);
            } catch (Exception e) {
                log.warn("Error closing invalid connection for pool {}: {}", config.getName(), e.getMessage());
            }
        }
        
        activeCount.decrementAndGet();
        semaphore.release();
    }
    
    public void pause() {
        statusLock.lock();
        try {
            if (status == PoolStatus.RUNNING) {
                status = PoolStatus.PAUSED;
                log.info("Pool {} paused", config.getName());
            }
        } finally {
            statusLock.unlock();
        }
    }
    
    public void resume() {
        statusLock.lock();
        try {
            if (status == PoolStatus.PAUSED) {
                status = PoolStatus.RUNNING;
                log.info("Pool {} resumed", config.getName());
            }
        } finally {
            statusLock.unlock();
        }
    }
    
    public void close() throws Exception {
        statusLock.lock();
        try {
            if (status == PoolStatus.CLOSED || status == PoolStatus.CLOSING) {
                return;
            }
            status = PoolStatus.CLOSING;
        } finally {
            statusLock.unlock();
        }
        
        while (activeCount.get() > 0) {
            Thread.sleep(100);
        }
        
        PooledConnection conn;
        while ((conn = idleConnections.poll()) != null) {
            destroyConnection(conn);
        }
        
        status = PoolStatus.CLOSED;
        log.info("Pool {} closed", config.getName());
    }
    
    public void maintainIdleConnections() {
        if (status != PoolStatus.RUNNING && status != PoolStatus.PAUSED) {
            return;
        }
        
        long idleTimeoutMs = (long) config.getIdleTimeoutSeconds() * 1000;
        long now = System.currentTimeMillis();
        
        int removed = 0;
        PooledConnection conn;
        while ((conn = idleConnections.peek()) != null) {
            long idleTime = now - conn.getLastReturnedAt().toEpochMilli();
            if (idleTime > idleTimeoutMs && idleConnections.size() > config.getMinIdle()) {
                idleConnections.poll();
                destroyConnection(conn);
                removed++;
            } else {
                break;
            }
        }
        
        while (idleConnections.size() < config.getMinIdle() && semaphore.availablePermits() > 0) {
            try {
                PooledConnection newConn = createPooledConnection();
                if (!idleConnections.offer(newConn)) {
                    destroyConnection(newConn);
                    break;
                }
            } catch (Exception e) {
                log.warn("Failed to create new connection for pool {} during maintenance: {}", 
                    config.getName(), e.getMessage());
                break;
            }
        }
        
        if (removed > 0) {
            log.debug("Pool {} removed {} idle connections during maintenance", config.getName(), removed);
        }
    }
    
    public PoolStatus getStatus() {
        return status;
    }
    
    public int getActiveCount() {
        return activeCount.get();
    }
    
    public int getIdleCount() {
        return idleConnections.size();
    }
    
    public int getWaitingCount() {
        return waitingCount.get();
    }
    
    private PooledConnection createPooledConnection() throws Exception {
        Object conn = connectionFactory.createConnection(connectionParams);
        return new PooledConnection(conn);
    }
    
    private void destroyConnection(PooledConnection pooledConn) {
        try {
            connectionFactory.closeConnection(pooledConn.getConnection());
        } catch (Exception e) {
            log.warn("Error closing connection for pool {}: {}", config.getName(), e.getMessage());
        }
    }
}
