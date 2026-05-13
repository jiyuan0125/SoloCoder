package com.example.proxy.loadbalance;

import com.example.proxy.model.Backend;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

@Slf4j
@Component
public class RoundRobinStrategy implements LoadBalanceStrategy {
    
    private final AtomicInteger counter = new AtomicInteger(0);
    
    @Override
    public Backend select(List<Backend> availableBackends) {
        if (availableBackends == null || availableBackends.isEmpty()) {
            return null;
        }
        
        int index = Math.abs(counter.getAndIncrement() % availableBackends.size());
        Backend selected = availableBackends.get(index);
        log.debug("RoundRobin selected backend: {}", selected.getUrl());
        return selected;
    }
    
    @Override
    public String getName() {
        return "round-robin";
    }
}
