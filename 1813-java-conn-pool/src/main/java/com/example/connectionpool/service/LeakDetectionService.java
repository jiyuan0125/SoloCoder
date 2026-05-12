package com.example.connectionpool.service;

import com.example.connectionpool.pool.ConnectionPoolManager;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class LeakDetectionService {
    private final ConnectionPoolManager poolManager;

    @Scheduled(fixedRate = 5000)
    public void checkLeakedConnections() {
        poolManager.checkLeakedConnections();
    }
}
