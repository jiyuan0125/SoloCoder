package com.example.connectionpool.pool;

import com.example.connectionpool.config.ConnectionPoolConfig;
import com.example.connectionpool.config.PoolDefaultConfig;
import com.example.connectionpool.model.PoolStatus;
import com.example.connectionpool.model.PooledConnection;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Component
@Slf4j
@RequiredArgsConstructor
public class ConnectionPoolManager {
    private final Map<String, HttpConnectionPool> pools = new ConcurrentHashMap<>();
    private final PoolDefaultConfig defaultConfig;

    public HttpConnectionPool getOrCreatePool(String poolName) {
        return pools.computeIfAbsent(poolName, name -> {
            ConnectionPoolConfig config = defaultConfig.toConfig(name);
            log.info("Creating new connection pool: {}", name);
            return new HttpConnectionPool(config);
        });
    }

    public HttpConnectionPool getPool(String poolName) {
        return pools.get(poolName);
    }

    public HttpConnectionPool createPool(ConnectionPoolConfig config) {
        return pools.computeIfAbsent(config.getPoolName(), name -> {
            log.info("Creating connection pool with custom config: {}", name);
            return new HttpConnectionPool(config);
        });
    }

    public boolean poolExists(String poolName) {
        return pools.containsKey(poolName);
    }

    public Set<String> getAllPoolNames() {
        return new HashSet<>(pools.keySet());
    }

    public Map<String, PoolStatus> getAllPoolStatuses() {
        Map<String, PoolStatus> statuses = new LinkedHashMap<>();
        for (Map.Entry<String, HttpConnectionPool> entry : pools.entrySet()) {
            statuses.put(entry.getKey(), entry.getValue().getStatus());
        }
        return statuses;
    }

    public PooledConnection borrowConnection(String poolName) throws InterruptedException {
        HttpConnectionPool pool = getOrCreatePool(poolName);
        return pool.borrowConnection();
    }

    public void returnConnection(PooledConnection connection) {
        if (connection == null) {
            return;
        }
        HttpConnectionPool pool = pools.get(connection.getPoolName());
        if (pool != null) {
            pool.returnConnection(connection);
        }
    }

    public void checkLeakedConnections() {
        for (HttpConnectionPool pool : pools.values()) {
            pool.checkLeakedConnections();
        }
    }
}
