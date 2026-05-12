package com.poolmgr.scheduler;

import com.poolmgr.pool.PoolManager;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
public class PoolMaintenanceTask {

    private final PoolManager poolManager;

    public PoolMaintenanceTask(PoolManager poolManager) {
        this.poolManager = poolManager;
    }

    @Scheduled(fixedRate = 30000)
    public void maintainPools() {
        poolManager.maintainAllPools();
    }
}
