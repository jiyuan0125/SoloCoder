package com.example.connectionpool.pool;

import com.example.connectionpool.config.ConnectionPoolConfig;
import com.example.connectionpool.exception.PoolFullException;
import com.example.connectionpool.exception.PoolTimeoutException;
import com.example.connectionpool.model.PoolMetrics;
import com.example.connectionpool.model.PoolStatus;
import com.example.connectionpool.model.PooledConnection;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;

import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

@Slf4j
public class HttpConnectionPool {
    private final String poolName;
    private final ReentrantLock lock = new ReentrantLock();
    private final Condition notEmpty = lock.newCondition();

    private final Queue<PooledConnection> idleConnections = new LinkedList<>();
    private final Set<PooledConnection> activeConnections = Collections.synchronizedSet(new HashSet<>());
    private final BlockingQueue<Waiter> waitingQueue;

    private final PoolMetrics metrics = new PoolMetrics();
    private volatile ConnectionPoolConfig config;

    private final AtomicInteger createdCount = new AtomicInteger(0);

    public HttpConnectionPool(ConnectionPoolConfig config) {
        this.poolName = config.getPoolName();
        this.config = config;
        this.waitingQueue = new ArrayBlockingQueue<>(config.getMaxQueueSize());
    }

    public String getPoolName() {
        return poolName;
    }

    public ConnectionPoolConfig getConfig() {
        return config;
    }

    public PooledConnection borrowConnection() throws InterruptedException {
        return borrowConnection(config.getWaitTimeoutMs(), TimeUnit.MILLISECONDS);
    }

    public PooledConnection borrowConnection(long timeout, TimeUnit unit) throws InterruptedException {
        long timeoutMs = unit.toMillis(timeout);
        long startTime = System.currentTimeMillis();

        Waiter waiter = null;
        lock.lock();
        try {
            PooledConnection conn = tryBorrow();
            if (conn != null) {
                return conn;
            }

            if (createdCount.get() < config.getMaxConnections()) {
                conn = createNewConnection();
                activeConnections.add(conn);
                metrics.incrementBorrowCount();
                conn.markBorrowed(Thread.currentThread().getName(), getStackTrace());
                log.debug("Pool [{}]: Created and borrowed new connection: {}", poolName, conn.getId());
                return conn;
            }

            waiter = new Waiter();
            if (!waitingQueue.offer(waiter)) {
                metrics.incrementTimeoutRejectCount();
                log.warn("Pool [{}]: Waiting queue is full (size: {}). Rejecting request.",
                        poolName, waitingQueue.size());
                throw new PoolFullException(poolName, waitingQueue.size());
            }
        } finally {
            lock.unlock();
        }

        boolean success = false;
        try {
            long remaining = timeoutMs - (System.currentTimeMillis() - startTime);
            if (remaining <= 0) {
                metrics.incrementTimeoutRejectCount();
                throw new PoolTimeoutException(poolName);
            }

            PooledConnection conn = waiter.waitForConnection(remaining);
            if (conn == null) {
                metrics.incrementTimeoutRejectCount();
                log.warn("Pool [{}]: Timeout waiting for connection after {} ms", poolName, timeoutMs);
                throw new PoolTimeoutException(poolName);
            }

            success = true;
            return conn;
        } finally {
            if (!success) {
                lock.lock();
                try {
                    waitingQueue.remove(waiter);
                } finally {
                    lock.unlock();
                }
            }
        }
    }

    public void returnConnection(PooledConnection connection) {
        if (connection == null) {
            return;
        }
        if (!connection.getPoolName().equals(this.poolName)) {
            log.warn("Trying to return connection from different pool. Connection pool: {}, Current pool: {}",
                    connection.getPoolName(), this.poolName);
            return;
        }

        lock.lock();
        try {
            if (!activeConnections.remove(connection)) {
                if (connection.isLeaked()) {
                    log.info("Pool [{}]: Leaked connection {} returned", poolName, connection.getId());
                } else {
                    log.warn("Pool [{}]: Returning connection that was not borrowed: {}", poolName, connection.getId());
                    return;
                }
            }

            connection.markReturned();

            int targetMax = config.getMaxConnections();
            int currentTotal = activeConnections.size() + idleConnections.size() + 1;

            if (currentTotal > targetMax) {
                log.info("Pool [{}]: Shrinking - destroying excess connection: {}", poolName, connection.getId());
                destroyConnection(connection);
                createdCount.decrementAndGet();
            } else {
                idleConnections.offer(connection);

                Waiter nextWaiter = waitingQueue.poll();
                if (nextWaiter != null) {
                    PooledConnection conn = idleConnections.poll();
                    if (conn != null) {
                        nextWaiter.deliver(conn);
                    } else {
                        idleConnections.offer(connection);
                    }
                }
            }
        } finally {
            lock.unlock();
        }
    }

    private PooledConnection tryBorrow() {
        PooledConnection conn = idleConnections.poll();
        if (conn != null) {
            activeConnections.add(conn);
            metrics.incrementBorrowCount();
            conn.markBorrowed(Thread.currentThread().getName(), getStackTrace());
            log.debug("Pool [{}]: Borrowed idle connection: {}", poolName, conn.getId());
            return conn;
        }
        return null;
    }

    private PooledConnection createNewConnection() {
        PooledConnection conn = new PooledConnection(poolName);
        createdCount.incrementAndGet();
        log.debug("Pool [{}]: Created new connection: {}", poolName, conn.getId());
        return conn;
    }

    private void destroyConnection(PooledConnection connection) {
        log.debug("Pool [{}]: Destroying connection: {}", poolName, connection.getId());
    }

    public void updateMaxConnections(int newMaxConnections) {
        lock.lock();
        try {
            int oldMax = config.getMaxConnections();
            config.setMaxConnections(newMaxConnections);

            if (newMaxConnections < oldMax) {
                int currentTotal = activeConnections.size() + idleConnections.size();
                int excess = Math.max(0, currentTotal - newMaxConnections);
                int destroyed = 0;

                while (destroyed < excess && !idleConnections.isEmpty()) {
                    PooledConnection conn = idleConnections.poll();
                    if (conn != null) {
                        destroyConnection(conn);
                        createdCount.decrementAndGet();
                        destroyed++;
                    }
                }

                log.info("Pool [{}]: Max connections updated from {} to {}, destroyed {} idle connections",
                        poolName, oldMax, newMaxConnections, destroyed);
            } else {
                log.info("Pool [{}]: Max connections updated from {} to {}", poolName, oldMax, newMaxConnections);
            }
        } finally {
            lock.unlock();
        }
    }

    public void updateWaitTimeout(long timeoutMs) {
        config.setWaitTimeoutMs(timeoutMs);
        log.info("Pool [{}]: Wait timeout updated to {} ms", poolName, timeoutMs);
    }

    public void updateLeakThresholdSeconds(int seconds) {
        config.setLeakThresholdSeconds(seconds);
        log.info("Pool [{}]: Leak threshold updated to {} seconds", poolName, seconds);
    }

    public void updateMaxQueueSize(int newSize) {
        lock.lock();
        try {
            int oldSize = config.getMaxQueueSize();
            config.setMaxQueueSize(newSize);
            log.info("Pool [{}]: Max queue size updated from {} to {}", poolName, oldSize, newSize);
        } finally {
            lock.unlock();
        }
    }

    public List<PooledConnection> checkLeakedConnections() {
        List<PooledConnection> leaked = new ArrayList<>();
        int threshold = config.getLeakThresholdSeconds();

        lock.lock();
        try {
            for (PooledConnection conn : new ArrayList<>(activeConnections)) {
                if (!conn.isLeaked() && conn.getBorrowedDurationSeconds() > threshold) {
                    conn.setLeaked(true);
                    leaked.add(conn);
                    metrics.incrementLeakDetectCount();
                    log.warn("Pool [{}]: LEAK DETECTED - Connection: {}, Borrowed: {}s ago, Thread: {},\nStackTrace:\n{}",
                            poolName,
                            conn.getId(),
                            conn.getBorrowedDurationSeconds(),
                            conn.getBorrowerThreadName(),
                            conn.getBorrowerStackTrace());
                }
            }
        } finally {
            lock.unlock();
        }

        return leaked;
    }

    public PoolStatus getStatus() {
        lock.lock();
        try {
            return PoolStatus.builder()
                    .poolName(poolName)
                    .activeCount(activeConnections.size())
                    .idleCount(idleConnections.size())
                    .waitingQueueSize(waitingQueue.size())
                    .totalBorrowCount(metrics.getTotalBorrowCount())
                    .totalTimeoutRejectCount(metrics.getTotalTimeoutRejectCount())
                    .totalLeakDetectCount(metrics.getTotalLeakDetectCount())
                    .maxConnections(config.getMaxConnections())
                    .waitTimeoutMs(config.getWaitTimeoutMs())
                    .leakThresholdSeconds(config.getLeakThresholdSeconds())
                    .maxQueueSize(config.getMaxQueueSize())
                    .build();
        } finally {
            lock.unlock();
        }
    }

    private String getStackTrace() {
        StringBuilder sb = new StringBuilder();
        StackTraceElement[] stackTrace = Thread.currentThread().getStackTrace();
        boolean foundPool = false;
        int count = 0;

        for (int i = 3; i < stackTrace.length && count < 20; i++) {
            StackTraceElement e = stackTrace[i];
            String className = e.getClassName();

            if (!foundPool) {
                if (className.startsWith("com.example.connectionpool")) {
                    continue;
                }
                foundPool = true;
            }

            sb.append("    at ").append(e).append("\n");
            count++;
        }

        return sb.toString();
    }

    private class Waiter {
        private final CountDownLatch latch = new CountDownLatch(1);
        private volatile PooledConnection connection;

        PooledConnection waitForConnection(long timeoutMs) throws InterruptedException {
            boolean awaited = latch.await(timeoutMs, TimeUnit.MILLISECONDS);
            return awaited ? connection : null;
        }

        void deliver(PooledConnection conn) {
            lock.lock();
            try {
                conn.markBorrowed(Thread.currentThread().getName(), getStackTrace());
                this.connection = conn;
                metrics.incrementBorrowCount();
                latch.countDown();
            } finally {
                lock.unlock();
            }
        }
    }
}
