package com.poolguard.pool;

import com.poolguard.model.PoolConfig;
import com.poolguard.model.PoolState;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.LocalDateTime;
import java.util.Map;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

public class ConnectionPool {
    private static final Logger logger = LoggerFactory.getLogger(ConnectionPool.class);
    
    private final String id;
    private final PoolConfig config;
    
    private volatile PoolState state;
    private final BlockingQueue<MockConnection> idleConnections;
    private final Map<String, MockConnection> activeConnections;
    private final ReentrantLock poolLock;
    private final Condition notEmpty;
    private final Condition notFull;
    private final Condition canClose;
    
    private final AtomicInteger saturationConsecutiveHealthChecks;
    private volatile LocalDateTime lastHealthCheckTime;
    
    private static final double SATURATION_THRESHOLD = 0.9;
    private static final int RECOVERY_HEALTH_CHECK_COUNT = 3;
    private static final long GET_CONNECTION_TIMEOUT = 10;
    
    public ConnectionPool(String id, PoolConfig config) {
        this.id = id;
        this.config = config;
        this.state = PoolState.RUNNING;
        this.idleConnections = new LinkedBlockingQueue<>(config.getMaxSize());
        this.activeConnections = new ConcurrentHashMap<>();
        this.poolLock = new ReentrantLock();
        this.notEmpty = poolLock.newCondition();
        this.notFull = poolLock.newCondition();
        this.canClose = poolLock.newCondition();
        this.saturationConsecutiveHealthChecks = new AtomicInteger(0);
        
        initializeConnections();
    }
    
    private void initializeConnections() {
        for (int i = 0; i < config.getMinIdle(); i++) {
            idleConnections.offer(createConnection());
        }
    }
    
    private MockConnection createConnection() {
        return new MockConnection();
    }
    
    public MockConnection getConnection() throws InterruptedException, TimeoutException, IllegalStateException {
        poolLock.lock();
        try {
            if (state == PoolState.CLOSING || state == PoolState.CLOSED) {
                throw new IllegalStateException("Pool is closing or already closed");
            }
            
            if (state == PoolState.ABNORMAL) {
                throw new IllegalStateException("Pool is in abnormal state");
            }
            
            while (idleConnections.isEmpty() && getTotalConnections() >= config.getMaxSize()) {
                if (!notFull.await(GET_CONNECTION_TIMEOUT, TimeUnit.SECONDS)) {
                    transitionToAbnormal("Connection acquisition timeout");
                    throw new TimeoutException("Timeout waiting for available connection");
                }
            }
            
            MockConnection conn = idleConnections.poll();
            if (conn == null && getTotalConnections() < config.getMaxSize()) {
                conn = createConnection();
            }
            
            if (conn == null) {
                transitionToAbnormal("Failed to create or acquire connection");
                throw new IllegalStateException("Failed to acquire connection");
            }
            
            activeConnections.put(conn.getId(), conn);
            notEmpty.signal();
            
            checkSaturation();
            
            return conn;
        } finally {
            poolLock.unlock();
        }
    }
    
    public void returnConnection(MockConnection conn) {
        if (conn == null) {
            return;
        }
        
        poolLock.lock();
        try {
            if (state == PoolState.CLOSING || state == PoolState.CLOSED) {
                conn.close();
                activeConnections.remove(conn.getId());
                if (state == PoolState.CLOSING && activeConnections.isEmpty()) {
                    state = PoolState.CLOSED;
                    canClose.signalAll();
                }
                return;
            }
            
            activeConnections.remove(conn.getId());
            
            if (conn.isValid()) {
                if (idleConnections.offer(conn)) {
                    notFull.signal();
                } else {
                    conn.close();
                }
            } else {
                conn.close();
            }
            
            checkRecoveryFromSaturation();
            
            if (state == PoolState.CLOSING && activeConnections.isEmpty()) {
                state = PoolState.CLOSED;
                canClose.signalAll();
            }
        } finally {
            poolLock.unlock();
        }
    }
    
    private void checkSaturation() {
        double usage = (double) getTotalConnections() / config.getMaxSize();
        if (state == PoolState.RUNNING && usage >= SATURATION_THRESHOLD) {
            transitionToSaturated();
        }
    }
    
    private void checkRecoveryFromSaturation() {
        double usage = (double) getTotalConnections() / config.getMaxSize();
        if (state == PoolState.SATURATED && usage < SATURATION_THRESHOLD) {
            transitionToRunning();
        }
    }
    
    public void performHealthCheck() {
        poolLock.lock();
        try {
            if (state == PoolState.CLOSING || state == PoolState.CLOSED) {
                return;
            }
            
            lastHealthCheckTime = LocalDateTime.now();
            
            boolean allConnectionsDead = true;
            int totalChecked = 0;
            int validCount = 0;
            
            for (MockConnection conn : idleConnections) {
                totalChecked++;
                if (conn.isValid()) {
                    validCount++;
                    allConnectionsDead = false;
                }
            }
            
            for (MockConnection conn : activeConnections.values()) {
                totalChecked++;
                if (conn.isValid()) {
                    validCount++;
                    allConnectionsDead = false;
                }
            }
            
            if (totalChecked == 0) {
                allConnectionsDead = false;
            }
            
            if (allConnectionsDead && totalChecked > 0) {
                transitionToAbnormal("All connections failed health check");
                return;
            }
            
            if (state == PoolState.SATURATED) {
                int currentChecks = saturationConsecutiveHealthChecks.incrementAndGet();
                if (currentChecks >= RECOVERY_HEALTH_CHECK_COUNT) {
                    transitionToRunning();
                    saturationConsecutiveHealthChecks.set(0);
                }
            }
            
            cleanupInvalidConnections();
            maintainMinimumIdleConnections();
            
        } finally {
            poolLock.unlock();
        }
    }
    
    private void cleanupInvalidConnections() {
        idleConnections.removeIf(conn -> {
            if (!conn.isValid()) {
                conn.close();
                return true;
            }
            return false;
        });
    }
    
    private void maintainMinimumIdleConnections() {
        while (idleConnections.size() < config.getMinIdle() && 
               getTotalConnections() < config.getMaxSize()) {
            idleConnections.offer(createConnection());
        }
    }
    
    private void transitionToRunning() {
        if (state == PoolState.SATURATED) {
            state = PoolState.RUNNING;
            saturationConsecutiveHealthChecks.set(0);
            logger.info("Pool [{}] transitioned from SATURATED to RUNNING at {}", 
                       config.getName(), LocalDateTime.now());
        }
    }
    
    private void transitionToSaturated() {
        if (state == PoolState.RUNNING) {
            state = PoolState.SATURATED;
            logAlert(config.getName(), PoolState.SATURATED);
        }
    }
    
    private void transitionToAbnormal(String reason) {
        if (state != PoolState.ABNORMAL && state != PoolState.CLOSING && state != PoolState.CLOSED) {
            state = PoolState.ABNORMAL;
            logAlert(config.getName(), PoolState.ABNORMAL);
            logger.error("Pool [{}] transitioned to ABNORMAL: {}", config.getName(), reason);
        }
    }
    
    private void logAlert(String poolName, PoolState poolState) {
        String alertMessage = String.format(
            "ALERT - Pool: %s, State: %s, Timestamp: %s",
            poolName,
            poolState,
            LocalDateTime.now()
        );
        logger.warn(alertMessage);
    }
    
    public void reset() {
        poolLock.lock();
        try {
            if (state != PoolState.ABNORMAL) {
                return;
            }
            
            for (MockConnection conn : activeConnections.values()) {
                conn.close();
            }
            activeConnections.clear();
            
            for (MockConnection conn : idleConnections) {
                conn.close();
            }
            idleConnections.clear();
            
            initializeConnections();
            state = PoolState.RUNNING;
            saturationConsecutiveHealthChecks.set(0);
            
            logger.info("Pool [{}] reset successfully, now in RUNNING state", config.getName());
        } finally {
            poolLock.unlock();
        }
    }
    
    public void close() throws InterruptedException {
        poolLock.lock();
        try {
            if (state == PoolState.CLOSING || state == PoolState.CLOSED) {
                return;
            }
            
            state = PoolState.CLOSING;
            
            while (!activeConnections.isEmpty()) {
                canClose.await();
            }
            
            for (MockConnection conn : idleConnections) {
                conn.close();
            }
            idleConnections.clear();
            
            state = PoolState.CLOSED;
            logger.info("Pool [{}] closed successfully", config.getName());
        } finally {
            poolLock.unlock();
        }
    }
    
    private int getTotalConnections() {
        return idleConnections.size() + activeConnections.size();
    }
    
    public String getId() {
        return id;
    }
    
    public PoolConfig getConfig() {
        return config;
    }
    
    public PoolState getState() {
        return state;
    }
    
    public int getActiveConnectionsCount() {
        return activeConnections.size();
    }
    
    public int getIdleConnectionsCount() {
        return idleConnections.size();
    }
    
    public LocalDateTime getLastHealthCheckTime() {
        return lastHealthCheckTime;
    }
    
    public void invalidateConnection(MockConnection conn) {
        if (conn != null) {
            conn.setValid(false);
        }
    }
}
