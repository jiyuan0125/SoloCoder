package com.poolguard.service;

import com.poolguard.model.PoolConfig;
import com.poolguard.model.PoolDetail;
import com.poolguard.model.PoolOverview;
import com.poolguard.pool.ConnectionPool;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class PoolManagerService {
    private static final Logger logger = LoggerFactory.getLogger(PoolManagerService.class);
    
    private final Map<String, ConnectionPool> pools;
    private final Map<String, String> nameToIdMap;
    
    public PoolManagerService() {
        this.pools = new ConcurrentHashMap<>();
        this.nameToIdMap = new ConcurrentHashMap<>();
    }
    
    public String registerPool(PoolConfig config) {
        if (nameToIdMap.containsKey(config.getName())) {
            throw new IllegalArgumentException("Pool with name '" + config.getName() + "' already exists");
        }
        
        String id = UUID.randomUUID().toString();
        ConnectionPool pool = new ConnectionPool(id, config);
        pools.put(id, pool);
        nameToIdMap.put(config.getName(), id);
        
        logger.info("Registered new pool: id={}, name={}", id, config.getName());
        return id;
    }
    
    public ConnectionPool getPool(String id) {
        return pools.get(id);
    }
    
    public List<PoolOverview> getAllPoolOverviews() {
        List<PoolOverview> overviews = new ArrayList<>();
        
        for (ConnectionPool pool : pools.values()) {
            overviews.add(new PoolOverview(
                pool.getId(),
                pool.getConfig().getName(),
                pool.getState(),
                pool.getActiveConnectionsCount(),
                pool.getIdleConnectionsCount()
            ));
        }
        
        return overviews;
    }
    
    public PoolDetail getPoolDetail(String id) {
        ConnectionPool pool = pools.get(id);
        if (pool == null) {
            return null;
        }
        
        return new PoolDetail(
            pool.getId(),
            pool.getConfig().getName(),
            pool.getConfig().getMaxSize(),
            pool.getConfig().getMinIdle(),
            pool.getConfig().getHealthCheckInterval(),
            pool.getState(),
            pool.getActiveConnectionsCount(),
            pool.getIdleConnectionsCount(),
            pool.getLastHealthCheckTime()
        );
    }
    
    public boolean resetPool(String id) {
        ConnectionPool pool = pools.get(id);
        if (pool == null) {
            return false;
        }
        pool.reset();
        return true;
    }
    
    public boolean removePool(String id) throws InterruptedException {
        ConnectionPool pool = pools.get(id);
        if (pool == null) {
            return false;
        }
        pool.close();
        pools.remove(id);
        nameToIdMap.remove(pool.getConfig().getName());
        return true;
    }
    
    @Scheduled(fixedRate = 30000)
    public void performGlobalHealthCheck() {
        logger.debug("Performing global health check for all pools");
        
        for (ConnectionPool pool : pools.values()) {
            try {
                pool.performHealthCheck();
            } catch (Exception e) {
                logger.error("Error during health check for pool: {}", pool.getConfig().getName(), e);
            }
        }
    }
    
    public boolean poolExists(String id) {
        return pools.containsKey(id);
    }
}
