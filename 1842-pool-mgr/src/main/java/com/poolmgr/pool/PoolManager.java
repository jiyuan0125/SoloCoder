package com.poolmgr.pool;

import com.poolmgr.model.AcquiredConnection;
import com.poolmgr.model.PoolConfig;
import com.poolmgr.model.PoolDetail;
import com.poolmgr.model.PoolStatus;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

@Component
public class PoolManager {

    private final Map<String, ManagedPool> pools = new ConcurrentHashMap<>();

    public boolean registerPool(PoolConfig config) {
        if (pools.containsKey(config.getName())) {
            return false;
        }
        
        ManagedPool pool = new ManagedPool(config);
        try {
            pool.initialize();
        } catch (Exception e) {
            return false;
        }
        
        pools.put(config.getName(), pool);
        return true;
    }

    public AcquiredConnection acquireConnection(String name) throws PoolTimeoutException, PoolDeletedException, PoolNotFoundException {
        ManagedPool pool = pools.get(name);
        if (pool == null) {
            throw new PoolNotFoundException("Pool not found: " + name);
        }
        return pool.acquire();
    }

    public void releaseConnection(String name, String connectionId) throws PoolNotFoundException {
        ManagedPool pool = pools.get(name);
        if (pool == null) {
            throw new PoolNotFoundException("Pool not found: " + name);
        }
        pool.release(connectionId);
    }

    public List<PoolStatus> getAllStatuses() {
        List<PoolStatus> statuses = new ArrayList<>();
        for (Map.Entry<String, ManagedPool> entry : pools.entrySet()) {
            statuses.add(entry.getValue().getStatus());
        }
        return statuses;
    }

    public PoolDetail getPoolDetail(String name) throws PoolNotFoundException {
        ManagedPool pool = pools.get(name);
        if (pool == null) {
            throw new PoolNotFoundException("Pool not found: " + name);
        }
        return pool.getDetail();
    }

    public boolean deletePool(String name) throws PoolNotFoundException, InterruptedException {
        ManagedPool pool = pools.remove(name);
        if (pool == null) {
            throw new PoolNotFoundException("Pool not found: " + name);
        }
        pool.shutdown();
        return true;
    }

    public void maintainAllPools() {
        for (ManagedPool pool : pools.values()) {
            pool.maintain();
        }
    }

    public static class PoolTimeoutException extends Exception {
        public PoolTimeoutException(String message) { super(message); }
    }

    public static class PoolDeletedException extends Exception {
        public PoolDeletedException(String message) { super(message); }
    }

    public static class PoolNotFoundException extends Exception {
        public PoolNotFoundException(String message) { super(message); }
    }

    private class ManagedPool {
        private final PoolConfig config;
        private final LinkedBlockingQueue<ConnectionWrapper> idleConnections;
        private final Map<String, ConnectionWrapper> activeConnections;
        private final AtomicInteger totalConnections;
        private final ReentrantLock lock;
        private final Condition poolChanged;
        private final AtomicBoolean deleted;
        private final AtomicInteger waitingCount;

        public ManagedPool(PoolConfig config) {
            this.config = config;
            this.idleConnections = new LinkedBlockingQueue<>();
            this.activeConnections = new ConcurrentHashMap<>();
            this.totalConnections = new AtomicInteger(0);
            this.lock = new ReentrantLock();
            this.poolChanged = lock.newCondition();
            this.deleted = new AtomicBoolean(false);
            this.waitingCount = new AtomicInteger(0);
        }

        public void initialize() throws Exception {
            int initialSize = Math.max(1, config.getMinIdle());
            for (int i = 0; i < initialSize; i++) {
                ConnectionWrapper conn = ConnectionFactory.createConnection(config);
                idleConnections.offer(conn);
                totalConnections.incrementAndGet();
            }
        }

        public AcquiredConnection acquire() throws PoolTimeoutException, PoolDeletedException {
            waitingCount.incrementAndGet();
            try {
                long timeoutMillis = config.getTimeoutSeconds() * 1000L;
                long startTime = System.currentTimeMillis();

                while (true) {
                    if (deleted.get()) {
                        throw new PoolDeletedException("Pool has been deleted: " + config.getName());
                    }

                    ConnectionWrapper conn = pollIdleConnection();
                    if (conn != null) {
                        activeConnections.put(conn.getId(), conn);
                        conn.setLastUsedTime(Instant.now());
                        return new AcquiredConnection(conn.getId(), config.getName(), conn.getType());
                    }

                    if (tryCreateNewConnection()) {
                        continue;
                    }

                    long elapsed = System.currentTimeMillis() - startTime;
                    long remaining = timeoutMillis - elapsed;
                    if (remaining <= 0) {
                        throw new PoolTimeoutException("Timeout waiting for connection: " + config.getName());
                    }

                    lock.lock();
                    try {
                        poolChanged.await(remaining, TimeUnit.MILLISECONDS);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        throw new PoolDeletedException("Interrupted while waiting for connection");
                    } finally {
                        lock.unlock();
                    }
                }
            } finally {
                waitingCount.decrementAndGet();
            }
        }

        private ConnectionWrapper pollIdleConnection() {
            ConnectionWrapper conn;
            while ((conn = idleConnections.poll()) != null) {
                if (conn.isValid()) {
                    return conn;
                } else {
                    conn.close();
                    totalConnections.decrementAndGet();
                }
            }
            return null;
        }

        private boolean tryCreateNewConnection() {
            lock.lock();
            try {
                if (totalConnections.get() < config.getMaxConnections()) {
                    try {
                        ConnectionWrapper conn = ConnectionFactory.createConnection(config);
                        idleConnections.offer(conn);
                        totalConnections.incrementAndGet();
                        return true;
                    } catch (Exception e) {
                        return false;
                    }
                }
                return false;
            } finally {
                lock.unlock();
            }
        }

        public void release(String connectionId) {
            ConnectionWrapper conn = activeConnections.remove(connectionId);
            if (conn == null) {
                return;
            }

            if (conn.isValid()) {
                conn.setLastUsedTime(Instant.now());
                idleConnections.offer(conn);
            } else {
                conn.close();
                totalConnections.decrementAndGet();
            }

            signalPoolChanged();
        }

        private void signalPoolChanged() {
            lock.lock();
            try {
                poolChanged.signalAll();
            } finally {
                lock.unlock();
            }
        }

        public PoolStatus getStatus() {
            return new PoolStatus(
                config.getName(),
                activeConnections.size(),
                idleConnections.size(),
                waitingCount.get()
            );
        }

        public PoolDetail getDetail() {
            return new PoolDetail(config, getStatus());
        }

        public void maintain() {
            if (deleted.get()) return;

            List<ConnectionWrapper> toRemove = new ArrayList<>();
            long idleTimeoutMillis = config.getTimeoutSeconds() * 1000L;
            Instant now = Instant.now();

            for (ConnectionWrapper conn : idleConnections) {
                long idleTime = now.toEpochMilli() - conn.getLastUsedTime().toEpochMilli();
                if (idleTime > idleTimeoutMillis || !conn.isValid()) {
                    toRemove.add(conn);
                }
            }

            for (ConnectionWrapper conn : toRemove) {
                if (idleConnections.remove(conn)) {
                    conn.close();
                    totalConnections.decrementAndGet();
                }
            }

            int currentIdle = idleConnections.size();
            int minIdle = config.getMinIdle();
            if (currentIdle < minIdle) {
                int toAdd = minIdle - currentIdle;
                for (int i = 0; i < toAdd; i++) {
                    lock.lock();
                    try {
                        if (totalConnections.get() >= config.getMaxConnections()) {
                            break;
                        }
                        try {
                            ConnectionWrapper conn = ConnectionFactory.createConnection(config);
                            idleConnections.offer(conn);
                            totalConnections.incrementAndGet();
                        } catch (Exception e) {
                            break;
                        }
                    } finally {
                        lock.unlock();
                    }
                }
            }

            signalPoolChanged();
        }

        public void shutdown() throws InterruptedException {
            deleted.set(true);
            signalPoolChanged();

            long timeoutMillis = config.getTimeoutSeconds() * 1000L;
            long startTime = System.currentTimeMillis();
            
            while (!activeConnections.isEmpty()) {
                long elapsed = System.currentTimeMillis() - startTime;
                if (elapsed > timeoutMillis) {
                    break;
                }
                Thread.sleep(100);
            }

            for (ConnectionWrapper conn : activeConnections.values()) {
                conn.close();
            }
            activeConnections.clear();

            ConnectionWrapper conn;
            while ((conn = idleConnections.poll()) != null) {
                conn.close();
            }
            totalConnections.set(0);
        }
    }
}
