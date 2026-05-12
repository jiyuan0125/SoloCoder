package com.example.connectionpool.service;

import com.example.connectionpool.model.PooledConnection;
import com.example.connectionpool.pool.ConnectionPoolManager;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
@Slf4j
@RequiredArgsConstructor
public class ConnectionService {
    private final ConnectionPoolManager poolManager;
    private final Map<String, PooledConnection> inUseConnections = new ConcurrentHashMap<>();

    public String borrowConnection(String poolName) throws InterruptedException {
        PooledConnection conn = poolManager.borrowConnection(poolName);
        inUseConnections.put(conn.getId(), conn);
        return conn.getId();
    }

    public void releaseConnection(String connectionId) {
        PooledConnection conn = inUseConnections.remove(connectionId);
        if (conn != null) {
            poolManager.returnConnection(conn);
        }
    }
}
