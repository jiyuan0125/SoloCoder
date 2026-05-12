package com.example.pool.service;

import com.example.pool.core.ConnectionPool;
import com.example.pool.factory.ConnectionFactory;
import com.example.pool.model.*;
import jakarta.annotation.PreDestroy;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
public class PoolManagerService {

    private final Map<String, ConnectionPool> pools = new ConcurrentHashMap<>();
    private final Map<DataSourceType, ConnectionFactory> factories = new EnumMap<>(DataSourceType.class);

    @Autowired
    public PoolManagerService(List<ConnectionFactory> factoryList) {
        for (ConnectionFactory factory : factoryList) {
            factories.put(factory.getType(), factory);
        }
    }

    public void registerPool(PoolConfig config) throws Exception {
        if (pools.containsKey(config.getName())) {
            throw new IllegalArgumentException("Pool with name " + config.getName() + " already exists");
        }

        ConnectionFactory factory = factories.get(config.getType());
        if (factory == null) {
            throw new IllegalArgumentException("No factory found for type: " + config.getType());
        }

        ConnectionPool pool = new ConnectionPool(config, factory);
        pool.initialize();
        pools.put(config.getName(), pool);
        log.info("Pool {} registered and initialized", config.getName());
    }

    public Object acquireConnection(String name) throws Exception {
        ConnectionPool pool = pools.get(name);
        if (pool == null) {
            throw new IllegalArgumentException("Pool not found: " + name);
        }
        return pool.acquire();
    }

    public void releaseConnection(String name, Object connection) throws Exception {
        ConnectionPool pool = pools.get(name);
        if (pool == null) {
            throw new IllegalArgumentException("Pool not found: " + name);
        }
        pool.release(connection);
    }

    public List<PoolInfo> listPools() {
        List<PoolInfo> result = new ArrayList<>();
        for (Map.Entry<String, ConnectionPool> entry : pools.entrySet()) {
            result.add(toPoolInfo(entry.getKey(), entry.getValue()));
        }
        return result;
    }

    public PoolDetail getPoolDetail(String name) {
        ConnectionPool pool = pools.get(name);
        if (pool == null) {
            throw new IllegalArgumentException("Pool not found: " + name);
        }
        
        PoolDetail detail = new PoolDetail();
        fillPoolInfo(detail, name, pool);
        detail.setConnectionParams(pool.getConfig().getConnectionParams());
        detail.setAcquireTimeoutSeconds(pool.getConfig().getAcquireTimeoutSeconds());
        detail.setIdleTimeoutSeconds(pool.getConfig().getIdleTimeoutSeconds());
        return detail;
    }

    public void deletePool(String name) throws Exception {
        ConnectionPool pool = pools.remove(name);
        if (pool != null) {
            pool.close();
            log.info("Pool {} deleted", name);
        }
    }

    @Scheduled(fixedRateString = "${pool.idle-check-interval-seconds:30}000")
    public void maintainAllPools() {
        for (ConnectionPool pool : pools.values()) {
            try {
                pool.maintainIdleConnections();
            } catch (Exception e) {
                log.warn("Error maintaining pool: {}", e.getMessage());
            }
        }
    }

    @PreDestroy
    public void shutdown() throws Exception {
        log.info("Shutting down all pools...");
        for (String name : new ArrayList<>(pools.keySet())) {
            try {
                deletePool(name);
            } catch (Exception e) {
                log.warn("Error closing pool {}: {}", name, e.getMessage());
            }
        }
    }

    private PoolInfo toPoolInfo(String name, ConnectionPool pool) {
        PoolInfo info = new PoolInfo();
        fillPoolInfo(info, name, pool);
        return info;
    }

    private void fillPoolInfo(PoolInfo info, String name, ConnectionPool pool) {
        info.setName(name);
        info.setType(pool.getConfig().getType());
        info.setStatus(pool.getStatus());
        info.setMaxConnections(pool.getConfig().getMaxConnections());
        info.setMinIdle(pool.getConfig().getMinIdle());
        info.setActiveConnections(pool.getActiveCount());
        info.setIdleConnections(pool.getIdleCount());
        info.setWaitingRequests(pool.getWaitingCount());
    }
}
