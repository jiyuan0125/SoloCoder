package com.example.proxy.loadbalance;

import com.example.proxy.model.Backend;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Random;

@Slf4j
@Component
public class WeightedRandomStrategy implements LoadBalanceStrategy {
    
    private final Random random = new Random();
    
    @Override
    public Backend select(List<Backend> availableBackends) {
        if (availableBackends == null || availableBackends.isEmpty()) {
            return null;
        }
        
        int totalWeight = availableBackends.stream()
                .mapToInt(Backend::getWeight)
                .sum();
        
        if (totalWeight <= 0) {
            return availableBackends.get(random.nextInt(availableBackends.size()));
        }
        
        int randomWeight = random.nextInt(totalWeight);
        int currentWeight = 0;
        
        for (Backend backend : availableBackends) {
            currentWeight += backend.getWeight();
            if (randomWeight < currentWeight) {
                log.debug("WeightedRandom selected backend: {}", backend.getUrl());
                return backend;
            }
        }
        
        Backend fallback = availableBackends.get(availableBackends.size() - 1);
        log.debug("WeightedRandom fallback selected backend: {}", fallback.getUrl());
        return fallback;
    }
    
    @Override
    public String getName() {
        return "weighted-random";
    }
}
